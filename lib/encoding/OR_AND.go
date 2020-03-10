package libdrynxencoding

import (
	"math/rand"

	"github.com/ldsec/drynx/lib"
	"github.com/ldsec/drynx/lib/range"
	"github.com/ldsec/unlynx/lib"
	"go.dedis.ch/kyber/v3"
)

//d in this case is (modulus - 2 = 2^255 - 19 - 2 = 2^255 - 21)
//This is because the random number R we want to generate should be in the set {1, 2, ..., modulus -1}
//But in this case with golang, we are generating it from the range [0, modulus-2] and then adding 1 to it
// to make sure that it belongs to the range [1, modulus - 1]

// ExecuteBitOr computes the result to encode, under the OR operation.
func ExecuteBitOr(input bool, wantProof bool) int64 {
	if !input {
		return 0
	}

	if wantProof {
		return 1
	}

	return rand.Int63()
}

//EncodeBitOr computes the encoding of bit Xi, under the OR operation
func EncodeBitOr(input bool, pubKey kyber.Point) (*libunlynx.CipherText, int64) {
	cipher, clear, _ := EncodeBitOrWithProof(input, pubKey, nil, 0, 0)
	return cipher, clear
}

//EncodeBitOrWithProof computes the encoding of bit Xi, under the OR operation with range proofs
func EncodeBitOrWithProof(input bool, pubKey kyber.Point, sigs []libdrynx.PublishSignature, l int64, u int64) (*libunlynx.CipherText, int64, libdrynxrange.CreateProof) {
	toEncrypt := ExecuteBitOr(input, sigs != nil)
	cipher, r := libunlynx.EncryptIntGetR(pubKey, toEncrypt)

	cp := libdrynxrange.CreateProof{}
	if sigs != nil {
		cp = libdrynxrange.CreateProof{Sigs: sigs, U: u, L: l, Secret: toEncrypt, R: r, CaPub: pubKey, Cipher: *cipher}
	}

	return cipher, toEncrypt, cp
}

//DecodeBitOR computes the decoding of bit Xi, under the OR operation
func DecodeBitOR(result libunlynx.CipherText, secKey kyber.Scalar) bool {
	//decrypt the bit representation
	output := libunlynx.DecryptCheckZero(secKey, result)
	//as per our convention, if R > 0, then the corresponding bit is a 1, else it is a 0
	return output != int64(0)
}

// ExecuteBitAnd computes the result to encode, under the AND operation.
func ExecuteBitAnd(input, wantProof bool) int64 {
	return ExecuteBitOr(!input, wantProof)
}

//EncodeBitAND computes the encoding of bit Xi, under the AND operation
func EncodeBitAND(input bool, pubKey kyber.Point) (*libunlynx.CipherText, int64) {
	cipher, clear, _ := EncodeBitANDWithProof(input, pubKey, nil, 0, 0)
	return cipher, clear
}

//EncodeBitANDWithProof computes the encoding of bit Xi, under the AND operation with range proofs
func EncodeBitANDWithProof(input bool, pubKey kyber.Point, sigs []libdrynx.PublishSignature, l int64, u int64) (*libunlynx.CipherText, int64, libdrynxrange.CreateProof) {
	toEncrypt := ExecuteBitAnd(input, sigs != nil)
	cipher, r := libunlynx.EncryptIntGetR(pubKey, toEncrypt)

	cp := libdrynxrange.CreateProof{}
	if sigs != nil {
		cp = libdrynxrange.CreateProof{Sigs: sigs, U: u, L: l, Secret: toEncrypt, R: r, CaPub: pubKey, Cipher: *cipher}

	}

	return cipher, toEncrypt, cp
}

//DecodeBitAND computes the decoding of bit Xi, under the AND operation
func DecodeBitAND(result libunlynx.CipherText, secKey kyber.Scalar) bool {
	//decrypt the bit representation
	output := libunlynx.DecryptCheckZero(secKey, result)
	//as per our convention, if R > 0, then the corresponding bit is a 1, else it is a 0
	return output == int64(0)
}

//LocalResultOR calculates the local result of the OR operation over all boolean values of the input array
func LocalResultOR(input []bool) bool {
	for i := int64(0); i < int64(len(input)); i++ {
		if input[i] {
			return true
		}
	}
	return false
}

//LocalResultAND calculates the local result of the AND operation over all boolean values of the input array
func LocalResultAND(input []bool) bool {
	for i := int64(0); i < int64(len(input)); i++ {
		if !input[i] {
			return false
		}
	}
	return true
}
