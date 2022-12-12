package jwt

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
	jose "gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"

	"libs.altipla.consulting/clock"
	"libs.altipla.consulting/errors"
)

type InvalidTokenError struct {
	Reason string
}

func (err *InvalidTokenError) Error() string {
	return "jwt: invalid token: " + err.Reason
}

type Generator struct {
	crypto cryptoImpl
	clock  clock.Clock
}

type cryptoImpl interface {
	Signer() (jose.Signer, error)
	Key(kid string) interface{}
	Close()
}

func NewHS256(key string) *Generator {
	return &Generator{
		crypto: &hs256{
			key: []byte(key),
		},
		clock: clock.New(),
	}
}

func NewHS256Base64(encodedKey string) *Generator {
	key, err := base64.StdEncoding.DecodeString(encodedKey)
	if err != nil {
		log.Fatalf("jwt key is not in base64 format: %s: %s", encodedKey, err.Error())
	}

	return &Generator{
		crypto: &hs256{
			key: key,
		},
		clock: clock.New(),
	}
}

// NewRS256FromWellKnown prepares a generator that only reads signed JWTs verifying
// them with the cached keys obtained from the well known URL.
//
// The well known URL is something like this:
//
//	https://{domain}/.well-known/jwks.json
func NewRS256FromWellKnown(wkurl string) *Generator {
	c := &rs256{
		wkurl: wkurl,
	}
	c.backgroundGetKeys()
	return &Generator{
		crypto: c,
		clock:  clock.New(),
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

	signer, err := g.crypto.Signer()
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

	// Verify the signature manually before reading the JWT token to detect
	// signing problems with a custom error. Also retrieves the correct key depending
	// on the signature of the token.
	signature, err := jose.ParseSigned(token)
	if err != nil {
		if err.Error() == "square/go-jose: compact JWS format must have three parts" {
			return &InvalidTokenError{
				Reason: "invalid parts",
			}
		}

		return errors.Trace(err)
	}
	if len(signature.Signatures) != 1 {
		return &InvalidTokenError{
			Reason: fmt.Sprintf("invalid number of signatures: %d", len(signature.Signatures)),
		}
	}
	key := g.crypto.Key(signature.Signatures[0].Header.KeyID)
	if key == nil {
		return &InvalidTokenError{
			Reason: fmt.Sprintf("invalid signature key id: %s", signature.Signatures[0].Header.KeyID),
		}
	}
	if _, err := signature.Verify(key); err != nil {
		if errors.Is(err, jose.ErrCryptoFailure) {
			return &InvalidTokenError{
				Reason: fmt.Sprintf("invalid signature"),
			}
		}

		return errors.Trace(err)
	}

	// Verify (again) and extract the claims from the JWT.
	parsedToken, err := jwt.ParseSigned(token)
	if err != nil {
		return errors.Trace(err)
	}
	var extract []interface{}
	var extractedClaims jwt.Claims
	extract = append(extract, &extractedClaims)
	for _, custom := range customs {
		extract = append(extract, custom)
	}
	if err := parsedToken.Claims(key, extract...); err != nil {
		return errors.Trace(err)
	}

	// Check claims against the expected values.
	exp := jwt.Expected{
		Issuer:   expected.Issuer,
		Audience: jwt.Audience{expected.Audience},
		Time:     g.clock.Now(),
	}
	if err := extractedClaims.Validate(exp); err != nil {
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

	// Read the claims extracted from the token.
	claims.Issuer = extractedClaims.Issuer
	claims.Subject = extractedClaims.Subject
	claims.Audience = extractedClaims.Audience[0]
	claims.Expiry = extractedClaims.Expiry.Time()
	claims.NotBefore = extractedClaims.NotBefore.Time()
	claims.IssuedAt = extractedClaims.IssuedAt.Time()

	return nil
}

func (g *Generator) Close() {
	g.crypto.Close()
}
