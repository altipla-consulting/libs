package delay

import (
	"context"
	"fmt"

	"libs.altipla.consulting/connect"
	pb "libs.altipla.consulting/delay/queues"
)

type productionConn struct {
	project      string
	queuesClient pb.QueuesServiceClient
}

func newProductionConn(project, clientID, clientSecret string) (Conn, error) {
	conn, err := connect.OAuthToken("api-v3.altipla.consulting", clientID, clientSecret)
	if err != nil {
		return nil, fmt.Errorf("delay: cannot connect to altipla api: %v", err)
	}

	return &productionConn{
		project:      project,
		queuesClient: pb.NewQueuesServiceClient(conn),
	}, nil
}

func (conn *productionConn) SendTasks(ctx context.Context, name string, tasks []*pb.SendTask) error {
	req := &pb.SendTasksRequest{
		Project:   conn.project,
		QueueName: name,
		Tasks:     tasks,
	}
	if _, err := conn.queuesClient.SendTasks(ctx, req); err != nil {
		return fmt.Errorf("delay: cannot send tasks: %v", err)
	}

	return nil
}

func (conn *productionConn) Listen(ctx context.Context, name string) (ConnListener, error) {
	stream, err := conn.queuesClient.Listen(ctx)
	if err != nil {
		return nil, fmt.Errorf("delay: cannot listen to the queue: %v", err)
	}

	initial := &pb.ListenRequest{
		Request: &pb.ListenRequest_Initial{
			Initial: &pb.ListenInitial{
				Project:   conn.project,
				QueueName: name,
			},
		},
	}
	if err := stream.Send(initial); err != nil {
		return nil, fmt.Errorf("delay: cannot send initial connection info: %v", err)
	}

	return &productionListener{
		stream: stream,
	}, nil
}

func (conn *productionConn) Project() string {
	return conn.project
}

type productionListener struct {
	stream pb.QueuesService_ListenClient
}

func (lis *productionListener) Next() ([]*pb.Task, error) {
	reply, err := lis.stream.Recv()
	if err != nil {
		return nil, fmt.Errorf("delay: cannot receive tasks: %v", err)
	}

	return []*pb.Task{reply.Task}, nil
}

func (lis *productionListener) ACK(task *pb.Task, success bool) error {
	req := &pb.ListenRequest{
		Request: &pb.ListenRequest_Ack{
			Ack: &pb.Ack{
				Code:    task.Code,
				Success: success,
			},
		},
	}
	if err := lis.stream.Send(req); err != nil {
		return fmt.Errorf("delay: cannot ack task: %v", err)
	}

	return nil
}
