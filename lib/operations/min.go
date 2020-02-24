package operations

import (
	"errors"
)

const minInputSize = 1
const minEncodedSize = 1

// Min computes the accumulation of values in a column.
type Min struct{ Range }

// MarshalID is the Operation's ID.
func (Min) MarshalID() [8]byte {
	ret := [8]byte{}
	copy(ret[:], []byte("dr.op.mi"))
	return ret
}

// MarshalBinary returns nil.
func (Min) MarshalBinary() ([]byte, error) {
	return nil, nil
}

// UnmarshalBinary does nothing.
func (Min) UnmarshalBinary([]byte) error {
	return nil
}

// ExecuteOnProvider encodes.
func (s Min) ExecuteOnProvider(loaded [][]float64) ([]float64, error) {
	if len(loaded) != minInputSize {
		return nil, errors.New("unexpected number of columns")
	}

	// TODO
	//converted := floatsToInts(loaded[0])
	//min := libdrynxencoding.ExecuteMinOnProvider(converted)

	return []float64{float64(0)}, nil
}

// ExecuteOnClient decodes.
func (s Min) ExecuteOnClient(aggregated []float64) ([]float64, error) {
	if len(aggregated) != minEncodedSize {
		return nil, errors.New("unexpected size of aggregated vector")
	}

	// TODO
	//return []float64{float64(libdrynxencoding.ExecuteMinOnClient(floatsToInts(aggregated)))}, nil
	return nil, nil
}

// GetInputSize returns 1.
func (Min) GetInputSize() uint {
	return minInputSize
}

// GetEncodedSize returns 1.
func (Min) GetEncodedSize() uint {
	return minEncodedSize
}
