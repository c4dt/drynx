package operations

import (
	"bytes"
	"encoding/binary"
)

func floatsToInts(arr []float64) []int64 {
	ret := make([]int64, len(arr))
	for i, v := range arr {
		ret[i] = int64(v)
	}
	return ret
}

func intsToFloats(arr []int64) []float64 {
	ret := make([]float64, len(arr))
	for i, v := range arr {
		ret[i] = float64(v)
	}
	return ret
}

// Range represents a width between two int64
type Range struct {
	min, max int64
}

// MarshalBinary encodes to binary
func (r Range) MarshalBinary() ([]byte, error) {
	buffer := new(bytes.Buffer)
	if err := binary.Write(buffer, binary.BigEndian, r.min); err != nil {
		return nil, err
	}
	if err := binary.Write(buffer, binary.BigEndian, r.max); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// UnmarshalBinary decodes from MarshalBinary
func (r *Range) UnmarshalBinary(buf []byte) error {
	buffer := bytes.NewBuffer(buf)
	if err := binary.Read(buffer, binary.BigEndian, &r.min); err != nil {
		return err
	}
	if err := binary.Read(buffer, binary.BigEndian, &r.max); err != nil {
		return err
	}
	return nil
}
