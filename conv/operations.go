package conv

import (
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

// OperationFromMarshallable safely converts to Operation2.
func OperationFromMarshallable(marshallable OperationMarshallable) (libdrynx.Operation2, error) {
	switch marshallable {
	case "sum":
		return operations.Sum{}, nil
	case "cosim":
		return operations.CosineSimilarity{}, nil
	case "frequencyCount":
		return operations.FrequencyCount{}, nil
	default:
		return nil, fmt.Errorf("unable to unmarshal operation: %v", marshallable)
	}
}

// ToOperationMarshallable converts from Operation2.
func ToOperationMarshallable(op libdrynx.Operation2) (OperationMarshallable, error) {
	switch op.(type) {
	case operations.Sum:
		return "sum", nil
	case operations.CosineSimilarity:
		return "cosim", nil
	case operations.FrequencyCount:
		return "frequencyCount", nil
	}

	return "", fmt.Errorf("unable to marshal operation: %v", op)
}
