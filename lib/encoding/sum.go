package libdrynxencoding

import (
	"github.com/ldsec/drynx/lib"
	"github.com/ldsec/drynx/lib/range"
	"github.com/ldsec/unlynx/lib"
	"go.dedis.ch/kyber/v3"
)

// EncodeSum computes the sum of query results
func EncodeSum(input []int64, pubKey kyber.Point) (*libunlynx.CipherText, []int64) {
	resultEnc, resultClear, _ := EncodeSumWithProofs(input, pubKey, nil, 0, 0)
	return resultEnc, resultClear
}

// ExecuteSumOnProvider computes the result to encode.
func ExecuteSumOnProvider(input []int64) int64 {
	var sum int64
	for _, el := range input {
		sum += el
	}
	return sum
}

// ExecuteSumOnClient computes the result from the aggregated results.
func ExecuteSumOnClient(agg []int64) int64 {
	return agg[0]
}

// EncodeSumWithProofs computes the sum of query results with the proof of range
func EncodeSumWithProofs(input []int64, pubKey kyber.Point, sigs []libdrynx.PublishSignature, l int64, u int64) (*libunlynx.CipherText, []int64, []libdrynxrange.CreateProof) {
	sum := ExecuteSumOnProvider(input)

	//encrypt the local DP's query result
	sumEncrypted, r := libunlynx.EncryptIntGetR(pubKey, sum)

	if sigs == nil {
		return sumEncrypted, []int64{sum}, nil
	}
	//input range validation proof
	cp := libdrynxrange.CreateProof{Sigs: sigs, U: u, L: l, Secret: sum, R: r, CaPub: pubKey, Cipher: *sumEncrypted}

	return sumEncrypted, []int64{sum}, []libdrynxrange.CreateProof{cp}
}

// DecodeSum computes the sum of local DP's query results
func DecodeSum(result libunlynx.CipherText, secKey kyber.Scalar) int64 {
	//decrypt the query results
	return libunlynx.DecryptIntWithNeg(secKey, result)

}
