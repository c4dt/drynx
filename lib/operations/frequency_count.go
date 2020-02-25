package operations

import (
	"errors"

	"github.com/ldsec/drynx/lib/encoding"
)

const fcInputSize = 1

// FrequencyCount computes the sum of occurence of values in a column.
type FrequencyCount struct{ Range }

// MarshalID is the Operation's ID.
func (FrequencyCount) MarshalID() [8]byte {
	ret := [8]byte{}
	copy(ret[:], []byte("dr.op.fc"))
	return ret
}

// NewFrequencyCount creates a new FrequencyCount bound to the given range.
func NewFrequencyCount(min, max int) (FrequencyCount, error) {
	if min > max {
		return FrequencyCount{}, errors.New("given minimum is greater than maximum")
	}
	return FrequencyCount{Range{min, max}}, nil
}

// ExecuteOnProvider encodes.
func (fc FrequencyCount) ExecuteOnProvider(loaded [][]float64) ([]float64, error) {
	if len(loaded) != fcInputSize {
		return nil, errors.New("unexpected number of columns")
	}

	converted := floatsToInts(loaded[0])
	freqCount := libdrynxencoding.ExecuteFreqCountOnProvider(converted, int64(fc.min), int64(fc.max))
	ret := make([]float64, len(freqCount))
	for i, v := range freqCount {
		ret[i] = float64(v)
	}
	return ret, nil
}

// ExecuteOnClient decodes.
func (fc FrequencyCount) ExecuteOnClient(aggregated []float64) ([]float64, error) {
	if uint(len(aggregated)) != fc.GetEncodedSize() {
		return nil, errors.New("unexpected size of aggregated vector")
	}

	freqCount := libdrynxencoding.ExecuteFreqCountOnClient(floatsToInts(aggregated))
	ret := make([]float64, len(freqCount))
	for i, v := range freqCount {
		ret[i] = float64(v)
	}
	return ret, nil
}

// GetInputSize returns 1.
func (FrequencyCount) GetInputSize() uint {
	return fcInputSize
}

// GetEncodedSize returns the size of the CipherVector.
func (fc FrequencyCount) GetEncodedSize() uint {
	return uint(fc.max-fc.min) + 1
}
