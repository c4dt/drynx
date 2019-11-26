package loaders

import (
	"errors"
	"math/rand"

	"github.com/ldsec/drynx/lib"
	"github.com/ldsec/drynx/lib/provider"
)

type random struct {
	min, max float64
	rows     uint
}

// NewRandom create a Loader of random values.
func NewRandom(min, max float64, rows uint) (provider.Loader, error) {
	if min > max {
		return nil, errors.New("minimum > maximum")
	}
	return random{min, max, rows}, nil
}

func (r random) Provide(query libdrynx.Query) ([][]float64, error) {
	ret := make([][]float64, len(query.Selector))

	for i := range ret {
		arr := make([]float64, r.rows)
		for j := range arr {
			arr[j] = r.min + rand.Float64()*(r.max-r.min)
		}
		ret[i] = arr
	}
	return ret, nil
}
