package conv

import (
	"github.com/ldsec/drynx/lib"

	onet_network "go.dedis.ch/onet/v3/network"
)

type SurveyQueryToDPMarshallable struct {
	SQ   SurveyQueryMarshallable
	Root *onet_network.ServerIdentity
}

func SurveyQueryToDPFromMarshallable(sqm SurveyQueryToDPMarshallable) (libdrynx.SurveyQueryToDP, error) {
	sq, err := SurveyQueryFromMarshallable(sqm.SQ)
	return libdrynx.SurveyQueryToDP{
		SQ:   sq,
		Root: sqm.Root,
	}, err
}

func ToSurveyQueryToDPMarshallable(sq libdrynx.SurveyQueryToDP) (SurveyQueryToDPMarshallable, error) {
	sqm, err := ToSurveyQueryMarshallable(sq.SQ)
	return SurveyQueryToDPMarshallable{
		SQ:   sqm,
		Root: sq.Root,
	}, err
}
