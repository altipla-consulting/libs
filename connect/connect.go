package connect

import (
	"context"
	"crypto/tls"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
)

// Endpoint completely defines a service, locally and remotely. Both are optional
// depending on the function of this package you call to connect.
type Endpoint struct {
	Internal, Remote string
}

type oauthAccess struct {
	accessToken string
}

func (oa oauthAccess) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": "Bearer " + oa.accessToken,
	}, nil
}

func (oa oauthAccess) RequireTransportSecurity() bool {
	return false
}

// WithBearer dials the remote server with a Bearer token in every request.
func WithBearer(accessToken string) grpc.DialOption {
	return grpc.WithPerRPCCredentials(&oauthAccess{accessToken})
}

// Remote opens a new connection to the remote part of the endpoint.
func Remote(endpoint Endpoint, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	if endpoint.Remote == "" {
		panic("remote endpoint required")
	}

	creds := credentials.NewTLS(&tls.Config{ServerName: string(endpoint.Remote)})
	opts = append(opts, grpc.WithTransportCredentials(creds))

	// Wait until the connection cannot be opened before failing.
	opts = append(opts, grpc.WithDefaultCallOptions(grpc.WaitForReady(true)))

	// A timeout of 7 minutes is more than the gRPC minimum of 5 minutes enforced
	// on the server but less than the 10 minutes automatic disconnection the Google
	// Cloud TCP Load Balancer has that drops our connections.
	opts = append(opts, grpc.WithKeepaliveParams(keepalive.ClientParameters{Time: 7 * time.Minute}))

	return grpc.Dial(string(endpoint.Remote)+":443", opts...)
}

// Internal opens a new connection to the internal part of the endpoint.
func Internal(endpoint Endpoint, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	if endpoint.Internal == "" {
		panic("internal endpoint required")
	}

	opts = append(opts, grpc.WithInsecure())

	// Wait until the connection cannot be opened before failing.
	opts = append(opts, grpc.WithDefaultCallOptions(grpc.WaitForReady(true)))

	// Move auto-* headers from one service to the next one.
	opts = append(opts, grpc.WithPerRPCCredentials(new(autoMetadataCredentials)))

	// We do not want internal connections through Envoy to fail once the timeout is
	// reached. We send a periodic keepalive every hour to keep it open.
	opts = append(opts, grpc.WithKeepaliveParams(keepalive.ClientParameters{Time: 1 * time.Hour}))

	return grpc.Dial(string(endpoint.Internal), opts...)
}

type autoMetadataCredentials struct{}

func (amc *autoMetadataCredentials) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	outgoing := map[string]string{}

	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		for k, v := range md {
			if strings.HasPrefix(k, "auto-") {
				outgoing[k] = v[0]
			}
		}
	}

	return outgoing, nil
}

func (amc *autoMetadataCredentials) RequireTransportSecurity() bool {
	return false
}
