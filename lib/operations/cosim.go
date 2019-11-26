package operations

import (
	"errors"

	"github.com/ldsec/drynx/lib/encoding"
	"github.com/ldsec/unlynx/lib"

	"go.dedis.ch/kyber/v3"
)

const cosimInputSize = 2
const cosimEncodedSize = 5

// CosineSimilarity computes the cosine similarity between two columns.
type CosineSimilarity struct{}

// ApplyOnProvider encodes.
func (CosineSimilarity) ApplyOnProvider(key kyber.Point, loaded [][]float64) (libunlynx.CipherVector, error) {
	if len(loaded) != cosimInputSize {
		return nil, errors.New("unexpected number of columns")
	}
	vec1, vec2 := floatsToInts(loaded[0]), floatsToInts(loaded[1])

	// TODO add support for proof
	encoded, _ := libdrynxencoding.EncodeCosim(vec1, vec2, key)
	return libunlynx.CipherVector(encoded), nil
}

// ApplyOnClient decodes.
func (CosineSimilarity) ApplyOnClient(key kyber.Scalar, aggregated libunlynx.CipherVector) ([]float64, error) {
	if len(aggregated) != cosimEncodedSize {
		return nil, errors.New("unexpected size of aggregated vector")
	}

	return []float64{float64(libdrynxencoding.DecodeCosim(aggregated, key))}, nil
}

// GetInputSize returns 2.
func (CosineSimilarity) GetInputSize() uint {
	return cosimInputSize
}

// GetEncodedSize returns 5.
func (CosineSimilarity) GetEncodedSize() uint {
	return cosimEncodedSize
}
