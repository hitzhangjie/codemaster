package errs_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hitzhangjie/codemaster/errs"
)

func Test_UnpackNilError(t *testing.T) {
	t.Run("return value type == error", func(t *testing.T) {
		t.Run("nil error", func(t *testing.T) {
			fn := func() error {
				var err *errs.Error
				return err
			}

			err := fn()
			//warn: Nil/NotNil will check the underlying value stored in err.
			//      Hence, it will know the underlying `*errs.Error == nil` is true.
			//      Here we could use require.Error instead.
			//require.NotNil(t, err)
			require.Error(t, err)

			require.Equal(t, errs.Code(err), int32(0))
			require.Equal(t, errs.Message(err), "")
		})

		t.Run("not nil error", func(t *testing.T) {
			fn := func() error {
				err := errs.New(1111, "xxxx")
				return err
			}

			err := fn()
			require.Error(t, err)
			require.Equal(t, errs.Code(err), int32(1111))
			require.Equal(t, errs.Message(err), "xxxx")
		})

	})

	t.Run("return value type == *errs.Error", func(t *testing.T) {
		t.Run("nil error", func(t *testing.T) {
			fn := func() (err *errs.Error) {
				return
			}

			err := fn()
			require.Nil(t, err)
			require.Equal(t, errs.Code(err), int32(0))
			require.Equal(t, errs.Message(err), "")
		})

		t.Run("not nil", func(t *testing.T) {
			fn := func() *errs.Error {
				return errs.New(1111, "xxxx")
			}

			err := fn()
			require.Error(t, err)
			require.Equal(t, errs.Code(err), err.Code())
			require.Equal(t, errs.Message(err), err.Msg())
		})

	})
}

func Test_PackedError(t *testing.T) {
	var err = errs.New(1111, "xxxx")
	require.Equal(t, int32(1111), errs.Code(err))

	var err2 = fmt.Errorf("%w yyyy", err)
	require.Equal(t, int32(1111), errs.Code(err2))
}
