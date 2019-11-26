package conv

import (
	"errors"
	"fmt"

	"github.com/ldsec/drynx/lib"
	"github.com/ldsec/drynx/lib/operations"
)

// Range is a segment.
type Range struct{ Min, Max int64 }

// Operation represents a serialisable Operation.
type Operation struct {
	Name  string
	Range *Range
}

// ToOperation converts to Operation2.
func (op Operation) ToOperation() (libdrynx.Operation2, error) {
	if op.Range == nil {
		switch op.Name {
		case "sum":
			return operations.Sum{}, nil
		case "cosim":
			return operations.CosineSimilarity{}, nil
		}
	} else {
		r := *op.Range
		if r.Min > r.Max {
			return nil, errors.New("min greater than max")
		}
		switch op.Name {
		case "frequencyCount":
			return operations.NewFrequencyCount(r.Min, r.Max)
		}
	}

	return nil, fmt.Errorf("unable to unmarshal operation: %v", op)
}

// FromOperation converts from Operation2.
func FromOperation(operation libdrynx.Operation2) (Operation, error) {
	switch op := operation.(type) {
	case operations.Sum:
		return Operation{Name: "sum"}, nil
	case operations.CosineSimilarity:
		return Operation{Name: "cosim"}, nil
	case operations.FrequencyCount:
		min, max := op.GetMinMax()
		return Operation{
			Name:  "frequencyCount",
			Range: &Range{min, max},
		}, nil
	}

	return Operation{}, fmt.Errorf("unable to marshal operation: %v", operation)
}
