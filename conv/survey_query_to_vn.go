package conv

import (
	"github.com/ldsec/drynx/lib"
)

// SurveyQueryToVN is a serializable libdrynx.SurveyQueryToVN.
type SurveyQueryToVN struct {
	SQ SurveyQuery
}

// ToSurveyQueryToVN deserialize.
func (sqvn SurveyQueryToVN) ToSurveyQueryToVN() (libdrynx.SurveyQueryToVN, error) {
	sq, err := sqvn.SQ.ToSurveyQuery()
	return libdrynx.SurveyQueryToVN{SQ: sq}, err
}

// FromSurveyQueryToVN serialize.
func FromSurveyQueryToVN(sq libdrynx.SurveyQueryToVN) (SurveyQueryToVN, error) {
	sqm, err := FromSurveyQuery(sq.SQ)
	return SurveyQueryToVN{sqm}, err
}
