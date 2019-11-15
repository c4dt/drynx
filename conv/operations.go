package conv

import (
	"errors"
	"fmt"

	"github.com/ldsec/drynx/lib"
	"github.com/ldsec/drynx/lib/operations"
)

// RangeMarshallable is a range.
type RangeMarshallable struct{ Min, Max int64 }

// OperationMarshallable represents a serialisable Operation.
type OperationMarshallable struct {
	Name  string
	Range *RangeMarshallable
}

func OperationFromMarshallable(marshallable OperationMarshallable) (libdrynx.Operation2, error) {
	if marshallable.Range == nil {
		switch marshallable.Name {
		case "sum":
			return operations.Sum{}, nil
		case "cosim":
			return operations.CosineSimilarity{}, nil
		}
	} else {
		r := *marshallable.Range
		if r.Min > r.Max {
			return nil, errors.New("min greater than max")
		}
		switch marshallable.Name {
		case "frequencyCount":
			return operations.NewFrequencyCount(r.Min, r.Max)
		}
	}

	return nil, fmt.Errorf("unable to unmarshal operation: %v", marshallable)
}

// ToOperationMarshallable converts from Operation2.
func ToOperationMarshallable(operation libdrynx.Operation2) (OperationMarshallable, error) {
	switch op := operation.(type) {
	case operations.Sum:
		return OperationMarshallable{Name: "sum"}, nil
	case operations.CosineSimilarity:
		return OperationMarshallable{Name: "cosim"}, nil
	case operations.FrequencyCount:
		min, max := op.GetMinMax()
		return OperationMarshallable{
			Name:  "frequencyCount",
			Range: &RangeMarshallable{min, max},
		}, nil
	}

	return OperationMarshallable{}, fmt.Errorf("unable to marshal operation: %v", operation)
}
