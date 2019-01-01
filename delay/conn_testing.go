package delay

import (
	"context"
	"fmt"

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
