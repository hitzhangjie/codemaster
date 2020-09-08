package timezone

import (
	"fmt"
	"testing"
	"time"
)

func TestUnixtime(t *testing.T) {
	secsSinceEpoc := time.Now().Unix()

	t1 := time.Unix(secsSinceEpoc, 0).UTC()
	fmt.Println(t1)

	t2 := time.Unix(secsSinceEpoc, 0).In(time.Local)
	fmt.Println(t2)
}
