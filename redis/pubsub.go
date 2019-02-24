package redis

import (
	"github.com/go-redis/redis"
	"github.com/golang/protobuf/proto"

	"libs.altipla.consulting/errors"
)

// PubSub represents a connection to a redis PubSub. It can be used to publish
// and receive protobuf messages.
type PubSub struct {
	db   *Database
	name string
}

// Subscribe opens a new connection to the server and starts downloading messages.
func (pubsub *PubSub) Subscribe() *PubSubSubscription {
	ps := pubsub.db.sess.Subscribe(pubsub.name)

	return &PubSubSubscription{
		ps: ps,
		ch: ps.Channel(),
	}
}

// Publish sends a new message to the server. Only subscription connected at
// the same time will receive the message.
func (pubsub *PubSub) Publish(msg proto.Message) error {
	serialized, err := proto.Marshal(msg)
	if err != nil {
		return errors.Wrapf(err, "cannot serialize pubsub message")
	}
	if err := pubsub.db.sess.Publish(pubsub.name, string(serialized)).Err(); err != nil {
		return errors.Wrapf(err, "cannot publish pubsub message")
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
		return errors.Wrapf(err, "cannot parse pubsub message")
	}

	return nil
}
