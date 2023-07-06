package cgo_test

import (
	"testing"
	"time"

	"github.com/hitzhangjie/codemaster/cgo/rand"
)

func Test_cgo_random(t *testing.T) {
	rand.Seed(int(time.Now().Unix()))
	println(rand.Random())
}
