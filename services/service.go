package services

import (
	"errors"
	"fmt"
	"github.com/coreos/bbolt"
	"github.com/fanliao/go-concurrentMap"
	"github.com/ldsec/drynx/lib"
	"github.com/ldsec/drynx/lib/proof"
	"github.com/ldsec/drynx/lib/provider"
	"github.com/ldsec/drynx/protocols"
	"github.com/ldsec/unlynx/lib"
	"github.com/ldsec/unlynx/lib/aggregation"
	"github.com/ldsec/unlynx/lib/differential_privacy"
	"github.com/ldsec/unlynx/lib/key_switch"
	"github.com/ldsec/unlynx/lib/shuffle"
	"github.com/ldsec/unlynx/lib/tools"
	"github.com/ldsec/unlynx/protocols"
	"go.dedis.ch/cothority/v3/skipchain"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/util/random"
	"go.dedis.ch/onet/v3"
	"go.dedis.ch/onet/v3/log"
	"go.dedis.ch/onet/v3/network"
	"sync"
	"time"
)

// ServiceName is the registered name for the drynx service.
const ServiceName = "drynx"

const gobFile = "pre_compute_multiplications.gob"

// Survey represents a survey with the corresponding params
type Survey struct {
	SurveyQuery        libdrynx.SurveyQuery
	QueryResponseState libdrynx.ResponseAllDPs // QueryResponse keeps track of the response from the data providers, the aggregated data, and the final results
	Noises             libunlynx.CipherVector
	ShufflePrecompute  []libunlynxshuffle.CipherVectorScalar
	MapPIs             map[string]onet.ProtocolInstance

	// mutex
	Mutex *sync.Mutex
}

func castToSurvey(object interface{}, err error) Survey {
	if err != nil {
		log.Fatalf("[SERVICE] <drynx> Server, Error reading map, %+v", err)
	}
	survey, ok := object.(Survey)
	if !ok {
		err := fmt.Errorf("unable to cast to Survey, is %#v", object)
		log.Fatalf("[SERVICE] <drynx> Server, %+v", err)
	}
	return survey
}

// DPqueryReceived is used to ensure that all DPs have received the query and can proceed with the data collection protocol
type DPqueryReceived struct {
	SurveyID string
}

// SyncDCP used to synchronize the computing nodes
type SyncDCP struct {
	SurveyID string
}

// DPdataFinished is used to ensure that all servers have received the data and can proceed with the collective aggregation
type DPdataFinished struct {
	SurveyID string
}

// ServiceDrynx defines a service in drynx with a survey.
type ServiceDrynx struct {
	*onet.ServiceProcessor

	// ---- Computing Nodes ----
	Survey *concurrent.ConcurrentMap
	// -------------------------

	// ---- Data Provider ----
	loader      provider.Loader
	neutralizer provider.Neutralizer
	// -------------------------

	// ---- Verifying Nodes ----
	Skipchain     *skipchain.Client
	LastSkipBlock *skipchain.SkipBlock
	//Contains size for a query, as well as the bitmap for this query
	Request *concurrent.ConcurrentMap
	//the name of DB and the DB in itself is dedicated to the server.
	DBPath string
	DB     *bbolt.DB
	//To make everything thread safe (database access and updating parameters)
	Mutex *sync.Mutex
	// Too receive the bitmaps from the nodes
	SharedBMChannel chan map[string]int64
	// To finish the protocols
	SharedBMChannelToTerminate chan struct{}
	// -------------------------
}

// MsgTypes defines the Message Type SurveyID for all the service's intra-messages.
type MsgTypes struct {
	msgSurveyQuery     network.MessageTypeID
	msgSurveyQueryToDP network.MessageTypeID
	msgDPqueryReceived network.MessageTypeID
	msgSyncDCP         network.MessageTypeID
	msgDPdataFinished  network.MessageTypeID
}

var msgTypes = MsgTypes{}

// Process implements the processor interface and is used to recognize messages broadcasted between servers
func (s *ServiceDrynx) Process(msg *network.Envelope) {
	if msg.MsgType.Equal(msgTypes.msgSurveyQuery) {
		tmp := (msg.Msg).(*libdrynx.SurveyQuery)
		_, err := s.HandleSurveyQuery(tmp)
		log.ErrFatal(err)
	} else if msg.MsgType.Equal(msgTypes.msgSurveyQueryToDP) {
		tmp := (msg.Msg).(*libdrynx.SurveyQueryToDP)
		_, err := s.HandleSurveyQueryToDP(tmp)
		log.ErrFatal(err)
	} else {
		log.Warnf("unprocessed message: %#v", msg)
	}
}

func (s *ServiceDrynx) waitForSurvey(id string) Survey {
	for {
		if obj, _ := s.Survey.Get(id); obj != nil {
			return obj.(Survey)
		}

		time.Sleep(time.Millisecond * 100)
	}
}

// Query Handlers
//______________________________________________________________________________________________________________________

// HandleSurveyQuery handles the reception of a survey creation query by instantiating the corresponding survey.
func (s *ServiceDrynx) HandleSurveyQuery(recq *libdrynx.SurveyQuery) (network.Message, error) {
	prefixWithID := func(args []interface{}) []interface{} {
		arr := make([]interface{}, len(args)+2)
		arr[0] = "[SERVICE] <drynx> Server"
		arr[1] = s.ServerIdentity().String()
		copy(arr[2:], args)
		return arr
	}
	info := func(args ...interface{}) {
		log.Info(prefixWithID(args)...)
	}
	die := func(args ...interface{}) {
		log.Fatal(prefixWithID(args)...)
	}

	info("received a [SurveyQuery]")

	// get the total number DPs
	nbrDPs := 0
	for _, v := range recq.ServerToDP {
		if v != nil {
			nbrDPs += len(v.Content)
		}
	}

	// only generate ProofCollection protocol instances if proofs is enabled
	var mapPIs map[string]onet.ProtocolInstance
	if recq.Query.Proofs != 0 {
		var err error
		mapPIs, err = s.generateMapPIs(recq)
		if err != nil {
			return nil, err
		}
	}

	// survey instantiation
	_, err := s.Survey.Put(recq.SurveyID, Survey{
		SurveyQuery: *recq,
		MapPIs:      mapPIs,
	})
	if err != nil {
		return nil, err
	}

	survey := castToSurvey(s.Survey.Get(recq.SurveyID))

	// prepares the precomputation for shuffling
	lineSize := 100 // + 1 is for the possible count attribute
	survey.ShufflePrecompute, _ = libunlynxshuffle.PrecomputationWritingForShuffling(false, gobFile, s.ServerIdentity().String(), libunlynx.SuiTe.Scalar().Pick(random.New()), recq.RosterServers.Aggregate, lineSize)

	// if is the root server: send query to all other servers and its data providers
	if recq.IntraMessage == false {
		info("broadcasting [SurveyQuery] to CNs ")
		recq.IntraMessage = true
		// to other computing servers
		err = libunlynxtools.SendISMOthers(s.ServiceProcessor, &recq.RosterServers, recq)
		if err != nil {
			die("broadcasting [SurveyQuery] to CNs error", err)
		}
		recq.IntraMessage = false
	}

	// to the DPs
	CNsToDPs := make(map[string]*[]network.ServerIdentity)
	for cn, dps := range recq.ServerToDP {
		CNsToDPs[cn] = &dps.Content
	}

	listDPs := generateDataCollectionRoster(s.ServerIdentity(), CNsToDPs)
	if listDPs == nil {
		log.Warn("no DPs to broadcast to")
	} else {
		info("broadcasting [SurveyQuery] to DPs")
		surveyToDPs := libdrynx.SurveyQueryToDP{SQ: *recq, Root: s.ServerIdentity()}
		if err := libunlynxtools.SendISMOthers(s.ServiceProcessor, listDPs, &surveyToDPs); err != nil {
			die("broadcasting [SurveyQuery] to DPs error", err)
		}
	}

	// DRO Phase
	if recq.IntraMessage == false {
		go func() {
			//diffPTimer := libDrynx.StartTimer(s.ServerIdentity().String() + "_DiffPPhase")
			if libdrynx.AddDiffP(castToSurvey(s.Survey.Get(recq.SurveyID)).SurveyQuery.Query.DiffP) {
				info("starting differential privacy proto")
				if err := s.DROPhase(castToSurvey(s.Survey.Get(recq.SurveyID)).SurveyQuery.SurveyID); err != nil {
					die("differential privacy error", err)
				}
			}
			//libDrynx.EndTimer(diffPTimer)
		}()
	}

	startDataCollectionProtocol := libunlynx.StartTimer(s.ServerIdentity().String() + "_DataCollectionProtocol")
	if listDPs != nil {
		info("starting data collection phase")
		// servers contact their DPs to get their response
		if err := s.DataCollectionPhase(recq.SurveyID); err != nil {
			die("data collection error", err)
		}
		libunlynx.EndTimer(startDataCollectionProtocol)
	}

	// ready to start the collective aggregation & key switching protocol
	if recq.IntraMessage == false {
		startJustExecution := libunlynx.StartTimer("JustExecution")
		if err := s.StartService(recq.SurveyID); err != nil {
			return nil, err
		}

		info("completed the query processing...")

		survey := castToSurvey(s.Survey.Get(recq.SurveyID))
		result := survey.QueryResponseState
		libunlynx.EndTimer(startJustExecution)

		ret := make(map[string]*libdrynx.CipherVector)
		for _, group := range result.Data {
			vec := make([]*libdrynx.CipherText, len(group.Data))

			for j, e := range group.Data {
				vec[j] = &libdrynx.CipherText{K: e.K, C: e.C}
			}

			ret[group.Group] = &libdrynx.CipherVector{Content: vec}
		}

		return &libdrynx.ResponseDP{Data: ret}, nil
	}

	return nil, nil
}

// Protocol Handlers
//______________________________________________________________________________________________________________________

// NewProtocol creates a protocol instance executed by all nodes
func (s *ServiceDrynx) NewProtocol(tn *onet.TreeNodeInstance, conf *onet.GenericConfig) (onet.ProtocolInstance, error) {
	if conf == nil {
		return nil, errors.New("conf is nil")
	}

	target := string(conf.Data)

	switch tn.ProtocolName() {
	case protocols.ProofCollectionProtocolName:
		return s.NewProofCollectionProtocolInstance(tn, target)
	case protocols.DataCollectionProtocolName:
		pi, err := protocols.NewDataCollectionProtocol(tn)
		if err != nil {
			return nil, err
		}

		dcp := pi.(*protocols.DataCollectionProtocol)
		dcp.Loader = s.loader
		dcp.Neutralizer = s.neutralizer

		if !tn.IsRoot() {
			survey := s.waitForSurvey(target)

			queryStatement := protocols.SurveyToDP{
				SurveyID:  survey.SurveyQuery.SurveyID,
				Aggregate: survey.SurveyQuery.RosterServers.Aggregate,
				Query:     survey.SurveyQuery.Query,
			}
			dcp.Survey = queryStatement
			dcp.MapPIs = survey.MapPIs
		}

		return dcp, nil

	case protocolsunlynx.CollectiveAggregationProtocolName:
		survey := s.waitForSurvey(target)
		pi, err := s.NewCollectiveAggregationProtocol(tn, survey)
		if err != nil {
			return nil, err
		}

		return pi, nil

	case protocols.ObfuscationProtocolName:
		survey := castToSurvey(s.Survey.Get(target))
		pi, err := protocols.NewObfuscationProtocol(tn)
		if err != nil {
			return nil, err
		}

		obfuscation := pi.(*protocols.ObfuscationProtocol)

		obfuscation.ToObfuscateData = *convertToCipherVector(&survey.QueryResponseState)
		obfuscation.Proofs = survey.SurveyQuery.Query.Proofs
		obfuscation.Query = &survey.SurveyQuery
		obfuscation.MapPIs = survey.MapPIs

		return pi, nil

	case protocolsunlynx.DROProtocolName:
		survey := castToSurvey(s.Survey.Get(target))
		log.Lvl2("SERVICE] <drynx> Server", s.ServerIdentity(), " Servers collectively add noise for differential privacy")
		pi, err := s.NewShufflingProtocol(tn, survey)
		if err != nil {
			return nil, err
		}

		return pi, nil

	case protocolsunlynx.KeySwitchingProtocolName:
		survey := castToSurvey(s.Survey.Get(target))
		pi, err := s.NewKeySwitchingProtocol(tn, survey)
		if err != nil {
			return nil, err
		}

		return pi, nil

	default:
		return nil, errors.New("Service attempts to start an unknown protocol: " + tn.ProtocolName() + ".")
	}
}

// NewCollectiveAggregationProtocol defines a new collective aggregation protocol
func (s *ServiceDrynx) NewCollectiveAggregationProtocol(tn *onet.TreeNodeInstance, survey Survey) (onet.ProtocolInstance, error) {
	pi, err := protocolsunlynx.NewCollectiveAggregationProtocol(tn)
	if err != nil {
		return nil, err
	}

	// convert the result to fit the collective aggregation protocol
	groupedData := libdrynx.ConvertToAggregationStruct(survey.QueryResponseState)

	collectiveAggr := pi.(*protocolsunlynx.CollectiveAggregationProtocol)
	collectiveAggr.GroupedData = &groupedData
	//TODO: change proofs
	if survey.SurveyQuery.Query.Proofs == 1 {
		collectiveAggr.Proofs = true
	} else if survey.SurveyQuery.Query.Proofs == 0 {
		collectiveAggr.Proofs = false
	}
	collectiveAggr.MapPIs = survey.MapPIs
	collectiveAggr.ProofFunc = func(data []libunlynx.CipherVector, res libunlynx.CipherVector) *libunlynxaggr.PublishedAggregationListProof {
		go func() {
			//TODO: find a better way to do this (cleaner way)
			cvMap := make(map[libunlynx.GroupingKey][]libunlynx.CipherVector)
			survey.QueryResponseState.FormatAggregationProofs(cvMap)
			aggrLocalProof := libunlynxaggr.PublishedAggregationListProof{}
			aggrLocalProof.List = make([]libunlynxaggr.PublishedAggregationProof, 0)
			for k, v := range cvMap {
				aggrLocalProof.List = append(aggrLocalProof.List, libunlynxaggr.AggregationListProofCreation(v, groupedData[k].AggregatingAttributes).List...)
			}

			pi := survey.MapPIs["aggregation/"+s.ServerIdentity().String()]
			pi.(*protocols.ProofCollectionProtocol).Proof = drynxproof.ProofRequest{AggregationProof: drynxproof.NewAggregationProofRequest(&aggrLocalProof, survey.SurveyQuery.SurveyID, s.ServerIdentity().String(), "", survey.SurveyQuery.Query.RosterVNs, tn.Private(), nil)}

			go func() {
				if err := pi.Dispatch(); err != nil {
					log.Fatal(err)
				}
			}()
			go func() {
				if err := pi.Start(); err != nil {
					log.Fatal(err)
				}
			}()
			<-pi.(*protocols.ProofCollectionProtocol).FeedbackChannel
		}()
		return nil
	}

	return pi, nil
}

// NewKeySwitchingProtocol defines a new key switching protocol
func (s *ServiceDrynx) NewKeySwitchingProtocol(tn *onet.TreeNodeInstance, survey Survey) (onet.ProtocolInstance, error) {
	pi, err := protocolsunlynx.NewKeySwitchingProtocol(tn)
	if err != nil {
		return nil, err
	}
	keySwitch := pi.(*protocolsunlynx.KeySwitchingProtocol)
	//TODO: change proofs
	if survey.SurveyQuery.Query.Proofs == 1 {
		keySwitch.Proofs = true
	} else if survey.SurveyQuery.Query.Proofs == 0 {
		keySwitch.Proofs = false
	}
	keySwitch.MapPIs = survey.MapPIs
	keySwitch.ProofFunc = func(pubKey, targetPubKey kyber.Point, secretKey kyber.Scalar, ks2s, rBNegs []kyber.Point, vis []kyber.Scalar) *libunlynxkeyswitch.PublishedKSListProof {
		go func() {
			proof, _ := libunlynxkeyswitch.KeySwitchListProofCreation(pubKey, targetPubKey, secretKey, ks2s, rBNegs, vis)
			pcp := keySwitch.MapPIs["keyswitch/"+keySwitch.ServerIdentity().String()]
			pcp.(*protocols.ProofCollectionProtocol).Proof = drynxproof.ProofRequest{KeySwitchProof: drynxproof.NewKeySwitchProofRequest(&proof, survey.SurveyQuery.SurveyID, keySwitch.ServerIdentity().String(), "", survey.SurveyQuery.Query.RosterVNs, keySwitch.Private(), nil)}
			go func() {
				if err := pcp.Dispatch(); err != nil {
					log.Fatal(err)
				}
			}()
			go func() {
				if err := pcp.Start(); err != nil {
					log.Fatal(err)
				}
			}()
			<-pcp.(*protocols.ProofCollectionProtocol).FeedbackChannel
		}()
		return nil
	}

	if tn.IsRoot() {
		if libdrynx.AddDiffP(survey.SurveyQuery.Query.DiffP) {
			for i, v := range survey.QueryResponseState.Data {
				survey.QueryResponseState.Data[i].Data.Add(v.Data, survey.Noises[:len(v.Data)])
			}
		}
		keySwitch.TargetOfSwitch = convertToCipherVector(&survey.QueryResponseState)
		tmp := survey.SurveyQuery.ClientPubKey
		keySwitch.TargetPublicKey = &tmp

		_, err = s.Survey.Put(survey.SurveyQuery.SurveyID, survey)
		if err != nil {
			return nil, err
		}

	}
	return pi, err
}

// NewShufflingProtocol defines a new shuffling protocol
func (s *ServiceDrynx) NewShufflingProtocol(tn *onet.TreeNodeInstance, survey Survey) (onet.ProtocolInstance, error) {
	pi, err := protocolsunlynx.NewShufflingProtocol(tn)
	if err != nil {
		return nil, err
	}
	shuffle := pi.(*protocolsunlynx.ShufflingProtocol)
	if survey.SurveyQuery.Query.Proofs == 1 {
		shuffle.Proofs = true
	} else if survey.SurveyQuery.Query.Proofs == 0 {
		shuffle.Proofs = false
	}
	shuffle.Precomputed = survey.ShufflePrecompute
	shuffle.MapPIs = survey.MapPIs
	shuffle.ProofFunc = func(shuffleTarget, shuffledData []libunlynx.CipherVector, collectiveKey kyber.Point, beta [][]kyber.Scalar, pi []int) *libunlynxshuffle.PublishedShufflingProof {
		go func() {
			proof, _ := libunlynxshuffle.ShuffleProofCreation(shuffleTarget, shuffledData, libunlynx.SuiTe.Point().Base(), collectiveKey, beta, pi)
			pcp := shuffle.MapPIs["shuffle/"+shuffle.ServerIdentity().String()]
			pcp.(*protocols.ProofCollectionProtocol).Proof = drynxproof.ProofRequest{ShuffleProof: drynxproof.NewShuffleProofRequest(&proof, survey.SurveyQuery.SurveyID, shuffle.ServerIdentity().String(), "", survey.SurveyQuery.Query.RosterVNs, shuffle.Private(), nil)}
			go func() {
				if err := pcp.Dispatch(); err != nil {
					log.Fatal(err)
				}
			}()
			go func() {
				if err := pcp.Start(); err != nil {
					log.Fatal(err)
				}
			}()
			<-pcp.(*protocols.ProofCollectionProtocol).FeedbackChannel
		}()
		return nil
	}

	if tn.IsRoot() {
		st := make([]libunlynx.CipherVector, 0)
		if survey.SurveyQuery.Query.DiffP.Scale == 0 {
			survey.SurveyQuery.Query.DiffP.Scale = 1
		}
		noiseArray := libunlynxdiffprivacy.GenerateNoiseValuesScale(int64(survey.SurveyQuery.Query.DiffP.NoiseListSize), survey.SurveyQuery.Query.DiffP.LapMean, survey.SurveyQuery.Query.DiffP.LapScale, survey.SurveyQuery.Query.DiffP.Quanta, survey.SurveyQuery.Query.DiffP.Scale, survey.SurveyQuery.Query.DiffP.Limit)
		for _, v := range noiseArray {
			st = append(st, libunlynx.IntArrayToCipherVector([]int64{int64(v)}))
		}
		shuffle.ShuffleTarget = &st
	}

	return pi, err
}

// StartProtocol starts a specific protocol
func (s *ServiceDrynx) StartProtocol(name string, targetSurvey string) (onet.ProtocolInstance, error) {
	// this generates the PIs of proof collection to be run inside the protocols
	tmp := castToSurvey(s.Survey.Get((string)(targetSurvey)))

	var tree *onet.Tree
	if name == protocols.DataCollectionProtocolName {
		CNsToDPs := make(map[string]*[]network.ServerIdentity)
		for cn, dps := range tmp.SurveyQuery.ServerToDP {
			CNsToDPs[cn] = &dps.Content
		}
		tree = generateDataCollectionRoster(s.ServerIdentity(), CNsToDPs).GenerateStar()
	} else {
		tree = tmp.SurveyQuery.RosterServers.GenerateBinaryTree()
	}
	if tree == nil {
		return nil, errors.New("unable to generate tree")
	}

	var tn *onet.TreeNodeInstance
	tn = s.NewTreeNodeInstance(tree, tree.Root, name)

	conf := onet.GenericConfig{Data: []byte(string(targetSurvey))}
	if err := tn.SetConfig(&conf); err != nil {
		return nil, err
	}

	pi, err := s.NewProtocol(tn, &conf)
	if err != nil {
		log.Fatal("Error running" + name)
	}

	err = s.RegisterProtocolInstance(pi)
	if err != nil {
		return nil, err
	}
	go func() {
		if err := pi.Dispatch(); err != nil {
			log.Fatal(err)
		}
	}()

	if err := pi.Start(); err != nil {
		return nil, err
	}

	return pi, nil
}

// Service Phases
//______________________________________________________________________________________________________________________

// StartService starts the service (with all its different steps/protocols)
func (s *ServiceDrynx) StartService(targetSurvey string) error {
	log.Lvl2("[SERVICE] <drynx> Server", s.ServerIdentity(), " starts a collective aggregation, (differential privacy) & key switching for survey ", targetSurvey)

	target := castToSurvey(s.Survey.Get((string)(targetSurvey)))

	// Aggregation Phase
	aggregationTimer := libunlynx.StartTimer(s.ServerIdentity().String() + "_AggregationPhase")
	err := s.AggregationPhase(target.SurveyQuery.SurveyID)
	if err != nil {
		return fmt.Errorf("Aggregation Phase: %v", err)
	}
	libunlynx.EndTimer(aggregationTimer)

	if target.SurveyQuery.Query.Obfuscation {
		//obfuscationTimer := libDrynx.StartTimer(s.ServerIdentity().String() + "_ObfuscationPhase")
		err := s.ObfuscationPhase(target.SurveyQuery.SurveyID)
		if err != nil {
			return fmt.Errorf("Obfuscation Phase: %v", err)
		}
		//libDrynx.EndTimer(obfuscationTimer)
	}

	// Key Switch Phase
	keySwitchTimer := libunlynx.StartTimer(s.ServerIdentity().String() + "_KeySwitchingPhase")
	err = s.KeySwitchingPhase(target.SurveyQuery.SurveyID)
	if err != nil {
		return fmt.Errorf("Key Switching Phase: %v", err)
	}
	libunlynx.EndTimer(keySwitchTimer)

	return nil
}

// DataCollectionPhase is the phase where data are collected from DPs
func (s *ServiceDrynx) DataCollectionPhase(targetSurvey string) error {
	pi, err := s.StartProtocol(protocols.DataCollectionProtocolName, targetSurvey)
	if err != nil {
		return err
	}
	dataDPs := <-pi.(*protocols.DataCollectionProtocol).FeedbackChannel

	survey := castToSurvey(s.Survey.Get((string)(targetSurvey)))
	// we convert the map into an object of [Group + CipherVector] to avoid later problems with protobuf
	for key, value := range dataDPs {
		if survey.SurveyQuery.Query.CuttingFactor != 0 {
			survey.QueryResponseState.Data = append(survey.QueryResponseState.Data, libdrynx.ResponseDPOneGroup{Group: key, Data: value[:int(len(value)/survey.SurveyQuery.Query.CuttingFactor)]})
		} else {
			survey.QueryResponseState.Data = append(survey.QueryResponseState.Data, libdrynx.ResponseDPOneGroup{Group: key, Data: value})

		}
	}
	_, err = s.Survey.Put(string(targetSurvey), survey)
	if err != nil {
		return err
	}
	return nil
}

// AggregationPhase performs the per-group aggregation on the currently grouped data.
func (s *ServiceDrynx) AggregationPhase(targetSurvey string) error {
	pi, err := s.StartProtocol(protocolsunlynx.CollectiveAggregationProtocolName, targetSurvey)
	if err != nil {
		return err
	}
	cothorityAggregatedData := <-pi.(*protocolsunlynx.CollectiveAggregationProtocol).FeedbackChannel

	survey := castToSurvey(s.Survey.Get((string)(targetSurvey)))

	survey.QueryResponseState = *libdrynx.ConvertFromAggregationStruct(cothorityAggregatedData)
	_, err = s.Survey.Put(string(targetSurvey), survey)
	if err != nil {
		return err
	}
	return nil
}

// ObfuscationPhase performs the obfuscation phase (multiply the aggregated data by a random value from each server)
func (s *ServiceDrynx) ObfuscationPhase(targetSurvey string) error {
	pi, err := s.StartProtocol(protocols.ObfuscationProtocolName, targetSurvey)
	if err != nil {
		return err
	}
	obfuscationData := <-pi.(*protocols.ObfuscationProtocol).FeedbackChannel

	survey := castToSurvey(s.Survey.Get((string)(targetSurvey)))
	survey.QueryResponseState = *convertFromKeySwitchingStruct(obfuscationData, survey.QueryResponseState)
	_, err = s.Survey.Put(string(targetSurvey), survey)
	if err != nil {
		return err
	}
	return nil
}

// DROPhase shuffles the list of noise values.
func (s *ServiceDrynx) DROPhase(targetSurvey string) error {
	pi, err := s.StartProtocol(protocolsunlynx.DROProtocolName, targetSurvey)
	if err != nil {
		return err
	}

	shufflingResult := <-pi.(*protocolsunlynx.ShufflingProtocol).FeedbackChannel

	survey := castToSurvey(s.Survey.Get((string)(targetSurvey)))
	noises := *libunlynx.NewCipherVector(len(shufflingResult))
	for i, v := range shufflingResult {
		noises[i] = v[0]
	}
	survey.Noises = noises
	_, err = s.Survey.Put(string(targetSurvey), survey)
	if err != nil {
		return err
	}
	return nil
}

// DROLocalPhase shuffles the list of noise values.
func (s *ServiceDrynx) DROLocalPhase(targetSurvey string) error {
	pi, err := s.StartProtocol(protocolsunlynx.DROProtocolName, targetSurvey)
	if err != nil {
		return err
	}
	shufflingResult := <-pi.(*protocolsunlynx.ShufflingProtocol).FeedbackChannel
	survey := castToSurvey(s.Survey.Get((string)(targetSurvey)))
	noises := *libunlynx.NewCipherVector(len(shufflingResult))
	for i, v := range shufflingResult {
		noises[i] = v[0]
	}
	survey.Noises = noises
	_, err = s.Survey.Put(string(targetSurvey), survey)
	if err != nil {
		return err
	}
	return nil
}

// KeySwitchingPhase performs the switch to the querier's key on the currently aggregated data.
func (s *ServiceDrynx) KeySwitchingPhase(targetSurvey string) error {
	pi, err := s.StartProtocol(protocolsunlynx.KeySwitchingProtocolName, targetSurvey)
	if err != nil {
		return err
	}
	keySwitchedAggregatedResponses := <-pi.(*protocolsunlynx.KeySwitchingProtocol).FeedbackChannel

	survey := castToSurvey(s.Survey.Get((string)(targetSurvey)))
	survey.QueryResponseState = *convertFromKeySwitchingStruct(keySwitchedAggregatedResponses, survey.QueryResponseState)
	_, err = s.Survey.Put(targetSurvey, survey)
	if err != nil {
		return err
	}
	return err
}

// Support Functions
//______________________________________________________________________________________________________________________

// these first four functions are used to adapat the existing protocols to the 'drynx' service structs
func convertToCipherVector(ad *libdrynx.ResponseAllDPs) *libunlynx.CipherVector {
	cv := make(libunlynx.CipherVector, 0)
	for _, response := range ad.Data {
		cv = append(cv, response.Data...)
	}
	return &cv
}

func convertFromKeySwitchingStruct(cv libunlynx.CipherVector, dpResponses libdrynx.ResponseAllDPs) *libdrynx.ResponseAllDPs {
	data := make([]libdrynx.ResponseDPOneGroup, 0)

	length := len(dpResponses.Data[0].Data)
	init := 0
	groupIndex := 0
	for i := 1; i <= len(cv); i++ {
		if i%length == 0 {
			tmp := cv[init:i]
			init = i
			data = append(data, libdrynx.ResponseDPOneGroup{Group: dpResponses.Data[groupIndex].Group, Data: tmp})
			groupIndex++
		}
	}
	return &libdrynx.ResponseAllDPs{Data: data}

}

func generateDataCollectionRoster(root *network.ServerIdentity, serverToDP map[string]*[]network.ServerIdentity) *onet.Roster {
	ids := serverToDP[root.String()]
	if ids == nil {
		return nil
	}

	roster := []*network.ServerIdentity{root}
	for _, id := range *ids {
		tmp := id // because go
		roster = append(roster, &tmp)
	}

	return onet.NewRoster(roster)
}
