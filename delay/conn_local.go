package delay

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/go-redis/redis"
	"github.com/golang/protobuf/proto"

	"libs.altipla.consulting/datetime"
	pb "libs.altipla.consulting/delay/queues"
)

type localConn struct {
	project     string
	redisClient *redis.Client
}

func newLocalConn(project string) (Conn, error) {
	return &localConn{
		project:     project,
		redisClient: redis.NewClient(&redis.Options{Addr: "redis:6379"}),
	}, nil
}

func (conn *localConn) SendTasks(ctx context.Context, name string, tasks []*pb.SendTask) error {
	var buf proto.Buffer
	for _, task := range tasks {
		if err := buf.EncodeMessage(task); err != nil {
			return fmt.Errorf("delay: cannot encode task: %v", err)
		}
	}
	if err := conn.redisClient.Publish(name, buf.Bytes()).Err(); err != nil {
		return fmt.Errorf("delay: cannot send tasks: %v", err)
	}

	return nil
}

func (conn *localConn) Listen(ctx context.Context, name string) (ConnListener, error) {
	pubsub := conn.redisClient.Subscribe(name)

	return &localListener{
		pubsub:  pubsub,
		ch:      pubsub.Channel(),
		project: conn.project,
	}, nil
}

func (conn *localConn) Project() string {
	return conn.project
}

type localListener struct {
	pubsub  *redis.PubSub
	ch      <-chan *redis.Message
	i       int
	project string
}

func (lis *localListener) Next() ([]*pb.Task, error) {
	msg := <-lis.ch

	var tasks []*pb.Task
	buf := proto.NewBuffer([]byte(msg.Payload))
	for {
		sendTask := new(pb.SendTask)
		if err := buf.DecodeMessage(sendTask); err != nil {
			if err == io.EOF {
				break
			}

			return nil, fmt.Errorf("delay: cannot decode incoming task: %v", err)
		}

		lis.i++
		tasks = append(tasks, &pb.Task{
			Code:    fmt.Sprintf("sim-%d", lis.i),
			Payload: sendTask.Payload,
			Created: datetime.SerializeTimestamp(time.Now()),
			Retry:   0,
			Project: lis.project,
			MinEta:  sendTask.MinEta,
		})
	}

	return tasks, nil
}

func (lis *localListener) ACK(task *pb.Task, success bool) error {
	return nil
}
