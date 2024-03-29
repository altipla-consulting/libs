package pubsub

import (
	"context"
	"encoding/json"
	"os"
	"sync"
	"time"

	"cloud.google.com/go/compute/metadata"
	"cloud.google.com/go/pubsub"
	"github.com/altipla-consulting/errors"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
	"github.com/altipla-consulting/env"
)

type Client struct {
	project              string
	disableLocalEmulator bool
	c                    *pubsub.Client

	mx     sync.Mutex
	topics []*pubsub.Topic
}

type ClientOption func(client *Client)

func WithProject(project string) ClientOption {
	return func(client *Client) {
		client.project = project
	}
}

func DisableLocalEmulator() ClientOption {
	return func(client *Client) {
		client.disableLocalEmulator = true
	}
}

func NewClient(opts ...ClientOption) (*Client, error) {
	client := new(Client)
	for _, opt := range opts {
		opt(client)
	}

	if !client.disableLocalEmulator && env.IsLocal() {
		client.project = "local"
		if os.Getenv("PUBSUB_EMULATOR_HOST") == "" {
			_ = os.Setenv("PUBSUB_EMULATOR_HOST", "localhost:12001")
		}
	}

	if client.project == "" {
		var err error
		client.project, err = metadata.ProjectID()
		if err != nil {
			return nil, errors.Trace(err)
		}
	}

	var err error
	client.c, err = pubsub.NewClient(context.Background(), client.project)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return client, nil
}

func (client *Client) Close() {
	client.mx.Lock()
	defer client.mx.Unlock()

	for _, topic := range client.topics {
		topic.Stop()
	}
	client.c.Close()
}

func (client *Client) Topic(name string) *Topic {
	topic := client.c.Topic(name)
	topic.PublishSettings.CountThreshold = 1
	topic.PublishSettings.DelayThreshold = 0

	client.mx.Lock()
	defer client.mx.Unlock()
	client.topics = append(client.topics, topic)

	return &Topic{
		project: client.project,
		c:       client.c,
		t:       topic,
	}
}

type Topic struct {
	project string
	c       *pubsub.Client
	t       *pubsub.Topic
}

type PublishOption func(msg *pubsub.Message)

func WithAttribute(key, value string) PublishOption {
	return func(msg *pubsub.Message) {
		if msg.Attributes == nil {
			msg.Attributes = make(map[string]string)
		}
		msg.Attributes[key] = value
	}
}

func (topic *Topic) PublishProto(ctx context.Context, data proto.Message, opts ...PublishOption) error {
	encoded, err := proto.Marshal(data)
	if err != nil {
		return errors.Trace(err)
	}
	return errors.Trace(topic.PublishBytes(ctx, encoded, opts...))
}

func (topic *Topic) PublishJSON(ctx context.Context, value interface{}, opts ...PublishOption) error {
	encoded, err := json.Marshal(value)
	if err != nil {
		return errors.Trace(err)
	}
	return errors.Trace(topic.PublishBytes(ctx, encoded, opts...))
}

func (topic *Topic) PublishBytes(ctx context.Context, value []byte, opts ...PublishOption) error {
	msg := &pubsub.Message{Data: value}
	for _, opt := range opts {
		opt(msg)
	}
	res := topic.t.Publish(ctx, msg)
	if res == nil {
		return errors.Errorf("cannot send message, result is nil")
	}
	_, err := res.Get(ctx)
	return errors.Trace(err)
}

type LocalSubscribeOption func(cnf *pubsub.SubscriptionConfig)

func WithFilter(filter string) LocalSubscribeOption {
	return func(cnf *pubsub.SubscriptionConfig) {
		cnf.Filter = filter
	}
}

func (topic *Topic) LocalAssertSubscription(ctx context.Context, name string, opts ...LocalSubscribeOption) error {
	if !env.IsLocal() {
		return nil
	}

	exists, err := topic.t.Exists(ctx)
	if err != nil {
		return errors.Trace(err)
	}
	if !exists {
		log.WithField("name", topic.t.ID()).Info("Creating fake topic in the local emulator")
		if _, err := topic.c.CreateTopic(ctx, topic.t.ID()); err != nil {
			return errors.Trace(err)
		}
	}

	exists, err = topic.c.Subscription(name).Exists(ctx)
	if err != nil {
		return errors.Trace(err)
	}
	if !exists {
		log.WithField("name", topic.t.ID()).Info("Creating fake subscription in the local emulator")
		cnf := pubsub.SubscriptionConfig{
			Topic: topic.t,
			PushConfig: pubsub.PushConfig{
				Endpoint: "http://" + env.ServiceName() + ":8080/_/pubsub/" + name,
			},
			AckDeadline: 60 * time.Second,
		}
		for _, opt := range opts {
			opt(&cnf)
		}
		if _, err := topic.c.CreateSubscription(ctx, name, cnf); err != nil {
			return errors.Trace(err)
		}
	}

	return nil
}

type PushRequest struct {
	Subscription string   `json:"subscription"`
	Message      *Message `json:"message"`
}

type Message struct {
	ID         string            `json:"messageId"`
	Attributes map[string]string `json:"attributes"`

	RawData []byte `json:"data"`
}

func (msg *Message) ReadProto(dest proto.Message) error {
	return errors.Trace(proto.Unmarshal(msg.RawData, dest))
}

func (msg *Message) ReadJSON(dest interface{}) error {
	return errors.Trace(json.Unmarshal(msg.RawData, dest))
}
