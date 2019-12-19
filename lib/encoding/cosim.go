package libdrynxencoding

import (
	"github.com/ldsec/drynx/lib"
	"github.com/ldsec/drynx/lib/range"
	"github.com/ldsec/unlynx/lib"
	"go.dedis.ch/kyber/v3"
	"math"
)

// ExecuteCosimOnProvider computes the result to encode, under the cosim operation.
func ExecuteCosimOnProvider(rijs, riks []int64) []int64 {
	//sum the rijs
	rijsSum := int64(0)
	riksSum := int64(0)
	rijs2Sum := int64(0)
	riks2Sum := int64(0)
	rijsXRijksSum := int64(0)

	for i, el := range rijs {
		el2 := riks[i]
		rijsSum = rijsSum + el
		riksSum = riksSum + el2
		rijs2Sum = rijs2Sum + el*el
		riks2Sum = riks2Sum + el2*el2
		rijsXRijksSum = rijsXRijksSum + el*el2

	}

	return []int64{rijsSum, riksSum, rijs2Sum, riks2Sum, rijsXRijksSum}
}

// ExecuteCosimOnClient computes the result from the aggregated results, under the cosim operation.
func ExecuteCosimOnClient(aggregated []int64) float64 {
	return float64(aggregated[4]) / (math.Sqrt(float64(aggregated[2])) * math.Sqrt(float64(aggregated[3])))
}

// EncodeCosim computes the elements needed to compute cosine similarity
func EncodeCosim(rijs, riks []int64, pubKey kyber.Point) ([]libunlynx.CipherText, []int64) {
	resultEnc, resultClear, _ := EncodeCosimWithProofs(rijs, riks, pubKey, nil, []*libdrynx.Int64List{})
	return resultEnc, resultClear
}

// EncodeCosimWithProofs computes the elements needed to compute cosine similarity with the proof of range
func EncodeCosimWithProofs(rijs, riks []int64, pubKey kyber.Point, sigs [][]libdrynx.PublishSignature, lu []*libdrynx.Int64List) ([]libunlynx.CipherText, []int64, []libdrynxrange.CreateProof) {
	resultClear := ExecuteCosimOnProvider(rijs, riks)

	resultEncrypteds := make([]libunlynx.CipherText, len(resultClear))
	resultRandomRS := make([]kyber.Scalar, len(resultClear))
	wg := libunlynx.StartParallelize(len(resultClear))
	for i, v := range resultClear {
		go func(i int, v int64) {
			defer wg.Done()
			tmp, r := libunlynx.EncryptIntGetR(pubKey, v)
			resultEncrypteds[i] = *tmp
			resultRandomRS[i] = r
		}(i, v)

	}
	libunlynx.EndParallelize(wg)

	if sigs == nil {
		return resultEncrypteds, resultClear, nil
	}

	createProofs := make([]libdrynxrange.CreateProof, len(resultClear))
	wg = libunlynx.StartParallelize(len(resultClear))
	for i, v := range resultClear {
		go func(i int, v int64) {
			defer wg.Done()
			//input range validation proof
			createProofs[i] = libdrynxrange.CreateProof{Sigs: libdrynxrange.ReadColumn(sigs, i), U: (*lu[i]).Content[0], L: (*lu[i]).Content[1], Secret: v, R: resultRandomRS[i], CaPub: pubKey, Cipher: resultEncrypteds[i]}
		}(i, v)
	}
	libunlynx.EndParallelize(wg)

	return resultEncrypteds, resultClear, createProofs
}

// DecodeCosim decodes (decrypts and computes) the cosine similarity result
func DecodeCosim(result []libunlynx.CipherText, secKey kyber.Scalar) float64 {
	resultsClears := make([]int64, len(result))
	wg := libunlynx.StartParallelize(len(result))
	for i, j := range result {
		go func(i int, j libunlynx.CipherText) {
			defer wg.Done()
			resultsClears[i] = libunlynx.DecryptIntWithNeg(secKey, j)
		}(i, j)

	}
	libunlynx.EndParallelize(wg)

	return ExecuteCosimOnClient(resultsClears)

}
