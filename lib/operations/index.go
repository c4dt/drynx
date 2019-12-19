package operations

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

type Range struct{ Min, Max int64 }

type Operation struct {
	Name  string
	Range *Range
}
