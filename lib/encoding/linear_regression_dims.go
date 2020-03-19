package libdrynxencoding

import (
	"fmt"
	"github.com/alex-ant/gomath/gaussian-elimination"
	"github.com/alex-ant/gomath/rational"
	"github.com/ldsec/drynx/lib"
	"github.com/ldsec/drynx/lib/range"
	"github.com/ldsec/unlynx/lib"
	"github.com/tonestuff/quadratic"
	"go.dedis.ch/kyber/v3"
	"math"
	"time"
)

//EncodeLinearRegressionDims implements a d-dimensional linear regression algorithm on the query results
func EncodeLinearRegressionDims(datas [][]int64, pubKey kyber.Point) ([]libunlynx.CipherText, []int64) {
	resultEnc, resultClear, _ := EncodeLinearRegressionDimsWithProofs(datas, pubKey, nil, nil)
	return resultEnc, resultClear
}

// ExecuteLinearRegressionOnProvider computes the result to encode.
func ExecuteLinearRegressionOnProvider(input [][]int64) []int64 {
	N := len(input)
	d := len(input[0]) - 1

	input1 := make([][]int64, N)
	input2 := make([]int64, N)
	for i, v := range input {
		input1[i] = v[:len(v)-1]
		input2[i] = v[len(v)-1]
	}

	//sum the Xs and their squares, the Ys and the product of every pair of X and Y
	sumXj := int64(0)
	sumY := int64(0)
	sumXjY := int64(0)
	sumXjX := int64(0)

	plaintextValues := []int64{int64(N)}
	var StoredVals []int64

	//loop over dimensions
	for j := 0; j < d; j++ {
		sumXj = int64(0)
		sumXjY = int64(0)
		for i := 0; i < N; i++ {
			x := input1[i][j]
			sumXj += x
			sumXjY += input2[i] * x
		}
		plaintextValues = append(plaintextValues, sumXj)
		StoredVals = append(StoredVals, sumXjY)
	}

	for j := 0; j < d; j++ {
		for k := j; k < d; k++ {
			sumXjX = int64(0)
			for i := 0; i < N; i++ {
				sumXjX += input1[i][j] * input1[i][k]
			}
			plaintextValues = append(plaintextValues, sumXjX)
		}
	}

	for _, el := range input2 {
		sumY += el
	}
	plaintextValues = append(plaintextValues, sumY)

	for j := 0; j < len(StoredVals); j++ {
		plaintextValues = append(plaintextValues, StoredVals[j])
	}

	return plaintextValues
}

//EncodeLinearRegressionDimsWithProofs implements a d-dimensional linear regression algorithm on the query results with range proofs
func EncodeLinearRegressionDimsWithProofs(datas [][]int64, pubKey kyber.Point, sigs [][]libdrynx.PublishSignature, lu []*libdrynx.Int64List) ([]libunlynx.CipherText, []int64, []libdrynxrange.CreateProof) {
	plaintextValues := ExecuteLinearRegressionOnProvider(datas)

	encoded := make([]libunlynx.CipherText, len(plaintextValues))
	var createProofs []libdrynxrange.CreateProof
	for i, v := range plaintextValues {
		encrypted, r := libunlynx.EncryptIntGetR(pubKey, v)
		encoded[i] = *encrypted

		if sigs != nil {
			createProofs = append(createProofs, libdrynxrange.CreateProof{
				Sigs:   libdrynxrange.ReadColumn(sigs, i),
				U:      lu[i].Content[0],
				L:      lu[i].Content[1],
				Secret: v,
				R:      r,
				CaPub:  pubKey,
				Cipher: *encrypted,
			})
		}
	}

	return encoded, plaintextValues, createProofs
}

//DecodeLinearRegressionDims implements a d-dimensional linear regression algorithm, in this encoding, we assume the system to have a perfect solution
//TODO least-square computation and not equality
func DecodeLinearRegressionDims(result []libunlynx.CipherText, secKey kyber.Scalar) []float64 {
	decoded := make([]int64, len(result))
	for i, v := range result {
		decoded[i] = libunlynx.DecryptIntWithNeg(secKey, v)
	}

	return ExecuteLinearRegressionOnClient(decoded)
}

// ExecuteLinearRegressionOnClient computes the result from the aggregated results.
func ExecuteLinearRegressionOnClient(result []int64) []float64 {
	//get the the number of dimensions by solving the equation: d^2 + 5d + 4 = 2*len(result)
	posSol, _ := quadratic.Solve(1, 5, complex128(complex(float32(4-2*len(result)), 0)))
	d := int(real(posSol))

	matrixAugmented := make([][]int64, d+1, d+2)
	for i := range matrixAugmented {
		matrixAugmented[i] = make([]int64, d+2)
	}

	//Build the augmented matrix
	s := 0
	l := d + 1
	k := d + 1
	i := 0
	for j := 0; j < len(result)-d-1; j++ {
		if j == l {
			k--
			l = l + k
			i++
			s = 0
		}
		matrixAugmented[i][i+s] = result[j]
		if i != i+s {
			matrixAugmented[i+s][i] = result[j]
		}
		s++
	}

	for j := len(result) - d - 1; j < len(result); j++ {
		matrixAugmented[j-len(result)+d+1][d+1] = result[j]
	}

	matrixRational := make([][]rational.Rational, d+1, d+2)
	for i := range matrixAugmented {
		matrixRational[i] = make([]rational.Rational, d+2)
	}
	for i := range matrixAugmented {
		for j := 0; j < d+2; j++ {
			matrixRational[i][j] = rational.New(matrixAugmented[i][j], 1)
		}
	}

	//Solve the linear system of equations and return x = [c0, c1, c2, ..., cd]
	var solution [][]rational.Rational
	solution, _ = gaussian.SolveGaussian(matrixRational, false)

	coeffs := make([]float64, d+1)
	for i := 0; i < len(solution); i++ {
		coeffs[i] = solution[i][0].Float64()
	}
	return coeffs
}

func h(weights []float64, x []float64) float64 {
	h := weights[0]
	for i := 0; i < len(x); i++ {
		h += weights[i+1] * x[i]
	}
	return h
}

// CostLinearRegression implements the cost function for the linear regression [TEST]
func CostLinearRegression(weights []float64, X [][]float64, y []float64) float64 {
	m := len(X)

	cost := float64(1 / (2 * m))
	sum := float64(0)
	for i, sample := range X {
		sum += math.Pow(h(weights, sample)-y[i], 2)
	}
	return cost * sum
}

// GradientLinearRegression implements the gradient descent algorithm for the linear regression [TEST]
func GradientLinearRegression(weights []float64, X [][]float64, y []float64, lambda float64) []float64 {
	dim := len(X[0])
	m := float64(len(X))
	gradients := make([]float64, dim)

	for i := 0; i < dim; i++ {
		gradientI := float64(0)
		for j, sample := range X {
			if i == 0 {
				gradientI += h(weights, sample) - y[j]
			} else {
				gradientI += (h(weights, sample) - y[j]) * sample[j]
			}
		}
		gradientI = gradientI * lambda / m
		gradients[i] = gradientI
	}
	return gradients
}

// FindMinimumWeightsLinearRegression runs a linear regression (to find the mininum weigths) [TEST]
func FindMinimumWeightsLinearRegression(initialWeights []float64, X [][]float64, y []float64, lambda float64, maxIterations int) []float64 {

	//weights := initialWeights
	weights := make([]float64, len(initialWeights))
	copy(weights, initialWeights)

	minCost := math.MaxFloat64
	minWeights := make([]float64, len(weights))

	start := time.Now()
	timeout := time.Duration(60 * 3 * time.Second)
	epsilon := time.Duration(2 * time.Second)

	for iter := 0; iter < maxIterations; iter++ {
		cost := CostLinearRegression(weights, X, y)

		if cost >= 0.0 {
			minCost = cost
			copy(minWeights, weights)
		}

		gradient := GradientLinearRegression(weights, X, y, lambda)
		for i := 0; i < len(weights); i++ {
			weights[i] = weights[i] - lambda*gradient[i]
		}

		if iter%int(float64(maxIterations)/10.0) == 0 {
			fmt.Printf("%6d cost, min. cost: %12.8f %12.8f \n", iter, cost, minCost)
		}

		t := time.Now()
		elapsed := t.Sub(start)
		if timeout-elapsed < epsilon {
			fmt.Println("elapsed:", elapsed)
			return minWeights
		}
	}

	return minWeights
}
