package crypt

import (
	"strings"
	"time"

	"github.com/altipla-consulting/errors"
	"github.com/gorilla/securecookie"
	"google.golang.org/protobuf/proto"
)

var (
	ErrEmptyToken   = errors.New("empty token")
	ErrExpiredToken = errors.New("expired token")
)

type Signer struct {
	signature, secret string
}

// NewSigner creates a new signer with the following params:
//   - signature: 32 random characters.
//   - secret: 16, 24, or 32 random characters to select AES-128, AES-192, or AES-256.
func NewSigner(signature, secret string) *Signer {
	return &Signer{signature, secret}
}

func (s *Signer) String() string {
	return "<libs.altipla.consulting/crypt.Signer>"
}

func (s *Signer) baseSC() *securecookie.SecureCookie {
	return securecookie.New([]byte(s.signature), []byte(s.secret)).MaxAge(0)
}

func (s *Signer) SignMessage(msg proto.Message, opts ...SignOption) (string, error) {
	sc := applyOpts(s.baseSC(), opts)

	encoded, err := proto.Marshal(msg)
	if err != nil {
		return "", errors.Trace(err)
	}

	token, err := sc.Encode("crypt", encoded)
	if err != nil {
		return "", errors.Trace(err)
	}

	return strings.TrimRight(token, "="), nil
}

func (s *Signer) ReadMessage(token string, msg proto.Message, opts ...SignOption) error {
	sc := applyOpts(s.baseSC(), opts)

	if token == "" {
		return errors.Trace(ErrEmptyToken)
	}
	if i := len(token) % 4; i != 0 {
		token += strings.Repeat("=", 4-i)
	}

	var encoded []byte
	if err := sc.Decode("crypt", token, &encoded); err != nil {
		if e, ok := err.(securecookie.Error); ok && e.IsDecode() {
			return errors.Trace(ErrExpiredToken)
		}

		return errors.Trace(err)
	}

	if err := proto.Unmarshal(encoded, msg); err != nil {
		return errors.Trace(err)
	}

	return nil
}

type configurator struct {
	ttl time.Duration
}

func applyOpts(sc *securecookie.SecureCookie, opts []SignOption) *securecookie.SecureCookie {
	c := new(configurator)
	for _, opt := range opts {
		opt(c)
	}

	if c.ttl != 0 {
		sc = sc.MaxAge(int(c.ttl / time.Second))
	}

	return sc
}

type SignOption func(c *configurator)

func WithTTL(ttl time.Duration) SignOption {
	return func(c *configurator) {
		c.ttl = ttl
	}
}
