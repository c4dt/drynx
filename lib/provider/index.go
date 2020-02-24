package provider

import "github.com/ldsec/drynx/lib"

// Loader is the way to retrieve local data.
type Loader interface {
	// Provide returns the queried rows to encode.
	// Returns a matrix of len Query.Operation.NbrInput
	Provide(libdrynx.Query) ([][]float64, error)
}

// Neutralizer decides to release or not the results of a query.
type Neutralizer interface {
	// Vet checks if the results can be safely released.
	Vet(libdrynx.Query, [][]float64) bool
}
