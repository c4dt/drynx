package operations

import (
	"errors"

	"github.com/ldsec/drynx/lib/encoding"
)

const linearRegressionInputSize = 1
const linearRegressionEncodedSize = 1

// LinearRegression computes the accumulation of values in a column.
type LinearRegression struct{}

// MarshalID is the Operation's ID.
func (LinearRegression) MarshalID() [8]byte {
	ret := [8]byte{}
	copy(ret[:], []byte("dr.op.su"))
	return ret
}

// MarshalBinary returns nil.
func (LinearRegression) MarshalBinary() ([]byte, error) {
	return nil, nil
}

// UnmarshalBinary does nothing.
func (LinearRegression) UnmarshalBinary([]byte) error {
	return nil
}

// ExecuteOnProvider encodes.
func (s LinearRegression) ExecuteOnProvider(loaded [][]float64) ([]float64, error) {
	if len(loaded) != linearRegressionInputSize {
		return nil, errors.New("unexpected number of columns")
	}

	converted := floats2DToInts(loaded)

	linearRegression := libdrynxencoding.ExecuteLinearRegressionOnProvider(converted)
	return ints1DToFloats(linearRegression), nil
}

// ExecuteOnClient decodes.
func (s LinearRegression) ExecuteOnClient(aggregated []float64) ([]float64, error) {
	if len(aggregated) != linearRegressionEncodedSize {
		return nil, errors.New("unexpected size of aggregated vector")
	}

	return libdrynxencoding.ExecuteLinearRegressionOnClient(floats1DToInts(aggregated)), nil
}

// GetInputSize returns 1.
func (LinearRegression) GetInputSize() uint {
	return linearRegressionInputSize
}

// GetEncodedSize returns 1.
func (LinearRegression) GetEncodedSize() uint {
	return linearRegressionEncodedSize
}
