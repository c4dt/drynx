package operations

import (
	"errors"

	"github.com/pelletier/go-toml"

	"github.com/ldsec/drynx/lib/encoding"
)

const cosimInputSize = 2
const cosimEncodedSize = 5

// CosineSimilarity computes the cosine similarity between two columns.
type CosineSimilarity struct{}

func (CosineSimilarity) MarshalID() [8]byte {
	ret := [8]byte{}
	copy(ret[:], []byte("dr.op.cs"))
	return ret
}

func (CosineSimilarity) MarshalBinary() ([]byte, error) {
	return nil, nil
}

func (CosineSimilarity) UnmarshalBinary([]byte) error {
	return nil
}

// ExecuteOnProvider executes.
func (CosineSimilarity) ExecuteOnProvider(loaded [][]float64) ([]float64, error) {
	if len(loaded) != cosimInputSize {
		return nil, errors.New("unexpected number of columns")
	}
	vec1, vec2 := floatsToInts(loaded[0]), floatsToInts(loaded[1])

	return intsToFloats(libdrynxencoding.ExecuteCosimOnProvider(vec1, vec2)), nil
}

func (CosineSimilarity) MarshalText() ([]byte, error) {
	return toml.Marshal(Operation{Name: "CosineSimilarity"})
}

func (CosineSimilarity) UnmarshalText([]byte) error {
	return nil
}

// ExecuteOnClient computes.
func (CosineSimilarity) ExecuteOnClient(aggregated []float64) ([]float64, error) {
	if len(aggregated) != cosimEncodedSize {
		return nil, errors.New("unexpected size of aggregated vector")
	}

	return []float64{libdrynxencoding.ExecuteCosimOnClient(floatsToInts(aggregated))}, nil
}

// GetInputSize returns 2.
func (CosineSimilarity) GetInputSize() uint {
	return cosimInputSize
}

// GetEncodedSize returns 5.
func (CosineSimilarity) GetEncodedSize() uint {
	return cosimEncodedSize
}
