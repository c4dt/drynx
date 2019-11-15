package conv

import (
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/onet/v3"
	onet_network "go.dedis.ch/onet/v3/network"

	"github.com/ldsec/drynx/lib"
)

type SurveyQueryMarshallable struct {
	Query QueryMarshallable

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

func SurveyQueryFromMarshallable(marshallable SurveyQueryMarshallable) (libdrynx.SurveyQuery, error) {
	query, err := QueryFromMarshallable(marshallable.Query)
	return libdrynx.SurveyQuery{
		Query:                      query,
		SurveyID:                   marshallable.SurveyID,
		RosterServers:              marshallable.RosterServers,
		ClientPubKey:               marshallable.ClientPubKey,
		IntraMessage:               marshallable.IntraMessage,
		ServerToDP:                 marshallable.ServerToDP,
		IDtoPublic:                 marshallable.IDtoPublic,
		Threshold:                  marshallable.Threshold,
		AggregationProofThreshold:  marshallable.AggregationProofThreshold,
		ObfuscationProofThreshold:  marshallable.ObfuscationProofThreshold,
		RangeProofThreshold:        marshallable.RangeProofThreshold,
		KeySwitchingProofThreshold: marshallable.KeySwitchingProofThreshold,
	}, err
}

func ToSurveyQueryMarshallable(sq libdrynx.SurveyQuery) (SurveyQueryMarshallable, error) {
	q, err := ToQueryMarshallable(sq.Query)
	return SurveyQueryMarshallable{
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
