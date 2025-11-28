// Package errs 自定义了一个error实现，它可以同时指定错误码、错误描述。
//
// 建立这个errs包的背景如下：
//
//  1. 接口处理函数，只能返回req对应的rsp，逻辑错误码只能通过rsp.rspult返回，
//     框架层错误才会通过包头字段返回。
//
//  2. 作为CS通信协议、SS通信协议，我们需要返回错误码给到对端，为了区分不同类型错误，
//     项目中需要维护大量错误码及字符串描述，如果我们通过`var err error`作为判断
//     条件然后`if err != nil { rsp.rspult = ErrCode_xxxxx }`，为了能正常
//     拿到这个错误码（包括通过自定义的函数、方法），会有些很差劲的写法：
//     - 比如把rsp作为入参（只是会修改错误码吗？）
//     - 比如把int32错误码、、error同时作为出参（违背go实践习惯，
//     go返回err!=nil时其他值为0值，若返回nil但是实际发生逻辑错误，也会引入特殊判定逻辑）
//
//  3. 有的逻辑错误码，可以代指一类错误，比如请求缺少参数1\缺少参数2等，可以统一
//     返回“参数无效”给请求方，但是在服务器侧要打印错误日志时错误描述应该能区别到底缺少
//     什么，不能把所有参数在日志全打出来，这会让定位问题变得耗时。
//
// 综上，设计了这个errs包，它可以同时指定错误码、错误描述，在自定义函数、方法中可以直接
// 返回该errs.Error代替error，不管是打印错误日志，还是将对应的逻辑错返回给请求方，
// rsp.rspult = err.Code()即可，可以缓解上面 2）中出现的可维护性差的写法。
// 错误码也可以比较容易复用，打日志时携带上下文信息也会更方便，还不用打log时代码行冗长。
//
// 也提供了包级别的errs.Code(error)\errs.Message(error)方法，这意味着也可以统一返回
// error类型，而非*errs.Error，这样更符合go惯例。
//
// btw, 更好的做法是不只业务逻辑层面直接返回该*errs.Error类型,框架层面也做类似
// 支持,但是要区分是框架错误码还是逻辑错误码. 这样整体的处理风格就一致了.
//
// ----------------------------------------------------------------------------
//
// ps: grpc里面使用的是Status,还允许返回错误时返回其他信息,就好比http
// statuscode!=0时,还会通过body返回其他信息.以前我们实现trpc时返回错误码情况下
// 是不返回其他信息的,但是有业务明确表达了http的做法\提出了相关的诉求. 当然这
// 是另一个特性了.
package errs

import (
	"errors"
	"fmt"
)

// Error 自定义error实现，可以保留错误码、错误描述信息
type Error struct {
	code int32
	msg  string
}

// New 创建一个错误实例
func New(code int32, msg string) *Error {
	return &Error{code, msg}
}

// Errorf 提供接口，支持大家可以直接按照格式化参数填入，使用方式类似于：errs.Errorf(ErrCode_xxx, "fail:%v", xxx)
func Errorf(code int32, format string, args ...any) *Error {
	return &Error{code, fmt.Sprintf(format, args...)}
}

// Error 返回错误描述信息
func (e *Error) Error() string {
	if e == nil {
		return "nil"
	}
	return fmt.Sprintf("code: %d, msg: %s", e.code, e.msg)
}

// Code 返回错误码
func (e *Error) Code() int32 {
	if e == nil {
		return 0
	}
	return e.code
}

// Msg 返回错误描述
func (e *Error) Msg() string {
	if e == nil {
		return ""
	}
	return e.msg
}

// String 返回error描述信息
func (e *Error) String() string {
	if e == nil {
		return "nil error"
	}
	return fmt.Sprintf("errcode: %d, errmsg: %s", e.code, e.msg)
}

var errUnknown = New(88888888, "unknown error")

// Code 返回错误的错误码，如果不是*Error类型，则返回unknownError
func Code(err error) int32 {
	if err == nil {
		return 0
	}
	var e *Error
	if errors.As(err, &e) {
		return e.Code()
	}
	return errUnknown.Code()
}

// Message 返回错误描述信息
func Message(err error) string {
	if err == nil {
		return ""
	}
	var e *Error
	if errors.As(err, &e) {
		return e.Msg()
	}
	return errUnknown.Msg()
}
