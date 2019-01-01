package delay

import (
	"context"

	pb "libs.altipla.consulting/delay/queues"
	"libs.altipla.consulting/services"
)

// Conn represents a connection to the queues server that can be implemented
// in multiple ways with an in-memory mock or a real connection to the remote
// queues server.
type Conn interface {
	SendTasks(ctx context.Context, name string, tasks []*pb.SendTask) error
	Listen(ctx context.Context, name string) (ConnListener, error)
	Project() string
}

type ConnListener interface {
	Next() ([]*pb.Task, error)
	ACK(task *pb.Task, success bool) error
}

// NewConn opens a new connection to a queues server. It needs the project and the OAuth
// client credentials to authenticate the requests.
//
// In the local dev environment it will use a Redis fallback that requires both the
// sender & receiver to be connected at the same time simulating the remote queues.
func NewConn(project, clientID, clientSecret string) (Conn, error) {
	if services.IsLocal() {
		return newLocalConn(project)
	}
	return newProductionConn(project, clientID, clientSecret)
}

// QueueSpec contains a reference to a queue to send to and receive tasks from that queue.
type QueueSpec struct {
	conn Conn
	name string
}

// Queue builds a new QueueSpec reference to a queue.
func Queue(conn Conn, name string) QueueSpec {
	return QueueSpec{conn, name}
}

// SendTasks sends a list of tasks in batch to the queue.
func (queue QueueSpec) SendTasks(ctx context.Context, tasks []*pb.SendTask) error {
	return queue.conn.SendTasks(ctx, queue.name, tasks)
}

// SendTask sends a tasks to the queue.
func (queue QueueSpec) SendTask(ctx context.Context, task *pb.SendTask) error {
	return queue.SendTasks(ctx, []*pb.SendTask{task})
}
