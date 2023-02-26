package ints

import (
	"fmt"
	"testing"

	"golang.org/x/net/context"
)

type Handler func(ctx context.Context, req interface{}) (rsp interface{}, err error)
type Interceptor func(ctx context.Context, req interface{}, handler Handler) (rsp interface{}, err error)

func getChainInterceptorHandler(ctx context.Context, ints []Interceptor, idx int, h Handler) Handler {
	if idx == len(ints)-1 {
		return h
	}
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		return ints[idx+1](ctx, req, getChainInterceptorHandler(ctx, ints, idx+1, h))
	}
}

func processHandler(ctx context.Context, req interface{}) (interface{}, error) {
	fmt.Println("process request")
	return nil, nil
}

func int1(ctx context.Context, req interface{}, handler Handler) (interface{}, error) {
	fmt.Println("int1 begin")
	handler(ctx, req)
	fmt.Println("int1 end")
	return nil, nil
}

func int2(ctx context.Context, req interface{}, handler Handler) (interface{}, error) {
	fmt.Println("int2 begin")
	handler(ctx, req)
	fmt.Println("int2 end")
	return nil, nil
}

func int3(ctx context.Context, req interface{}, handler Handler) (interface{}, error) {
	fmt.Println("int3 begin")
	handler(ctx, req)
	fmt.Println("int3 end")
	return nil, nil
}

func TestInterceptors(t *testing.T) {

	interceptors := []Interceptor{int1, int2, int3}

	first := func(ctx context.Context, req interface{}, h Handler) (rsp interface{}, err error) {
		return interceptors[0](ctx, req, h)
	}

	ctx := context.TODO()
	req := struct{}{}
	first(ctx, req, getChainInterceptorHandler(ctx, interceptors, 0, processHandler))
}
