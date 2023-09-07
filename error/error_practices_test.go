package error

import (
	stderrors "errors"
	"fmt"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func Test_pkgerrs_withstack(t *testing.T) {
	do := func() error {
		return stderrors.New("this is an error")
	}
	fn := func() error {
		err := func() error {
			err := func() error {
				if err := do(); err != nil {
					return errors.WithStack(err)
				}
				return nil
			}()
			return errors.Wrap(err, "x")
		}()
		return errors.Wrap(err, "y")
	}

	err := fn()
	require.Error(t, err)
	fmt.Println("by Println:", err)
	fmt.Printf("by Printf(\"%%v\"): %v\n", err)
	fmt.Printf("by Printf(\"%%+v\"): %+v\n", err)

}
