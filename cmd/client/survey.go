package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/urfave/cli"

	"github.com/ldsec/drynx/cmd"
	libdrynx "github.com/ldsec/drynx/lib"
	"github.com/ldsec/drynx/lib/operations"
	"github.com/ldsec/drynx/services"
	_ "github.com/ldsec/drynx/services"
	kyber "go.dedis.ch/kyber/v3"
	onet "go.dedis.ch/onet/v3"
	onet_network "go.dedis.ch/onet/v3/network"
)

func surveyNew(c *cli.Context) error {
	args := c.Args()
	if len(args) != 1 {
		return errors.New("need a name")
	}
	name := args.Get(0)

	conf := config{Survey: &configSurvey{Name: &name}}

	return conf.writeTo(os.Stdout)
}

func getRoster(conf configNetwork) (onet.Roster, error) {
	ids := make([]*onet_network.ServerIdentity, len(conf.Nodes))
	for i, e := range conf.Nodes {
		e := e
		ids[i] = &e
	}

	rosterRaw := onet.NewRoster(ids)
	if rosterRaw == nil {
		return onet.Roster{}, errors.New("unable to gen roster based on config")
	}

	return *rosterRaw, nil
}

func surveySetSources(c *cli.Context) error {
	args := c.Args()
	if len(args) == 0 {
		return errors.New("usually, operation needs inputs")
	}

	conf, err := readConfigFrom(os.Stdin)
	if err != nil {
		return err
	}

	sources := make([]libdrynx.ColumnID, len(args))
	for i, a := range args {
		sources[i] = libdrynx.ColumnID(a)
	}
	conf.Survey.Sources = &sources

	return conf.writeTo(os.Stdout)
}

func surveySetOperation(c *cli.Context) error {
	args := c.Args()
	if len(args) != 1 {
		return errors.New("need an operation")
	}
	name := args[0]

	var parsedRange *cmd.Range
	if rawRange := c.String("range"); rawRange != "" {
		splitted := strings.SplitN(rawRange, ",", 2)
		if len(splitted) != 2 {
			return errors.New("range should be ','-separated")
		}

		min, err := strconv.ParseInt(splitted[0], 10, 0)
		if err != nil {
			return err
		}

		max, err := strconv.ParseInt(splitted[1], 10, 0)
		if err != nil {
			return err
		}

		parsedRange = &cmd.Range{Min: int(min), Max: int(max)}
	}

	conf, err := readConfigFrom(os.Stdin)
	if err != nil {
		return err
	}

	conf.Survey.Operation = &cmd.Operation{
		Name:  name,
		Range: parsedRange,
	}

	return conf.writeTo(os.Stdout)
}

func operationToOperation2(op cmd.Operation) (libdrynx.Operation2, error) {
	switch op.Name {
	case "frequencyCount":
		if op.Range == nil {
			return nil, errors.New("requires a range")
		}
		ret, err := operations.NewFrequencyCount(op.Range.Min, op.Range.Max)
		return &ret, err
	case "sum":
		return operations.Sum{}, nil
	case "cosim":
		return operations.CosineSimilarity{}, nil
	}

	return nil, errors.New("unknown operation name")
}

func surveyRun(c *cli.Context) error {
	if args := c.Args(); len(args) != 0 {
		return errors.New("no args expected")
	}

	conf, err := readConfigFrom(os.Stdin)
	if err != nil {
		return err
	}

	if conf.Network == nil {
		return errors.New("need some network config")
	}
	roster, err := getRoster(*conf.Network)
	if err != nil {
		return err
	}

	if conf.Network.Client == nil {
		return errors.New("no client defined")
	}
	client := services.NewDrynxClient(conf.Network.Client, os.Args[0])

	if conf.Survey == nil {
		return errors.New("need some survey config")
	}
	if conf.Survey.Name == nil {
		return errors.New("need a survey name")
	}
	if conf.Survey.Sources == nil {
		return errors.New("need some survey operation sources")
	}
	if conf.Survey.Operation == nil {
		return errors.New("need a survey operation")
	}
	opMin, opMax := 0, 0
	if opRange := conf.Survey.Operation.Range; opRange != nil {
		opMin, opMax = int(opRange.Min), int(opRange.Max)
	}
	operation := libdrynx.ChooseOperation(
		string(conf.Survey.Operation.Name), // operation
		opMin,                              // lower bound of range
		opMax,                              // upper bound of range
		len(*conf.Survey.Sources)-1,        // dimension for linear regression
		0)                                  // "cutting factor", how much to remove of gen data[0:#/n]
	if operation.NbrInput != len(*conf.Survey.Sources) {
		return errors.New("Operation can't take #Sources")
	}

	CNsToDPs := make(map[string]*libdrynx.ServerIdentityList)
	for i, cn := range roster.List {
		var dps []onet_network.ServerIdentity
		for j, dp := range roster.List {
			// TODO can't query itself
			if i != j {
				dps = append(dps, *dp)
			}
		}
		CNsToDPs[cn.String()] = &libdrynx.ServerIdentityList{Content: dps}
	}

	nodeToPublicKeys := make(map[string]kyber.Point)
	for _, node := range roster.List {
		nodeToPublicKeys[node.String()] = node.Public
	}

	query := libdrynx.Query{
		Operation:   operation,
		Ranges:      []*libdrynx.Int64List{}, // range for each output of operation
		Proofs:      int(0),                  // 0 == no proof, 1 == proof, 2 == optimized proof
		Obfuscation: false,
		DiffP: libdrynx.QueryDiffP{ // differential privacy
			LapMean: 0.0, LapScale: 0.0, NoiseListSize: 0, Quanta: 0.0, Scale: 0},
		IVSigs: libdrynx.QueryIVSigs{
			InputValidationSigs: make([]*libdrynx.PublishSignatureBytesList, 0),
		},
		RosterVNs:     &roster,
		CuttingFactor: 0,
		Selector:      *conf.Survey.Sources,
	}

	_, aggregations, err := client.SendSurveyQuery(libdrynx.SurveyQuery{
		SurveyID:      *conf.Survey.Name,
		RosterServers: roster,
		ServerToDP:    CNsToDPs,         // map CN to DPs
		IDtoPublic:    nodeToPublicKeys, // map CN|DP|VN to pub key

		Query: query,

		Threshold:                  0,
		AggregationProofThreshold:  0,
		ObfuscationProofThreshold:  0,
		RangeProofThreshold:        0,
		KeySwitchingProofThreshold: 0,
	})
	if err != nil {
		return err
	}

	if len(*aggregations) != 1 {
		return errors.New("single group expected")
	}
	for _, a := range (*aggregations)[0] {
		fmt.Println(a)
	}

	return nil
}
