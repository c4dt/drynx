package libdrynxencoding

import (
	"github.com/ldsec/drynx/lib"
	"github.com/ldsec/drynx/lib/range"
	"github.com/ldsec/unlynx/lib"
	"go.dedis.ch/kyber/v3"
)

//EncodeFreqCount computes the frequency count of query results
//Note: min and max are such that all values are in the range [min, max], i.e. max (min) is the largest (smallest) possible value the attribute in question can take
func EncodeFreqCount(input []int64, min int64, max int64, pubKey kyber.Point) ([]libunlynx.CipherText, []int64) {
	resultEnc, resultClear, _ := EncodeFreqCountWithProofs(input, min, max, pubKey, nil, nil)
	return resultEnc, resultClear
}

// ExecuteFreqCountOnProvider computes the result to encode.
func ExecuteFreqCountOnProvider(input []int64, min int64, max int64) []uint64 {
	freqcount := make([]uint64, max-min+1)
	for _, el := range input {
		if el < min || el > max {
			panic("found out of range data")
		}
		freqcount[el-min]++
	}

	return freqcount
}

// ExecuteFreqCountOnClient computes the result from the aggregated results.
func ExecuteFreqCountOnClient(input []int64) []uint64 {
	ret := make([]uint64, len(input))
	for i, v := range input {
		if v < 0 {
			panic("frequency count < 0")
		}
		ret[i] = uint64(v)
	}
	return ret
}

// EncodeFreqCountWithProofs computes the frequency count of query results with the proof of range
func EncodeFreqCountWithProofs(input []int64, min int64, max int64, pubKey kyber.Point, sigs [][]libdrynx.PublishSignature, lu []*[]int64) ([]libunlynx.CipherText, []int64, []libdrynxrange.CreateProof) {
	freqcountUint := ExecuteFreqCountOnProvider(input, min, max)
	freqcount := make([]int64, len(freqcountUint))
	for i, v := range freqcountUint {
		freqcount[i] = int64(v)
	}

	//encrypt the local DP's query results
	ciphertextTuples := make([]libunlynx.CipherText, max-min+1)
	wg := libunlynx.StartParallelize(int(max-min) + 1)
	r := make([]kyber.Scalar, len(freqcount))
	for i := int64(0); i <= max-min; i++ {
		go func(i int64) {
			defer wg.Done()
			countIEncrypted, ri := libunlynx.EncryptIntGetR(pubKey, int64(freqcount[i]))
			r[i] = ri
			ciphertextTuples[i] = *countIEncrypted
		}(i)

	}
	libunlynx.EndParallelize(wg)

	if sigs == nil {
		return ciphertextTuples, freqcount, nil
	}

	createRangeProof := make([]libdrynxrange.CreateProof, len(freqcount))
	wg1 := libunlynx.StartParallelize(len(freqcount))
	for i, v := range freqcount {
		go func(i int, v int64) {
			defer wg1.Done()
			//input range validation proof
			createRangeProof[i] = libdrynxrange.CreateProof{Sigs: libdrynxrange.ReadColumn(sigs, i), U: (*lu[i])[0], L: (*lu[i])[1], Secret: v, R: r[i], CaPub: pubKey, Cipher: ciphertextTuples[i]}
		}(i, v)
	}
	libunlynx.EndParallelize(wg1)
	return ciphertextTuples, freqcount, createRangeProof
}

//DecodeFreqCount computes the frequency count of local DP's query results
func DecodeFreqCount(result []libunlynx.CipherText, secKey kyber.Scalar) []int64 {
	PlaintextTuple := make([]int64, len(result))

	//get the counts for all integer values in the range {1, 2, ..., max}
	wg := libunlynx.StartParallelize(len(result))
	for i := int64(0); i < int64(len(result)); i++ {
		go func(i int64) {
			defer wg.Done()
			PlaintextTuple[i] = libunlynx.DecryptIntWithNeg(secKey, result[i])
		}(i)

	}
	libunlynx.EndParallelize(wg)

	//return the array of counts
	return PlaintextTuple
}
