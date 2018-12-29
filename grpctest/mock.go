package grpctest

import (
	"context"
	"fmt"
	"net"
	"path/filepath"
	"reflect"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

// ClientMock contains all the info needed to mock client calls.
type ClientMock struct {
	methods map[string]reflect.Value
}

// Mock creates a new mock client.
func Mock() *ClientMock {
	return &ClientMock{
		methods: make(map[string]reflect.Value),
	}
}

// Set assigns a method to a handler function that will be called in the mock. The name
// should be only the method name, not the service nor the package.
func (mock *ClientMock) Set(name string, handler interface{}) {
	ft := reflect.TypeOf(handler)
	fv := reflect.ValueOf(handler)

	if ft.NumOut() != 2 {
		panic("mocks must return the reply type and the error")
	}
	if ft.NumIn() != 1 {
		panic("mocks must receive only the input type")
	}

	mock.methods[name] = fv
}

// Conn returns a GRPC connection that uses the mock client when connecting.
func (mock *ClientMock) Conn() *grpc.ClientConn {
	listener := bufconn.Listen(0)

	dialer := func(hostname string, timeout time.Duration) (net.Conn, error) {
		return listener.Dial()
	}
	interceptor := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		name := filepath.Base(method)
		fv, ok := mock.methods[name]
		if !ok {
			return fmt.Errorf("MOCK method not found: %s", method)
		}

		rets := fv.Call([]reflect.Value{reflect.ValueOf(req)})
		if !rets[1].IsNil() {
			return rets[1].Interface().(error)
		}

		replyv := reflect.ValueOf(reply).Elem()
		replyv.Set(rets[0].Elem())

		return nil
	}
	conn, err := grpc.Dial("", grpc.WithUnaryInterceptor(interceptor), grpc.WithDialer(dialer), grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	return conn
}
