package operations

import (
	"errors"

	"github.com/ldsec/drynx/lib/encoding"
	"github.com/ldsec/unlynx/lib"

	"go.dedis.ch/kyber/v3"
)

const sumInputSize = 1

// Sum computes the accumulation of values in a column.
type Sum struct{}

// ApplyOnProvider encodes.
func (s Sum) ApplyOnProvider(key kyber.Point, loaded [][]float64) (libunlynx.CipherVector, error) {
	if len(loaded) != sumInputSize {
		return nil, errors.New("unexpected number of columns")
	}

	converted := floatsToInts(loaded[0])

	// TODO add support for proof
	encoded, _ := libdrynxencoding.EncodeSum(converted, key)
	return libunlynx.CipherVector{*encoded}, nil
}

// ApplyOnClient decodes.
func (s Sum) ApplyOnClient(key kyber.Scalar, aggregated libunlynx.CipherVector) ([]float64, error) {
	if len(aggregated) != 1 {
		return nil, errors.New("unexpected size of aggregated vector")
	}

	return []float64{float64(libdrynxencoding.DecodeSum(aggregated[0], key))}, nil
}

// GetInputSize returns 1.
func (Sum) GetInputSize() uint {
	return sumInputSize
}
