package operations

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/ldsec/drynx/lib/encoding"
)

const fcInputSize = 1

// FrequencyCount computes the sum of occurence of values in a column.
type FrequencyCount struct{ min, max int64 }

// MarshalID is the Operation's ID.
func (FrequencyCount) MarshalID() [8]byte {
	ret := [8]byte{}
	copy(ret[:], []byte("dr.op.fc"))
	return ret
}

// MarshalBinary serialise the instance into []byte.
func (fc FrequencyCount) MarshalBinary() ([]byte, error) {
	buffer := new(bytes.Buffer)
	if err := binary.Write(buffer, binary.BigEndian, fc.min); err != nil {
		return nil, err
	}
	if err := binary.Write(buffer, binary.BigEndian, fc.max); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// UnmarshalBinary deserialise []byte, as given by MarshalBinary, into the instance.
func (fc *FrequencyCount) UnmarshalBinary(buf []byte) error {
	buffer := bytes.NewBuffer(buf)
	if err := binary.Read(buffer, binary.BigEndian, &fc.min); err != nil {
		return err
	}
	if err := binary.Read(buffer, binary.BigEndian, &fc.max); err != nil {
		return err
	}
	return nil
}

// NewFrequencyCount creates a new FrequencyCount bound to the given range.
func NewFrequencyCount(min, max int64) (FrequencyCount, error) {
	if min > max {
		return FrequencyCount{}, errors.New("given minimum is greater than maximum")
	}
	return FrequencyCount{min, max}, nil
}

// ExecuteOnProvider encodes.
func (fc FrequencyCount) ExecuteOnProvider(loaded [][]float64) ([]float64, error) {
	if len(loaded) != fcInputSize {
		return nil, errors.New("unexpected number of columns")
	}

	converted := floatsToInts(loaded[0])
	freqCount := libdrynxencoding.ExecuteFreqCountOnProvider(converted, fc.min, fc.max)
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
