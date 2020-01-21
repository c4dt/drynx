package operations

import (
	"errors"

	"github.com/ldsec/drynx/lib/encoding"
	"github.com/ldsec/unlynx/lib"

	"go.dedis.ch/kyber/v3"
)

const meanInputSize = 1
const meanEncodedSize = 2

// Mean computes the average value of a column.
type Mean struct{}

// MarshalID is the Operation's ID.
func (Mean) MarshalID() [8]byte {
	ret := [8]byte{}
	copy(ret[:], []byte("dr.op.me"))
	return ret
}

// MarshalBinary returns nil.
func (Mean) MarshalBinary() ([]byte, error) {
	return nil, nil
}

// UnmarshalBinary does nothing.
func (Mean) UnmarshalBinary([]byte) error {
	return nil
}

// ExecuteOnProvider encodes.
func (Mean) ExecuteOnProvider(key kyber.Point, loaded [][]float64) (libunlynx.CipherVector, error) {
	if len(loaded) != meanInputSize {
		return nil, errors.New("unexpected number of columns")
	}

	// TODO add support for proof
	encoded, _ := libdrynxencoding.EncodeMean(floatsToInts(loaded[0]), key)
	return libunlynx.CipherVector(encoded), nil
}

// ExecuteOnClient decodes.
func (Mean) ExecuteOnClient(key kyber.Scalar, aggregated libunlynx.CipherVector) ([]float64, error) {
	if len(aggregated) != meanEncodedSize {
		return nil, errors.New("unexpected size of aggregated vector")
	}

	return []float64{libdrynxencoding.DecodeMean(aggregated, key)}, nil
}

// GetInputSize returns 1.
func (Mean) GetInputSize() uint {
	return meanInputSize
}

// GetEncodedSize returns 2.
func (Mean) GetEncodedSize() uint {
	return meanEncodedSize
}
