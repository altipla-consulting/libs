package security

import (
	"context"
	"strings"

	"google.golang.org/grpc/metadata"
)

func ReadServiceAuthorization(ctx context.Context) string {
	md, _ := metadata.FromIncomingContext(ctx)
	if auth := md.Get("authorization"); len(auth) > 0 {
		parts := strings.SplitN(auth[0], " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1]
		}
	}
	return ""
}
