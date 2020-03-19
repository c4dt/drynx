package operations

import (
	"errors"

	"github.com/ldsec/drynx/lib/encoding"
)

const cosimInputSize = 2
const cosimEncodedSize = 5

// CosineSimilarity computes the cosine similarity between two columns.
type CosineSimilarity struct{}

// MarshalID is the Operation's ID.
func (CosineSimilarity) MarshalID() [8]byte {
	ret := [8]byte{}
	copy(ret[:], []byte("dr.op.cs"))
	return ret
}

// MarshalBinary returns nil.
func (CosineSimilarity) MarshalBinary() ([]byte, error) {
	return nil, nil
}

// UnmarshalBinary does nothing.
func (CosineSimilarity) UnmarshalBinary([]byte) error {
	return nil
}

// ExecuteOnProvider executes.
func (CosineSimilarity) ExecuteOnProvider(loaded [][]float64) ([]float64, error) {
	if len(loaded) != cosimInputSize {
		return nil, errors.New("unexpected number of columns")
	}
	vec1, vec2 := floats1DToInts(loaded[0]), floats1DToInts(loaded[1])

	return ints1DToFloats(libdrynxencoding.ExecuteCosimOnProvider(vec1, vec2)), nil
}

// ExecuteOnClient computes.
func (CosineSimilarity) ExecuteOnClient(aggregated []float64) ([]float64, error) {
	if len(aggregated) != cosimEncodedSize {
		return nil, errors.New("unexpected size of aggregated vector")
	}

	return []float64{libdrynxencoding.ExecuteCosimOnClient(floats1DToInts(aggregated))}, nil
}

// GetInputSize returns 2.
func (CosineSimilarity) GetInputSize() uint {
	return cosimInputSize
}

// GetEncodedSize returns 5.
func (CosineSimilarity) GetEncodedSize() uint {
	return cosimEncodedSize
}
