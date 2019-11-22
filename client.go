package loginsrv_grpc

import (
	"context"

	"google.golang.org/grpc"
	md "google.golang.org/grpc/metadata"
)

const (
	// AuthTokenMetadataKey is the key of for auth token in the metadata of RPCs
	AuthTokenMetadataKey = "authorization"
)

// NewClientTokenInterceptor attaches a token to the outgoing RPC
func NewClientTokenInterceptor(tokenGetter TokenGetter) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req interface{},
		reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption) error {

		if token := tokenGetter(); token != nil && len(*token) > 0 {
			ctx = md.AppendToOutgoingContext(ctx, AuthTokenMetadataKey, "bearer "+*token)
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// TokenGetter returns a jwt token
type TokenGetter func() *string
