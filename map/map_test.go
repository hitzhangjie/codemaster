package map_test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_accessNilMap(t *testing.T) {
	var m map[int]bool

	// read nil map: won't panic
	require.NotPanics(t, func() {
		v, ok := m[1]
		_ = v
		_ = ok
	})

	// write nil map: will panic
	require.Panics(t, func() {
		m[1] = true
	})
}
