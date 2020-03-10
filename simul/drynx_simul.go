package main

import (
	"github.com/ldsec/drynx/lib/range"
	"go.dedis.ch/kyber/v3"
	"os"

	"sync"

	"time"

	"github.com/BurntSushi/toml"
	"github.com/ldsec/drynx/lib"
	"github.com/ldsec/drynx/lib/provider/loaders"
	"github.com/ldsec/drynx/lib/provider/neutralizers"
	"github.com/ldsec/drynx/services"
	"github.com/ldsec/unlynx/lib"
	"go.dedis.ch/cothority/v3/skipchain"
	"go.dedis.ch/onet/v3"
	"go.dedis.ch/onet/v3/log"
	"go.dedis.ch/onet/v3/network"
	"gopkg.in/satori/go.uuid.v1"
)

// SimulationDrynx state of a simulation.
type SimulationDrynx struct {
	onet.SimulationBFTree
	// Topology
	NbrServers      int
	NbrVNs          int
	NbrDPs          int
	NbrDPsPerServer int

	//Log Reg
	NbrRecords int

	//Proofs
	Proofs           int
	Ranges           int
	InputValidation  bool
	Obfuscation      bool
	ThresholdGeneral float64
	ThresholdOther   float64

	// Query
	OperationName string
	NbrInput      int
	NbrOutput     int

	//DiffP
	DiffPEpsilon float64
	DiffPDelta   float64
	DiffPSize    int
	DiffPQuanta  float64
	DiffPScale   float64
	DiffPLimit   float64
	DiffPOpti    bool

	// Data and query response
	GroupByValues []int64
	DPRows        uint
	MinData       int64
	MaxData       int64

	CuttingFactor int
	MaxIterations int
}

// NewSimulationDrynx constructs a full Drynx service simulation.
func NewSimulationDrynx(config string) (onet.Simulation, error) {
	sl := &SimulationDrynx{}
	_, err := toml.Decode(config, sl)
	if err != nil {
		return nil, err
	}

	loader, err := loaders.NewRandom(float64(sl.MinData), float64(sl.MaxData), sl.DPRows)
	if err != nil {
		panic(err)
	}
	services.NewBuilder().
		WithComputingNode().
		WithDataProvider(loader, neutralizers.NewMinimumResultsSize(0)).
		WithVerifyingNode().
		Start()

	return sl, nil
}

// Setup creates the tree used for that simulation
func (sim *SimulationDrynx) Setup(dir string, hosts []string) (*onet.SimulationConfig, error) {
	sc := &onet.SimulationConfig{}
	sim.CreateRoster(sc, hosts, 2000)
	err := sim.CreateTree(sc)

	if err != nil {
		return nil, err
	}

	log.Lvl2("Setup done")

	return sc, nil
}

// Run starts the simulation.
func (sim *SimulationDrynx) Run(config *onet.SimulationConfig) error {
	os.Remove("pre_compute_multiplications.gob")

	// has to be set here because cannot be in toml file
	diffP := libdrynx.QueryDiffP{LapMean: sim.DiffPEpsilon, LapScale: sim.DiffPDelta, Quanta: sim.DiffPQuanta, NoiseListSize: sim.DiffPSize, Scale: sim.DiffPScale, Limit: sim.DiffPLimit}

	//logistic regression
	m := int64(sim.DPRows) - 1
	means := make([]float64, m)
	stds := make([]float64, m)
	for i := 0; i < int(m); i++ {
		stds[i] = 100
	}

	lrParameters := libdrynx.LogisticRegressionParameters{
		FilePath:                    "",
		NbrRecords:                  int64(sim.NbrRecords),
		NbrFeatures:                 m,
		Means:                       means,
		StandardDeviations:          stds,
		K:                           2,
		PrecisionApproxCoefficients: 1,
		Lambda:                      1.0,
		Step:                        0.1,
		MaxIterations:               sim.MaxIterations,
		InitialWeights:              make([]float64, m+1),
	}

	// operation
	operation := libdrynx.Operation{NameOp: sim.OperationName, NbrInput: sim.NbrInput, NbrOutput: sim.NbrOutput, QueryMin: sim.MinData, QueryMax: sim.MaxData, LRParameters: lrParameters}

	// create the ranges for input validation
	ranges := make([]*libdrynx.Int64List, operation.NbrOutput)

	switch sim.Ranges {
	case -1:
		ranges = nil
	case 16:
		for i := range ranges {
			u := int64(16)
			l := int64(16)
			ranges[i] = &libdrynx.Int64List{Content: []int64{u, l}}
		}
	case 99:
		for i := range ranges {
			var u, l int64
			if i%3 == 0 {
				u = int64(16)
				l = int64(2)
			} else if i%3 == 1 {
				u = int64(16)
				l = int64(4)
			} else if i%3 == 2 {
				u = int64(2)
				l = int64(1)
			} else {
				log.Fatal("You are not running the variance you naughty boy!")
			}
			ranges[i] = &libdrynx.Int64List{Content: []int64{u, l}}
		}
	case 100:
		for i := range ranges {
			var u, l int64
			if i%3 == 0 {
				u = int64(4)
				l = int64(11)
			} else if i%3 == 1 {
				u = int64(4)
				l = int64(7)
			} else if i%3 == 2 {
				u = int64(4)
				l = int64(3)
			} else {
				log.Fatal("You are not running the variance you naughty boy!")
			}
			ranges[i] = &libdrynx.Int64List{Content: []int64{u, l}}
		}
	case 101:
		for i := range ranges {
			var u, l int64
			if i%3 == 0 {
				u = int64(4)
				l = int64(13)
			} else if i%3 == 1 {
				u = int64(4)
				l = int64(9)
			} else if i%3 == 2 {
				u = int64(4)
				l = int64(5)
			} else {
				log.Fatal("You are not running the variance you naughty boy!")
			}
			ranges[i] = &libdrynx.Int64List{Content: []int64{u, l}}
		}
	case 102:
		for i := range ranges {
			var u, l int64
			if i%3 == 0 {
				u = int64(4)
				l = int64(15)
			} else if i%3 == 1 {
				u = int64(4)
				l = int64(11)
			} else if i%3 == 2 {
				u = int64(4)
				l = int64(7)
			} else {
				log.Fatal("You are not running the variance you naughty boy!")
			}
			ranges[i] = &libdrynx.Int64List{Content: []int64{u, l}}
		}
	case 103:
		for i := range ranges {
			var u, l int64
			if i%3 == 0 {
				u = int64(4)
				l = int64(17)
			} else if i%3 == 1 {
				u = int64(4)
				l = int64(13)
			} else if i%3 == 2 {
				u = int64(4)
				l = int64(9)
			} else {
				log.Fatal("You are not running the variance you naughty boy!")
			}
			ranges[i] = &libdrynx.Int64List{Content: []int64{u, l}}
		}
	case 1:
		for i := range ranges {
			u := int64(2)
			l := int64(1)
			ranges[i] = &libdrynx.Int64List{Content: []int64{u, l}}
		}
	case 0:
		for i := range ranges {
			u := int64(0)
			l := int64(0)
			ranges[i] = &libdrynx.Int64List{Content: []int64{u, l}}
		}
	case 17:
		for i := range ranges {
			u := int64(8)
			l := int64(3)
			/*if sim.CuttingFactor != 0 {
				if i != 0 && i%(sim.NbrOutput/sim.CuttingFactor) == 0 {
					log.Lvl1("[0,1]")
					u = int64(2)
					l = int64(1)
				}
			}
			if i == len(ranges)-1 {
				log.Lvl1("[0,1]")
				u = int64(2)
				l = int64(1)
			}*/

			ranges[i] = &libdrynx.Int64List{Content: []int64{u, l}}
		}
	case 18:
		for i := range ranges {
			u := int64(16)
			l := int64(5)
			ranges[i] = &libdrynx.Int64List{Content: []int64{u, l}}
		}
	case 19:
		for i := range ranges {
			u := int64(4)
			l := int64(16)
			ranges[i] = &libdrynx.Int64List{Content: []int64{u, l}}
		}
	}

	// signatures for Input Validation
	ps := make([]*libdrynx.PublishSignatureBytesList, sim.NbrServers)
	if !(ranges == nil) && sim.Ranges != 0 {
		wg := libunlynx.StartParallelize(sim.NbrServers)
		for i := 0; i < sim.NbrServers; i++ {
			go func(index int) {
				defer wg.Done()
				temp := make([]libdrynx.PublishSignatureBytes, len(ranges))
				for j := 0; j < len(ranges); j++ {
					if sim.CuttingFactor != 0 {
						temp[j] = libdrynxrange.InitRangeProofSignatureDeterministic((*ranges[j]).Content[0])
					} else {
						temp[j] = libdrynxrange.InitRangeProofSignature((*ranges[j]).Content[0]) // u is the first elem
					}
				}
				ps[index] = &libdrynx.PublishSignatureBytesList{Content: temp}
				log.Lvl1("Finished creating signatures for server", index)
			}(i)
		}
		libunlynx.EndParallelize(wg)
	} else {
		ps = nil
	}

	//define servers and data providers in the set of nodes + adapt the aggregate key (CA public key)
	elTotal := (*config.Tree).Roster
	elServers := elTotal.List[:sim.NbrServers]
	elVNs := elTotal.List[sim.NbrServers : sim.NbrServers+sim.NbrVNs]
	elDPs := elTotal.List[sim.NbrServers+sim.NbrVNs : sim.NbrServers+sim.NbrVNs+sim.NbrDPs]

	if sim.NbrDPs%sim.NbrDPsPerServer != 0 {
		log.Fatal("The total number of servers must match the number of servers per data provider")
	}

	dpToServers := make(map[string]*[]network.ServerIdentity)
	dpIndex := 0
	for _, v := range elServers {
		if dpIndex < len(elDPs) {
			key := v.String()
			value := make([]network.ServerIdentity, sim.NbrDPsPerServer)
			dpToServers[key] = &value
			for j := range *dpToServers[key] {
				if dpIndex < len(elDPs) {
					val := elDPs[dpIndex]
					(*dpToServers[key])[j] = *val
				}
				dpIndex++
			}
		}

	}

	rosterServers := onet.NewRoster(elServers)
	rosterVNs := onet.NewRoster(elVNs)

	idToPublic := make(map[string]kyber.Point)
	for _, v := range rosterServers.List {
		idToPublic[v.String()] = v.ServicePublic(services.ServiceName)
	}
	if rosterVNs != nil {
		for _, v := range rosterVNs.List {
			idToPublic[v.String()] = v.ServicePublic(services.ServiceName)
		}
	}

	for _, v := range elDPs {
		idToPublic[v.String()] = v.ServicePublic(services.ServiceName)
	}

	// Create a client (querier) for the service)
	client := services.NewDrynxClient(rosterServers.List[0], "simul-Drynx")
	// query generation
	surveyID := uuid.NewV4().String()

	var thresholdEntityProofsVerif []float64
	if sim.Obfuscation {
		thresholdEntityProofsVerif = []float64{sim.ThresholdGeneral, sim.ThresholdOther, sim.ThresholdOther, sim.ThresholdOther, sim.ThresholdOther}
	} else {
		thresholdEntityProofsVerif = []float64{sim.ThresholdGeneral, sim.ThresholdOther, sim.ThresholdOther, 0.0, sim.ThresholdOther}
	}
	sq := client.GenerateSurveyQuery(rosterServers, rosterVNs, dpToServers, idToPublic, surveyID, operation, ranges, ps, sim.Proofs, sim.Obfuscation, thresholdEntityProofsVerif, diffP, sim.CuttingFactor)
	if !libdrynx.CheckParameters(sq, diffP.NoiseListSize > 0) {
		log.Fatal("Oups!")
	}

	overallTimer := time.Now()
	startSimulation := libunlynx.StartTimer("Simulation")
	var wg *sync.WaitGroup
	var block *skipchain.SkipBlock
	var err error

	if sim.Proofs != 0 {
		// send query to the skipchain and 'wait' for all proofs' verification to be done
		clientSkip := services.NewDrynxClient(elVNs[0], "simul-skip-"+sim.OperationName)

		wg = libunlynx.StartParallelize(1)
		go func(elVNs *onet.Roster) {
			defer wg.Done()

			err := clientSkip.SendSurveyQueryToVNs(elVNs, &sq)
			if err != nil {
				log.Fatal("Error sending query to VNs:", err)
			}
		}(rosterVNs)
		libunlynx.EndParallelize(wg)

		wg = libunlynx.StartParallelize(1)
		go func(si *network.ServerIdentity) {
			defer wg.Done()

			block, err = clientSkip.SendEndVerification(si, surveyID)
			if err != nil {
				log.Fatal("Error starting the 'waiting' threads:", err)
			}
		}(elVNs[0])
	}

	// send query and receive results
	grp, aggr, err := client.SendSurveyQuery(sq)

	if err != nil {
		log.Fatal("'Drynx' service did not start.", err)
	}

	// Result printing
	if len(*grp) != 0 && len(*grp) != len(*aggr) {
		log.Fatal("Results format problem")
	} else {
		for i, v := range *aggr {
			log.Lvl1((*grp)[i], ": ", v)
		}
	}

	if len(elVNs) > 0 {
		clientSkip := services.NewDrynxClient(elVNs[0], "simul-skip")
		if sim.Proofs != 0 {
			libunlynx.EndParallelize(wg)
			// close DB
			if err := clientSkip.SendCloseDB(rosterVNs, &libdrynx.CloseDB{Close: 1}); err != nil {
				log.Fatal("Error closing the DB:", err)
			}
		}

		retrieveBlock := time.Now()
		sb, err := clientSkip.SendGetLatestBlock(rosterVNs, block)
		if err != nil || sb == nil {
			log.Fatal("Something wrong when fetching the last block")
		}
		log.Lvl1(time.Since(retrieveBlock))
	}

	libunlynx.EndTimer(startSimulation)
	log.Lvl1(time.Since(overallTimer))

	return nil
}
