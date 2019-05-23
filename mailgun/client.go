package mailgun

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"net/mail"
	"time"

	"github.com/mailgun/mailgun-go"
	log "github.com/sirupsen/logrus"

	"libs.altipla.consulting/errors"
)

var (
	ErrTimeout        = errors.New("timeout")
	ErrInvalidAddress = errors.New("invalid address")
)

type Sender interface {
	Send(ctx context.Context, domain string, email *Email) error
}

type SenderReturnID interface {
	SendReturnID(ctx context.Context, domain string, email *Email) (string, error)
}

type Validator interface {
	ValidateEmail(ctx context.Context, email string) (bool, error)
}

type Client struct {
	apiKey string
}

func NewClient(apiKey string) *Client {
	return &Client{apiKey}
}

type Email struct {
	From        mail.Address
	To          mail.Address
	Subject     string
	HTML        string
	Tags        []string
	ReplyTo     string
	Attachments []*Attachment
}

type Attachment struct {
	Filename string
	Content  []byte
}

type sendError struct {
	Message string `json:"message"`
}

func (client *Client) SendReturnID(ctx context.Context, domain string, email *Email) (string, error) {
	mgclient := mailgun.NewMailgun(domain, client.apiKey)

	deadline, ok := ctx.Deadline()
	if ok {
		mgclient.SetClient(&http.Client{
			Timeout: deadline.Sub(time.Now()),
		})
	}

	msg := mailgun.NewMessage(email.From.String(), email.Subject, "", email.To.String())
	msg.SetHtml(email.HTML)
	if email.ReplyTo != "" {
		msg.SetReplyTo(email.ReplyTo)
	}
	for _, tag := range email.Tags {
		msg.AddTag(tag)
	}

	for _, attachment := range email.Attachments {
		msg.AddReaderAttachment(attachment.Filename, ioutil.NopCloser(bytes.NewReader(attachment.Content)))
	}

	message, id, err := mgclient.Send(msg)
	if err != nil {
		if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
			return "", ErrTimeout
		}
		if mgerr, ok := err.(*mailgun.UnexpectedResponseError); ok {
			errdata := new(sendError)
			if err := json.Unmarshal(mgerr.Data, errdata); err == nil {
				switch errdata.Message {
				case "'to' parameter is not a valid address. please check documentation":
					return "", errors.Wrapf(ErrInvalidAddress, "email: %s", email.To.String())
				}
			}
		}

		return "", errors.Wrapf(err, "send failed")
	}

	log.WithFields(log.Fields{"id": id, "message": message}).Info("Mailgun email sent")

	return id, nil
}

func (client *Client) Send(ctx context.Context, domain string, email *Email) error {
	if _, err := client.SendReturnID(ctx, domain, email); err != nil {
		return errors.Trace(err)
	}

	return nil
}

func (client *Client) ValidateEmail(ctx context.Context, email string) (bool, error) {
	validator := mailgun.NewEmailValidator(client.apiKey)

	deadline, ok := ctx.Deadline()
	if ok {
		validator.SetClient(&http.Client{
			Timeout: deadline.Sub(time.Now()),
		})
	}

	ev, err := validator.ValidateEmail(email, false)
	if err != nil {
		return false, errors.Wrapf(err, "validate failed")
	}

	return ev.IsValid, nil
}
