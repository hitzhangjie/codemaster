package mathtest

import (
	"math"
	"sort"
	"testing"
)

func TestXXXX(t *testing.T) {
	v := []float64{1, 2, 3, 4, math.Inf(+1)}
	sort.Float64s(v)
	println(sort.SearchFloat64s(v, 0.5))
	println(sort.SearchFloat64s(v, 0))
}
