package tcp_test

import (
	"testing"
)

// 细分为如下cases：
// - 写之前，连接就被关闭了，会直接返回 n==0 && err=use of closed connection
// - 部分成功，写的过程中，连接被关闭了，会返回 n!=0 && err=use of closed connection，这种情况下对使用方而言还是失败的，
// - 写完后，连接被关闭了，这种就不讨论了
//
// ---------------------------------------------------------------------------
//
// ps: 关于部分成功这个问题，我今天重新梳理了一下 syscall interrupted 并且 syscall restarted的过程：
//
// 举个golang标准库中tcpconn.Write(data)的实现：
//
//	func (fd *FD) Write(p []byte) (int, error) {
//	    if err := fd.writeLock(); err != nil {
//	        return 0, err
//	    }
//	    defer fd.writeUnlock()
//	    if err := fd.pd.prepareWrite(fd.isFile); err != nil {
//	        return 0, err
//	    }
//	    var nn int
//	    for {
//	        max := len(p)
//	        if fd.IsStream && max-nn > maxRW {
//	            max = nn + maxRW
//	        }
//	        n, err := ignoringEINTRIO(syscall.Write, fd.Sysfd, p[nn:max])
//	        if n > 0 {
//	            nn += n
//	        }
//	        if nn == len(p) {
//	            return nn, err
//	        }
//	        if err == syscall.EAGAIN && fd.pd.pollable() {
//	            if err = fd.pd.waitWrite(fd.isFile); err == nil {
//	                continue
//	            }
//	        }
//	        if err != nil {
//	            return nn, err
//	        }
//	        if n == 0 {
//	            return nn, io.ErrUnexpectedEOF
//	        }
//	    }
//	}
//
//	func ignoringEINTRIO(fn func(fd int, p []byte) (int, error), fd int, p []byte) (int, error) {
//	    for {
//	        n, err := fn(fd, p)
//	        if err != syscall.EINTR {
//	            return n, err
//	        }
//	    }
//	}
//
// 有个疑问如果一次syscall_write遇到了EINTR错误，难道下次重试时不会重复发送数据吗？好问题！
// ---------------------------------------------------------------------------
// 让我来解释一下 syscall_write 系统调用被中断时的处理机制，以及为什么重启不会导致重复写入的问题。
// 当 syscall_write 系统调用被中断时，设置 ret=-1 和 errno=EINTR 的过程通常如下：
//
// a) 中断发生： 当 syscall_write 执行过程中发生中断（例如信号中断），内核会暂停当前的系统调用处理。
// b) 中断处理： 内核处理完中断后，会检查是否需要提前结束被中断的系统调用。
// c) 设置返回值： 如果决定提前结束系统调用，内核会设置系统调用的返回值。在这种情况下：
// - 将返回值 (ret) 设置为 -1
// - 将 errno 设置为 EINTR（通常定义为 4，表示被中断的系统调用）
// d) 返回用户空间： 系统调用结束，控制权返回到用户空间，应用程序可以检查返回值和 errno 来确定系统调用是否被中断。
//
// 避免重复写入
// syscall_write 重启时不会出现重复写入的原因主要有以下几点：
// a) 部分写入的处理： 操作系统通常会记录已经写入的字节数。如果 write 调用被中断，它会返回已经成功写入的字节数。
// b) 文件偏移量的更新： 每次成功写入后，文件的偏移量会相应地更新。这意味着即使系统调用被中断，下次写入时也会从正确的位置开始。
// c) 应用程序的责任： 良好设计的应用程序在处理 EINTR 错误时，应该检查已写入的字节数，并相应地调整剩余需要写入的数据。
func Test_WriteClosedTcpConn(t *testing.T) {
}
