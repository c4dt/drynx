package conv

import (
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/onet/v3"
	onet_network "go.dedis.ch/onet/v3/network"

	"github.com/ldsec/drynx/lib"
)

// SurveyQuery is a serializable libdrynx.SurveyQuery.
type SurveyQuery struct {
	Query Query

	// from old SurveyQuery

	SurveyID      string
	RosterServers onet.Roster
	ClientPubKey  kyber.Point
	IntraMessage  bool // to define whether the query was sent by the querier or not
	ServerToDP    map[string]*[]onet_network.ServerIdentity

	//map of DP/Server to Public key
	IDtoPublic map[string]kyber.Point

	//Threshold for verification in skipChain service
	Threshold                  float64
	AggregationProofThreshold  float64
	ObfuscationProofThreshold  float64
	RangeProofThreshold        float64
	KeySwitchingProofThreshold float64
}

// ToSurveyQuery deserialize.
func (sq SurveyQuery) ToSurveyQuery() (libdrynx.SurveyQuery, error) {
	query, err := sq.Query.ToQuery()
	return libdrynx.SurveyQuery{
		Query:                      query,
		SurveyID:                   sq.SurveyID,
		RosterServers:              sq.RosterServers,
		ClientPubKey:               sq.ClientPubKey,
		IntraMessage:               sq.IntraMessage,
		ServerToDP:                 sq.ServerToDP,
		IDtoPublic:                 sq.IDtoPublic,
		Threshold:                  sq.Threshold,
		AggregationProofThreshold:  sq.AggregationProofThreshold,
		ObfuscationProofThreshold:  sq.ObfuscationProofThreshold,
		RangeProofThreshold:        sq.RangeProofThreshold,
		KeySwitchingProofThreshold: sq.KeySwitchingProofThreshold,
	}, err
}

// FromSurveyQuery serialize.
func FromSurveyQuery(sq libdrynx.SurveyQuery) (SurveyQuery, error) {
	q, err := FromQuery(sq.Query)
	return SurveyQuery{
		Query:                      q,
		SurveyID:                   sq.SurveyID,
		RosterServers:              sq.RosterServers,
		ClientPubKey:               sq.ClientPubKey,
		IntraMessage:               sq.IntraMessage,
		ServerToDP:                 sq.ServerToDP,
		IDtoPublic:                 sq.IDtoPublic,
		Threshold:                  sq.Threshold,
		AggregationProofThreshold:  sq.AggregationProofThreshold,
		ObfuscationProofThreshold:  sq.ObfuscationProofThreshold,
		RangeProofThreshold:        sq.RangeProofThreshold,
		KeySwitchingProofThreshold: sq.KeySwitchingProofThreshold,
	}, err
}
