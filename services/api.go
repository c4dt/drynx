package services

import (
	"github.com/ldsec/drynx/lib"
	"github.com/ldsec/drynx/lib/encoding"
	"github.com/ldsec/drynx/lib/obfuscation"
	"github.com/ldsec/drynx/lib/range"
	"github.com/ldsec/unlynx/lib"
	"github.com/ldsec/unlynx/lib/aggregation"
	"github.com/ldsec/unlynx/lib/key_switch"
	"github.com/ldsec/unlynx/lib/shuffle"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/util/key"
	"go.dedis.ch/onet/v3"
	"go.dedis.ch/onet/v3/log"
	"go.dedis.ch/onet/v3/network"
)

// API represents a client with the server to which he is connected and its public/private key pair.
type API struct {
	*onet.Client
	clientID   string
	entryPoint *network.ServerIdentity
	public     kyber.Point
	private    kyber.Scalar
}

// NewDrynxClient constructor of a client.
func NewDrynxClient(entryPoint *network.ServerIdentity, clientID string) *API {
	network.RegisterMessage(libdrynx.GetLatestBlock{})
	network.RegisterMessage(libdrynxrange.RangeProofListBytes{})
	network.RegisterMessage(libunlynxshuffle.PublishedShufflingProofBytes{})
	network.RegisterMessage(libunlynxkeyswitch.PublishedKSListProofBytes{})
	network.RegisterMessage(libunlynxaggr.PublishedAggregationListProofBytes{})
	network.RegisterMessage(libdrynxobfuscation.PublishedListObfuscationProofBytes{})

	keys := key.NewKeyPair(libunlynx.SuiTe)
	newClient := &API{
		Client:     onet.NewClient(libdrynx.Suite, ServiceName),
		clientID:   clientID,
		entryPoint: entryPoint,
		public:     keys.Public,
		private:    keys.Private,
	}

	limit := int64(10000)
	libunlynx.CreateDecryptionTable(limit, newClient.public, newClient.private)
	return newClient
}

// Send Query
//______________________________________________________________________________________________________________________

// GenerateSurveyQuery generates a query with all the information in parameters
func (c *API) GenerateSurveyQuery(rosterServers, rosterVNs *onet.Roster, dpToServer map[string]*[]network.ServerIdentity, idToPublic map[string]kyber.Point, surveyID string, operation libdrynx.Operation, ranges []*libdrynx.Int64List, ps []*libdrynx.PublishSignatureBytesList, proofs int, obfuscation bool, thresholds []float64, diffP libdrynx.QueryDiffP, cuttingFactor int) libdrynx.SurveyQuery {
	iVSigs := libdrynx.QueryIVSigs{InputValidationSigs: ps}
	query := libdrynx.Query{
		Selector: make([]libdrynx.ColumnID, operation.NbrInput),

		Operation:   operation,
		Ranges:      ranges,
		DiffP:       diffP,
		Proofs:      proofs,
		Obfuscation: obfuscation,

		// identity blockchain infos
		IVSigs:        iVSigs,
		RosterVNs:     rosterVNs,
		CuttingFactor: cuttingFactor,
	}

	CNsToDPs := make(map[string]*libdrynx.ServerIdentityList)
	for cn, dps := range dpToServer {
		if dps != nil {
			CNsToDPs[cn] = &libdrynx.ServerIdentityList{Content: *dps}
		}
	}

	return libdrynx.SurveyQuery{
		SurveyID:                   surveyID,
		RosterServers:              *rosterServers,
		IntraMessage:               false,
		ServerToDP:                 CNsToDPs,
		IDtoPublic:                 idToPublic,
		Threshold:                  thresholds[0],
		AggregationProofThreshold:  thresholds[1],
		RangeProofThreshold:        thresholds[2],
		ObfuscationProofThreshold:  thresholds[3],
		KeySwitchingProofThreshold: thresholds[4],

		Query: query,
	}
}

// SendSurveyQuery creates a survey based on a set of entities (servers) and a survey description.
func (c *API) SendSurveyQuery(sq libdrynx.SurveyQuery) (*[]string, *[][]float64, error) {
	log.Lvl2("[API] <Drynx> Client", c.clientID, "is creating a query with SurveyID: ", sq.SurveyID)

	if sq.ClientPubKey == nil {
		sq.ClientPubKey = c.public
	}

	//send the query and get the answer
	sr := libdrynx.ResponseDP{}
	err := c.SendProtobuf(c.entryPoint, &sq, &sr)
	if err != nil {
		return nil, nil, err
	}

	log.Lvl2("[API] <Drynx> Client", c.clientID, "successfully executed the query with SurveyID ", sq.SurveyID)

	// decrypt/decode the result
	clientDecode := libunlynx.StartTimer("Decode")
	log.Lvl2("[API] <Drynx> Client", c.clientID, "is decrypting the results")

	grp := make([]string, len(sr.Data))
	aggr := make([][]float64, len(sr.Data))
	count := 0
	for i, res := range sr.Data {
		vec := make(libunlynx.CipherVector, len(res.Content))
		for j, e := range res.Content {
			vec[j] = libunlynx.CipherText{K: e.K, C: e.C}
		}
		grp[count] = i
		aggr[count] = libdrynxencoding.Decode(vec, c.private, sq.Query.Operation)
	}
	libunlynx.EndTimer(clientDecode)

	log.Lvl2("[API] <Drynx> Client", c.clientID, "finished decrypting the results")
	return &grp, &aggr, nil
}
