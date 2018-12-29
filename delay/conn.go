package delay

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/go-redis/redis"
	"github.com/golang/protobuf/proto"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	pb "libs.altipla.consulting/delay/queues"
)

const beauthTokenEndpoint = "https://beauth.io/token"

// Conn represents a connection to the queues server.
type Conn struct {
	project      string
	queuesClient pb.QueuesServiceClient
	redisClient  *redis.Client
}

// NewConn opens a new connection to a queues server. It needs the project and the OAuth
// client credentials to authenticate the requests.
func NewConn(project, clientID, clientSecret string) (*Conn, error) {
	config := &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     beauthTokenEndpoint,
	}
	rpcCreds := grpc.WithPerRPCCredentials(oauthAccess{config.TokenSource(context.Background())})
	creds := credentials.NewTLS(&tls.Config{ServerName: "api-v3.altipla.consulting"})
	conn, err := grpc.Dial("api-v3.altipla.consulting:443", grpc.WithTransportCredentials(creds), rpcCreds)
	if err != nil {
		return nil, fmt.Errorf("delay: cannot connect to altipla api: %v", err)
	}

	return &Conn{
		project:      project,
		queuesClient: pb.NewQueuesServiceClient(conn),
	}, nil
}

// NewDebugConn creates a new local debugging connection that uses a direct Redis
// queue to simulate the queue. The downside is both the sender and receiver should
// be connected at the same time to send the message; there is no storage.
func NewDebugConn() (*Conn, error) {
	return &Conn{
		redisClient: redis.NewClient(&redis.Options{Addr: "redis:6379"}),
	}, nil
}

type oauthAccess struct {
	tokenSource oauth2.TokenSource
}

func (oa oauthAccess) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	token, err := oa.tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("delay: cannot update token: %v", err)
	}

	return map[string]string{
		"authorization": token.Type() + " " + token.AccessToken,
	}, nil
}

func (oa oauthAccess) RequireTransportSecurity() bool {
	return false
}

// QueueSpec contains a reference to a queue to send to and receive tasks from that queue.
type QueueSpec struct {
	conn *Conn
	name string
}

// Queue builds a new QueueSpec reference to a queue.
func Queue(conn *Conn, name string) QueueSpec {
	return QueueSpec{conn, name}
}

// SendTasks sends a list of tasks in batch to a queue.
func (queue QueueSpec) SendTasks(ctx context.Context, tasks []*pb.SendTask) error {
	if queue.conn.redisClient != nil {
		var buf proto.Buffer
		for _, task := range tasks {
			if err := buf.EncodeMessage(task); err != nil {
				return fmt.Errorf("delay: cannot encode task: %v", err)
			}
		}
		if err := queue.conn.redisClient.Publish(queue.name, buf.Bytes()).Err(); err != nil {
			return fmt.Errorf("delay: cannot send to the debug queue: %v", err)
		}

		return nil
	}

	req := &pb.SendTasksRequest{
		Project:   queue.conn.project,
		QueueName: queue.name,
		Tasks:     tasks,
	}
	var err error
	_, err = queue.conn.queuesClient.SendTasks(ctx, req)
	if err != nil {
		return fmt.Errorf("delay: cannot send tasks: %v", err)
	}

	return nil
}
