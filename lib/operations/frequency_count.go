package operations

import (
	"errors"

	"github.com/ldsec/drynx/lib/encoding"
	"github.com/ldsec/unlynx/lib"

	"go.dedis.ch/kyber/v3"
)

const fcInputSize = 1

// FrequencyCount computes the sum of occurence of values in a column.
type FrequencyCount struct{ min, max int64 }

// NewFrequencyCount creates a new FrequencyCount bound to the given range.
func NewFrequencyCount(min, max int64) (FrequencyCount, error) {
	if min > max {
		return FrequencyCount{}, errors.New("given minimum is greater than maximum")
	}
	return FrequencyCount{min, max}, nil
}

// ApplyOnProvider encodes.
func (fc FrequencyCount) ApplyOnProvider(key kyber.Point, loaded [][]float64) (libunlynx.CipherVector, error) {
	if len(loaded) != fcInputSize {
		return nil, errors.New("unexpected number of columns")
	}

	converted := floatsToInts(loaded[0])

	encoded, _ := libdrynxencoding.EncodeFreqCount(converted, fc.min, fc.max, key)
	return encoded, nil
}

// ApplyOnClient decodes.
func (fc FrequencyCount) ApplyOnClient(key kyber.Scalar, aggregated libunlynx.CipherVector) ([]float64, error) {
	if uint(len(aggregated)) != fc.GetEncodedSize() {
		return nil, errors.New("unexpected size of aggregated vector")
	}

	return intsToFloats(libdrynxencoding.DecodeFreqCount(aggregated, key)), nil
}

// GetMinMax returns the given min and max.
func (fc FrequencyCount) GetMinMax() (int64, int64) {
	return fc.min, fc.max
}

// GetInputSize returns 1.
func (FrequencyCount) GetInputSize() uint {
	return fcInputSize
}

// GetEncodedSize returns the size of the CipherVector.
func (fc FrequencyCount) GetEncodedSize() uint {
	return uint(fc.max-fc.min) + 1
}
