package syscall

import (
	"syscall"
	"testing"
)

func TestSyscall(t *testing.T) {
	var status syscall.WaitStatus

	// fork and single step
	signo := status.StopSignal()

	// go1.14 non-cooperative preemption, choose SIGURG as the signal to notify
	// the thread to stop goroutine.
	// https://github.com/golang/go/issues/38290#issuecomment-610551277
	//
	// So, SIGURG should be ignored here.
	_ = signo
}
