package conv

import (
	"go.dedis.ch/onet/v3"

	"github.com/ldsec/drynx/lib"
)

type QueryMarshallable struct {
	Operation2 OperationMarshallable

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

func QueryFromMarshallable(marshallable QueryMarshallable) (libdrynx.Query, error) {
	ivSigs := marshallable.IVSigs
	ivSigs.InputValidationSigs = recreateRangeSignatures(marshallable.IVSigs)

	op, err := OperationFromMarshallable(marshallable.Operation2)

	return libdrynx.Query{
		Operation2: op,

		Operation:     marshallable.Operation,
		Ranges:        marshallable.Ranges,
		Proofs:        marshallable.Proofs,
		Obfuscation:   marshallable.Obfuscation,
		DiffP:         marshallable.DiffP,
		IVSigs:        ivSigs,
		RosterVNs:     marshallable.RosterVNs,
		CuttingFactor: marshallable.CuttingFactor,
		Selector:      marshallable.Selector,
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

func ToQueryMarshallable(q libdrynx.Query) (QueryMarshallable, error) {
	op, err := ToOperationMarshallable(q.Operation2)
	return QueryMarshallable{
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
