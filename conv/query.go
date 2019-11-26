package conv

import (
	"go.dedis.ch/onet/v3"

	"github.com/ldsec/drynx/lib"
)

// Query is a serializable libdrynx.Query.
type Query struct {
	Operation2 Operation

	// from old Query

	// query statement
	Operation   libdrynx.Operation
	Ranges      []*[]int64
	Proofs      int
	Obfuscation bool
	DiffP       libdrynx.QueryDiffP

	// identity skipchain simulation
	IVSigs    libdrynx.QueryIVSigs
	RosterVNs *onet.Roster

	//simulation
	CuttingFactor int

	// allow to select which column to compute operation on
	Selector []libdrynx.ColumnID
}

// ToQuery deserialize.
func (q Query) ToQuery() (libdrynx.Query, error) {
	ivSigs := q.IVSigs
	ivSigs.InputValidationSigs = recreateRangeSignatures(q.IVSigs)

	op, err := q.Operation2.ToOperation()

	return libdrynx.Query{
		Operation2: op,

		Operation:     q.Operation,
		Ranges:        q.Ranges,
		Proofs:        q.Proofs,
		Obfuscation:   q.Obfuscation,
		DiffP:         q.DiffP,
		IVSigs:        ivSigs,
		RosterVNs:     q.RosterVNs,
		CuttingFactor: q.CuttingFactor,
		Selector:      q.Selector,
	}, err
}

func recreateRangeSignatures(ivSigs libdrynx.QueryIVSigs) []*[]libdrynx.PublishSignatureBytes {
	recreate := make([]*[]libdrynx.PublishSignatureBytes, 0)

	// transform the one-dimensional array (because of protobuf) to the original two-dimensional array
	indexInit := 0
	for i := 1; i <= len(ivSigs.InputValidationSigs); i++ {
		if i%ivSigs.InputValidationSize2 == 0 {
			tmp := make([]libdrynx.PublishSignatureBytes, ivSigs.InputValidationSize2)
			for j := range tmp {
				tmp[j] = (*ivSigs.InputValidationSigs[indexInit])[0]
				indexInit++
			}
			recreate = append(recreate, &tmp)

			indexInit = i
		}

	}
	return recreate
}

// FromQuery serialize.
func FromQuery(q libdrynx.Query) (Query, error) {
	op, err := FromOperation(q.Operation2)
	return Query{
		Operation2: op,

		Operation:     q.Operation,
		Ranges:        q.Ranges,
		Proofs:        q.Proofs,
		Obfuscation:   q.Obfuscation,
		DiffP:         q.DiffP,
		IVSigs:        q.IVSigs,
		RosterVNs:     q.RosterVNs,
		CuttingFactor: q.CuttingFactor,
		Selector:      q.Selector,
	}, err
}
