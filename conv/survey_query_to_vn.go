package conv

import (
	"github.com/ldsec/drynx/lib"
)

type SurveyQueryToVNMarshallable struct {
	SQ SurveyQueryMarshallable
}

func SurveyQueryToVNFromMarshallable(marshallable SurveyQueryToVNMarshallable) (libdrynx.SurveyQueryToVN, error) {
	sq, err := SurveyQueryFromMarshallable(marshallable.SQ)
	return libdrynx.SurveyQueryToVN{SQ: sq}, err
}

func ToSurveyQueryToVNMarshallable(sq libdrynx.SurveyQueryToVN) (SurveyQueryToVNMarshallable, error) {
	sqm, err := ToSurveyQueryMarshallable(sq.SQ)
	return SurveyQueryToVNMarshallable{sqm}, err
}
