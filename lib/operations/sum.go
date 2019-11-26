package operations

import (
	"errors"

	"github.com/ldsec/drynx/lib/encoding"
)

const sumInputSize = 1
const sumEncodedSize = 1

// Sum computes the accumulation of values in a column.
type Sum struct{}

// ExecuteOnProvider encodes.
func (s Sum) ExecuteOnProvider(loaded [][]float64) ([]float64, error) {
	if len(loaded) != sumInputSize {
		return nil, errors.New("unexpected number of columns")
	}

	converted := floatsToInts(loaded[0])

	sum := libdrynxencoding.ExecuteSumOnProvider(converted)
	return []float64{float64(sum)}, nil
}

// ExecuteOnClient decodes.
func (s Sum) ExecuteOnClient(aggregated []float64) ([]float64, error) {
	if len(aggregated) != sumEncodedSize {
		println("SUM", aggregated)
		for _, v := range aggregated {
			println("SUM aggregated:", v)
		}
		return nil, errors.New("unexpected size of aggregated vector")
	}

	return []float64{float64(libdrynxencoding.ExecuteSumOnClient(floatsToInts(aggregated)))}, nil
}

// GetInputSize returns 1.
func (Sum) GetInputSize() uint {
	return sumInputSize
}

// GetEncodedSize returns 1.
func (Sum) GetEncodedSize() uint {
	return sumEncodedSize
}
