package delay

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"reflect"
	"strings"

	pb "libs.altipla.consulting/delay/queues"
)

type TestingConn struct {
	Queues map[string][]*pb.SendTask
}

func NewTestingConn() *TestingConn {
	return &TestingConn{
		Queues: make(map[string][]*pb.SendTask),
	}
}

func (conn *TestingConn) SendTasks(ctx context.Context, name string, tasks []*pb.SendTask) error {
	conn.Queues[name] = append(conn.Queues[name], tasks...)
	return nil
}

func (conn *TestingConn) Listen(ctx context.Context, name string) (ConnListener, error) {
	return nil, fmt.Errorf("delay: testing connection does not implement listen")
}

func (conn *TestingConn) Project() string {
	return "foo-project"
}

func (conn *TestingConn) Receive(name, expectedTask string, f interface{}) {
	fv := reflect.ValueOf(f)

	task := conn.Queues[name][0]
	if len(conn.Queues[name]) > 1 {
		conn.Queues[name] = conn.Queues[name][1:]
	} else {
		conn.Queues[name] = nil
	}

	r := bytes.NewReader(task.Payload)
	var inv invocation
	if err := gob.NewDecoder(r).Decode(&inv); err != nil {
		panic(err)
	}

	parts := strings.Split(inv.Key, ":")
	if parts[1] != expectedTask {
		panic(fmt.Sprintf("expected task for %s, got %s", expectedTask, parts[1]))
	}

	ft := fv.Type()
	var in []reflect.Value
	for _, arg := range inv.Args {
		var v reflect.Value
		if arg != nil {
			v = reflect.ValueOf(arg)
		} else {
			// Task was passed a nil argument, so we must construct
			// the zero value for the argument here.
			n := len(in) // we're constructing the nth argument
			var at reflect.Type
			if !ft.IsVariadic() || n < ft.NumIn()-1 {
				at = ft.In(n)
			} else {
				at = ft.In(ft.NumIn() - 1).Elem()
			}
			v = reflect.Zero(at)
		}
		in = append(in, v)
	}
	fv.Call(in)
}
