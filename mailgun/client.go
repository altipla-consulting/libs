package mailgun

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/mail"
	"strings"

	"github.com/mailgun/mailgun-go/v3"
	log "github.com/sirupsen/logrus"
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
	From          mail.Address
	To            mail.Address
	Subject       string
	HTML          string
	Tags          []string
	ReplyTo       string
	Attachments   []*Attachment
	Inlines       []*Attachment
	UserVariables map[string]string
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

	msg := mgclient.NewMessage(email.From.String(), email.Subject, "", email.To.String())
	msg.SetHtml(email.HTML)
	if email.ReplyTo != "" {
		msg.SetReplyTo(email.ReplyTo)
	}
	for _, tag := range email.Tags {
		if err := msg.AddTag(tag); err != nil {
			return "", fmt.Errorf("cannot add tag %q of %v: %w", tag, email.Tags, err)
		}
	}
	for _, attachment := range email.Attachments {
		msg.AddReaderAttachment(attachment.Filename, ioutil.NopCloser(bytes.NewReader(attachment.Content)))
	}
	for _, attachment := range email.Inlines {
		msg.AddReaderInline(attachment.Filename, ioutil.NopCloser(bytes.NewReader(attachment.Content)))
	}
	for k, v := range email.UserVariables {
		if err := msg.AddVariable(k, v); err != nil {
			return "", fmt.Errorf("cannot add variable %q of %v: %w", k, email.UserVariables, err)
		}
	}

	message, id, err := mgclient.Send(ctx, msg)
	if err != nil {
		if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
			return "", ErrTimeout
		}
		if strings.HasPrefix(err.Error(), "remote server prematurely closed connection:") {
			return "", ErrTimeout
		}
		if strings.HasPrefix(err.Error(), "while making http request:") && strings.Contains(err.Error(), "read: connection reset by peer") {
			return "", ErrTimeout
		}

		var mgerr *mailgun.UnexpectedResponseError
		if errors.As(err, &mgerr) {
			errdata := new(sendError)
			if err := json.Unmarshal(mgerr.Data, errdata); err == nil {
				if errdata.Message == "'to' parameter is not a valid address. please check documentation" {
					return "", fmt.Errorf("email %q: %w", email.To.String(), ErrInvalidAddress)
				}
				if errdata.Message == "to parameter is not a valid address. please check documentation" {
					return "", fmt.Errorf("email %q: %w", email.To.String(), ErrInvalidAddress)
				}
			}
		}

		return "", fmt.Errorf("send failed: %w", err)
	}

	log.WithFields(log.Fields{"id": id, "message": message}).Debug("Mailgun email sent")

	return id, nil
}

func (client *Client) Send(ctx context.Context, domain string, email *Email) error {
	_, err := client.SendReturnID(ctx, domain, email)
	return err
}

func (client *Client) ValidateEmail(ctx context.Context, email string) (bool, error) {
	validator := mailgun.NewEmailValidator(client.apiKey)

	ev, err := validator.ValidateEmail(ctx, email, false)
	if err != nil {
		return false, fmt.Errorf("validate %q failed: %w", email, err)
	}

	return ev.IsValid, nil
}
