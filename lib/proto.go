package libdrynx

import (
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/onet/v3"
	"go.dedis.ch/onet/v3/network"
)

// PROTOSTART
// package libdrynx;
// type :skipchain.SkipBlockID:bytes
// type :darc.ID:bytes
// type :darc.Action:string
// type :Arguments:[]Argument
// type :Instructions:[]Instruction
// type :TxResults:[]TxResult
// type :InstanceID:bytes
// type :ColumnID:string
// type :Version:sint32
// type :Operation2:bytes
// type :map\[string\]\*CipherVector:map<string, CipherVector>
// type :map\[string\]kyber.Point:map<string, bytes>
// type :map\[string\]\*ServerIdentityList:map<string, ServerIdentityList>

// only for dedis/protobuf
type CipherVector struct {
	// optional
	Content []*CipherText
}

// only for dedis/protobuf
type CipherText struct {
	// optional
	K kyber.Point
	// optional
	C kyber.Point
}

// ServerIdentityList exists solely because dedis/protobuf doesn't support map[ServerIdentity][]ServerIdentity
type ServerIdentityList struct {
	// optional
	Content []network.ServerIdentity
}

// SurveyQuery is a serializable libdrynx.SurveyQuery.
type SurveyQuery struct {
	// optional
	SurveyID string
	// optional
	Query Query

	// optional
	RosterServers onet.Roster
	// optional
	ServerToDP map[string]*ServerIdentityList
	// optional
	IDtoPublic map[string]kyber.Point

	// optional
	ClientPubKey kyber.Point
	// optional
	IntraMessage bool
	// to define whether the query was sent by the querier or not

	//Threshold for verification in skipChain service
	// optional
	Threshold float64
	// optional
	AggregationProofThreshold float64
	// optional
	ObfuscationProofThreshold float64
	// optional
	RangeProofThreshold float64
	// optional
	KeySwitchingProofThreshold float64
}

// ResponseDP contains the data provider's response to be sent to the server.
type ResponseDP struct {
	// group -> value(s)
	// optional
	Data map[string]*CipherVector
}

//PublishSignatureBytes is the same as PublishSignature but the signatures are in bytes
//need this because of G2 in protobuf not working
type PublishSignatureBytes struct {
	// y
	// optional
	Public kyber.Point
	// A_i
	// optional
	Signature []byte
}

// QueryDiffP contains diffP parameters for a query
type QueryDiffP struct {
	// optional
	LapMean float64
	// optional
	LapScale float64
	// optional
	NoiseListSize int
	// optional
	Quanta float64
	// optional
	Scale float64
	// optional
	Limit float64
}

type PublishSignatureBytesList struct {
	// optional
	Content []PublishSignatureBytes
}

// QueryIVSigs contains parameters for input validation
type QueryIVSigs struct {
	// optional
	InputValidationSigs []*PublishSignatureBytesList
}

type Int64List struct {
	// optional
	Content []int64
}

// Query is used to transport query information through servers, to DPs
type Query struct {
	// query statement
	// optional
	Operation Operation
	// optional
	Ranges []*Int64List
	// optional
	Proofs int
	// optional
	Obfuscation bool
	// optional
	DiffP QueryDiffP

	// identity skipchain simulation
	// optional
	IVSigs QueryIVSigs
	// optional
	RosterVNs *onet.Roster

	//simulation
	// optional
	CuttingFactor int

	// allow to select which column to compute operation on
	// optional
	Selector []ColumnID
}

// Operation defines the operation in the query
type Operation struct {
	// optional
	NameOp string
	// optional
	NbrInput int
	// optional
	NbrOutput int
	// optional
	QueryMin int64
	// optional
	QueryMax int64
	// optional
	LRParameters LogisticRegressionParameters
}

// LogisticRegressionParameters are the parameters specific to logistic regression
type LogisticRegressionParameters struct {
	// logistic regression specific
	// optional
	DatasetName string
	// optional
	FilePath string
	// optional
	NbrRecords int64
	// optional
	NbrFeatures int64
	// optional
	Means []float64
	// optional
	StandardDeviations []float64

	// parameters
	// optional
	Lambda float64
	// optional
	Step float64
	// optional
	MaxIterations int
	// optional
	InitialWeights []float64

	// approximation
	// optional
	K int
	// optional
	PrecisionApproxCoefficients float64
}
