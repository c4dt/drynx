package conv

import (
	"github.com/ldsec/drynx/lib"

	onet_network "go.dedis.ch/onet/v3/network"
)

// SurveyQueryToDP is a serializable libdrynx.SurveyQueryToDP.
type SurveyQueryToDP struct {
	SQ   SurveyQuery
	Root *onet_network.ServerIdentity
}

// ToSurveyQueryToDP deserialize.
func (sqm SurveyQueryToDP) ToSurveyQueryToDP() (libdrynx.SurveyQueryToDP, error) {
	sq, err := sqm.SQ.ToSurveyQuery()
	return libdrynx.SurveyQueryToDP{
		SQ:   sq,
		Root: sqm.Root,
	}, err
}

// FromSurveyQueryToDP serialize.
func FromSurveyQueryToDP(sq libdrynx.SurveyQueryToDP) (SurveyQueryToDP, error) {
	sqm, err := FromSurveyQuery(sq.SQ)
	return SurveyQueryToDP{
		SQ:   sqm,
		Root: sq.Root,
	}, err
}
