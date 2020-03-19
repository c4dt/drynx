package operations

import (
	"bytes"
	"encoding/binary"
	"errors"
)

func floats1DToInts(arr []float64) []int64 {
	ret := make([]int64, len(arr))
	for i, v := range arr {
		ret[i] = int64(v)
	}
	return ret
}

func ints1DToFloats(arr []int64) []float64 {
	ret := make([]float64, len(arr))
	for i, v := range arr {
		ret[i] = float64(v)
	}
	return ret
}

func floats2DToInts(arr [][]float64) [][]int64 {
	ret := make([][]int64, len(arr))
	for i, v := range arr {
		ret[i] = floats1DToInts(v)
	}
	return ret
}

func ints2DToFloats(arr [][]int64) [][]float64 {
	ret := make([][]float64, len(arr))
	for i, v := range arr {
		ret[i] = ints1DToFloats(v)
	}
	return ret
}

// Range represents a width between two int
type Range struct{ min, max int }

// NewRange construct a Range instance
func NewRange(min, max int) (Range, error) {
	if min > max {
		return Range{}, errors.New("minimum of Range is greater than maximum")
	}

	return Range{min, max}, nil
}

// MarshalBinary encodes to binary
func (r Range) MarshalBinary() ([]byte, error) {
	buffer := new(bytes.Buffer)
	if err := binary.Write(buffer, binary.BigEndian, int64(r.min)); err != nil {
		return nil, err
	}
	if err := binary.Write(buffer, binary.BigEndian, int64(r.max)); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// UnmarshalBinary decodes from MarshalBinary
func (r *Range) UnmarshalBinary(buf []byte) error {
	buffer := bytes.NewBuffer(buf)
	var min, max int64
	if err := binary.Read(buffer, binary.BigEndian, &min); err != nil {
		return err
	}
	if err := binary.Read(buffer, binary.BigEndian, &max); err != nil {
		return err
	}

	var err error
	var nextRange Range
	if nextRange, err = NewRange(int(min), int(max)); err != nil {
		return err
	}
	*r = nextRange

	return nil
}
