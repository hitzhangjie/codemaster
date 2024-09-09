package file

import (
	"os"
	"syscall"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	SYS_KCMP  = 312
	KCMP_FILE = 0
)

func kcmp_files(pid1, pid2, fd1, fd2 int) (int, error) {
	r1, _, err := syscall.Syscall6(SYS_KCMP, uintptr(pid1), uintptr(pid2), KCMP_FILE, uintptr(fd1), uintptr(fd2), 0)

	return int(r1), err
}

// 比较不同进程打开的进程fd是否是同一个
func Test_kcmp_files(t *testing.T) {
	pid := os.Getpid()
	f1, err := os.OpenFile("./test", os.O_RDONLY|os.O_CREATE, 0644)
	require.Nil(t, err)

	f2, err := os.OpenFile("./test", os.O_RDONLY|os.O_CREATE, 0644)
	require.Nil(t, err)

	r1, err := kcmp_files(pid, pid, int(f1.Fd()), int(f2.Fd()))
	require.Equal(t, "errno 0", err.Error())
	require.NotZero(t, r1)
}
