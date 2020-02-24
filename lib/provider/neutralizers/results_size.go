package neutralizers

import (
	"github.com/ldsec/drynx/lib"
	"github.com/ldsec/drynx/lib/provider"
)

type minimumResultsSize struct {
	minimum uint
}

// NewMinimumResultsSize creates a Neutralizer vetting only when len(results) >= minimum
func NewMinimumResultsSize(minimum uint) provider.Neutralizer {
	return minimumResultsSize{minimum}
}

func (rs minimumResultsSize) Vet(_ libdrynx.Query, results [][]float64) bool {
	return uint(len(results)) >= rs.minimum
}
