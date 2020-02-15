package jwt

import (
	"fmt"
	"strconv"
	"time"

	jose "gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
	"libs.altipla.consulting/errors"

	"libs.altipla.consulting/clock"
)

type InvalidTokenError struct {
	Reason string
}

func (err *InvalidTokenError) Error() string {
	return "jwt: invalid token: " + err.Reason
}

type Generator struct {
	key   string
	clock clock.Clock
}

func NewHS256(key string) *Generator {
	return &Generator{
		key:   key,
		clock: clock.New(),
	}
}

type Claims struct {
	// Identifies the principal that issued the JWT.
	Issuer string

	// Identifies the principal that is the subject of the JWT.
	Subject string

	// Identifies the recipients that the JWT is intended for.
	Audience string

	// Identifies the expiration time on or after which the JWT MUST NOT be accepted
	// for processing.
	Expiry time.Time

	// Identifies the time before which the JWT MUST NOT be accepted for processing.
	//
	// It is optional.
	NotBefore time.Time

	// Identifies the time at which the JWT was issued.
	//
	// It is optional, as it will be filled with the current timestamp when signing.
	IssuedAt time.Time
}

func (claims *Claims) SetSubjectInt(subject int64) {
	claims.Subject = strconv.FormatInt(subject, 10)
}

func (claims Claims) SubjectInt() int64 {
	n, _ := strconv.ParseInt(claims.Subject, 10, 64)
	return n
}

type CustomClaimsImplementation interface {
	isCustomClaims()
}

type CustomClaims struct{}

func (cc CustomClaims) isCustomClaims() {}

func (g *Generator) Sign(claims Claims, customs ...CustomClaimsImplementation) (string, error) {
	if claims.Issuer == "" {
		return "", errors.Errorf("issuer is required in jwt claims")
	}
	if claims.Subject == "" {
		return "", errors.Errorf("subject is required in jwt claims")
	}
	if claims.Audience == "" {
		return "", errors.Errorf("intended audience is required in jwt claims")
	}
	if claims.Expiry.IsZero() {
		return "", errors.Errorf("expiry timestamp is required in jwt claims")
	}
	// TODO(ernesto): Validar que la fecha de caducidad sea mayor a la fecha actual.

	if claims.IssuedAt.IsZero() {
		claims.IssuedAt = g.clock.Now()
	}

	var opts = jose.SignerOptions{}
	opts.WithType("JWT")
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.HS256, Key: []byte(g.key)}, &opts)
	if err != nil {
		return "", errors.Trace(err)
	}

	builder := jwt.Signed(signer)
	builder = builder.Claims(jwt.Claims{
		Issuer:    claims.Issuer,
		Subject:   claims.Subject,
		Audience:  jwt.Audience{claims.Audience},
		Expiry:    jwt.NewNumericDate(claims.Expiry),
		NotBefore: jwt.NewNumericDate(claims.NotBefore),
		IssuedAt:  jwt.NewNumericDate(claims.IssuedAt),
	})

	for _, custom := range customs {
		builder = builder.Claims(custom)
	}

	token, err := builder.CompactSerialize()
	if err != nil {
		return "", errors.Trace(err)
	}

	return token, nil
}

type Expected struct {
	// Expected principal that issued the JWT.
	Issuer string

	// Expected audience for the token in case it is intended for another use.
	Audience string
}

func (g *Generator) Extract(token string, expected Expected, claims *Claims, customs ...CustomClaimsImplementation) error {
	if expected.Issuer == "" {
		return errors.Errorf("extract expected issuer required")
	}
	if expected.Audience == "" {
		return errors.Errorf("extract expected audience required")
	}

	webSignature, err := jose.ParseSigned(token)
	if err != nil {
		if err.Error() == "square/go-jose: compact JWS format must have three parts" {
			return &InvalidTokenError{
				Reason: fmt.Sprintf("invalid parts"),
			}
		}

		return errors.Trace(err)
	}
	if _, err := webSignature.Verify([]byte(g.key)); err != nil {
		if errors.Is(err, jose.ErrCryptoFailure) {
			return &InvalidTokenError{
				Reason: fmt.Sprintf("invalid signature"),
			}
		}

		return errors.Trace(err)
	}

	webToken, err := jwt.ParseSigned(token)
	if err != nil {
		return errors.Trace(err)
	}

	var webClaims jwt.Claims
	var extract []interface{}
	extract = append(extract, &webClaims)
	for _, custom := range customs {
		extract = append(extract, custom)
	}
	if err := webToken.Claims([]byte(g.key), extract...); err != nil {
		return errors.Trace(err)
	}

	exp := jwt.Expected{
		Issuer:   expected.Issuer,
		Audience: jwt.Audience{expected.Audience},
		Time:     g.clock.Now(),
	}
	if err := webClaims.Validate(exp); err != nil {
		switch {
		case errors.Is(err, jwt.ErrInvalidIssuer):
			return &InvalidTokenError{
				Reason: fmt.Sprintf("invalid issuer: expected <%v> claims <%v>", expected.Issuer, claims.Issuer),
			}

		case errors.Is(err, jwt.ErrInvalidAudience):
			return &InvalidTokenError{
				Reason: fmt.Sprintf("invalid audience: expected <%v> claims <%v>", expected.Audience[0], claims.Audience),
			}

		case errors.Is(err, jwt.ErrNotValidYet):
			return &InvalidTokenError{
				Reason: fmt.Sprintf("invalid not before: expected <%v> claims <%v>", exp.Time.Format(time.RFC3339), claims.NotBefore.Format(time.RFC3339)),
			}

		case errors.Is(err, jwt.ErrExpired):
			return &InvalidTokenError{
				Reason: fmt.Sprintf("expired: expected <%v> claims <%v>", exp.Time.Format(time.RFC3339), claims.Expiry.Format(time.RFC3339)),
			}

		case errors.Is(err, jwt.ErrIssuedInTheFuture):
			return &InvalidTokenError{
				Reason: fmt.Sprintf("issued in the future: expected <%v> claims <%v>", exp.Time.Format(time.RFC3339), claims.IssuedAt.Format(time.RFC3339)),
			}

		default:
			return errors.Trace(err)
		}
	}

	claims.Issuer = webClaims.Issuer
	claims.Subject = webClaims.Subject
	claims.Audience = webClaims.Audience[0]
	claims.Expiry = webClaims.Expiry.Time()
	claims.NotBefore = webClaims.NotBefore.Time()
	claims.IssuedAt = webClaims.IssuedAt.Time()

	return nil
}
