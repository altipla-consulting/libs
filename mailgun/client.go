package mailgun

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/mail"
	"time"

	"github.com/mailgun/mailgun-go"
	log "github.com/sirupsen/logrus"
)

var ErrTimeout = fmt.Errorf("mailgun: timeout")

type Sender interface {
	Send(ctx context.Context, domain string, email *Email) error
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

type SendRejectedError struct {
	Reason string
}

func (err SendRejectedError) Error() string {
	return "mailgun: send rejected: " + err.Reason
}

func (client *Client) Send(ctx context.Context, domain string, email *Email) error {
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
			return ErrTimeout
		}

		return fmt.Errorf("mailgun: send failed: %s", err)
	}

	log.WithFields(log.Fields{"id": id, "message": message}).Info("Mailgun email sent")

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
		return false, fmt.Errorf("mailgun: validate failed: %s", err)
	}

	return ev.IsValid, nil
}
