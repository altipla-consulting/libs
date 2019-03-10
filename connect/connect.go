package connect

import (
	"context"
	"crypto/tls"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"

	"libs.altipla.consulting/errors"
	"libs.altipla.consulting/services"
)

type Endpoint = services.Endpoint
type RemoteAddress string

const beauthTokenEndpoint = "https://beauth.io/token"

type oauthAccess struct {
	tokenSource oauth2.TokenSource
}

func (oa oauthAccess) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	token, err := oa.tokenSource.Token()
	if err != nil {
		return nil, errors.Wrapf(err, "cannot update token")
	}

	return map[string]string{
		"authorization": token.Type() + " " + token.AccessToken,
	}, nil
}

func (oa oauthAccess) RequireTransportSecurity() bool {
	return false
}

// OAuthToken opens a connection using OAuth2 tokens obtained from BeAuth.io users or clients.
func OAuthToken(address, clientID, clientSecret string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	if address == "" {
		return nil, errors.Errorf("remote address required")
	}
	if clientID == "" || clientSecret == "" {
		return nil, errors.Errorf("client credentials required to connect with oauth to a remote address")
	}

	config := &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     beauthTokenEndpoint,
	}
	rpcCreds := grpc.WithPerRPCCredentials(oauthAccess{
		tokenSource: config.TokenSource(context.Background()),
	})
	creds := credentials.NewTLS(&tls.Config{
		ServerName: address,
	})
	opts = append(opts, grpc.WithTransportCredentials(creds), grpc.WithKeepaliveParams(keepalive.ClientParameters{Time: 7 * time.Minute}), rpcCreds)
	conn, err := grpc.Dial(address+":443", opts...)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot connect to remote address %s", address)
	}
	return conn, nil
}

// Insecure opens an insecure local connection to debug in the dev machines. Unsuitable
// for production because everything would have to be in the same Kubernetes cluster.
func Insecure(address string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	if address == "" {
		return nil, errors.Errorf("remote address required")
	}

	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(address, opts...)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot connect to remote address %s", address)
	}
	return conn, nil
}

// Local opens a connection to the nearest endpoint. In development it uses an insecure local
// channel and in production it uses the client credentials to request an OAuth2 token from
// BeAuth.io and authenticates every request with it.
func Local(address, clientID, clientSecret string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	if services.IsLocal() {
		return Insecure(address, opts...)
	}
	return OAuthToken(address, clientID, clientSecret, opts...)
}

type bearerAccess struct {
	bearer string
}

func (access bearerAccess) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{"authorization": "Bearer " + access.bearer}, nil
}

func (access bearerAccess) RequireTransportSecurity() bool {
	return true
}

func WithBearer(bearer string) grpc.DialOption {
	return grpc.WithPerRPCCredentials(bearerAccess{bearer})
}

func Remote(address RemoteAddress, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	creds := credentials.NewTLS(&tls.Config{
		ServerName: string(address),
	})
	opts = append(opts, grpc.WithTransportCredentials(creds))
	return grpc.Dial(string(address), opts...)
}

func Internal(endpoint Endpoint, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	opts = append(opts, grpc.WithInsecure())
	return grpc.Dial(string(endpoint), opts...)
}
