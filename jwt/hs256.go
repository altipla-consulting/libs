package jwt

import (
	jose "gopkg.in/square/go-jose.v2"
	"libs.altipla.consulting/errors"
)

type hs256 struct {
	key []byte
}

func (crypto *hs256) Signer() (jose.Signer, error) {
	var opts = jose.SignerOptions{}
	opts.WithType("JWT")
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.HS256, Key: crypto.key}, &opts)
	return signer, errors.Trace(err)
}

func (crypto *hs256) Key(kid string) interface{} {
	return crypto.key
}

func (crypto *hs256) Close() {
}
