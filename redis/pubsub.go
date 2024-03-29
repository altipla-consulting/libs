package redis

import (
	"context"
	"fmt"

	"github.com/go-redis/redis"
	"google.golang.org/protobuf/proto"
)

// PubSub represents a connection to a redis PubSub. It can be used to publish
// and receive protobuf messages.
type PubSub struct {
	db   *Database
	name string
}

// Subscribe opens a new connection to the server and starts downloading messages.
func (pubsub *PubSub) Subscribe(ctx context.Context) *PubSubSubscription {
	client, ok := pubsub.db.Cmdable(ctx).(*redis.Client)
	if !ok {
		panic("cannot subscribe inside a redis transaction")
	}
	ps := client.Subscribe(pubsub.name)

	return &PubSubSubscription{
		ps: ps,
		ch: ps.Channel(),
	}
}

// Publish sends a new message to the server. Only subscription connected at
// the same time will receive the message.
func (pubsub *PubSub) Publish(ctx context.Context, msg proto.Message) error {
	serialized, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("cannot serialize pubsub message: %v", err)
	}
	if err := pubsub.db.Cmdable(ctx).Publish(pubsub.name, string(serialized)).Err(); err != nil {
		return fmt.Errorf("cannot publish pubsub message: %v", err)
	}
	return nil
}

// PubSubSubscription stores the state of an active connection to the server.
type PubSubSubscription struct {
	ps *redis.PubSub
	ch <-chan *redis.Message
}

// Close exits the connection.
func (sub *PubSubSubscription) Close() {
	sub.ps.Close()
}

// Next waits for the next message and decodes it in the destination.
func (sub *PubSubSubscription) Next(dest proto.Message) error {
	msg := <-sub.ch
	if msg == nil {
		return ErrDone
	}

	if err := proto.Unmarshal([]byte(msg.Payload), dest); err != nil {
		return fmt.Errorf("cannot parse pubsub message %w", err)
	}

	return nil
}
