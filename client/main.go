// Package main implements a client for Greeter service.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	auth "github.com/motia/proto-auth/auth"
	"google.golang.org/grpc"
	md "google.golang.org/grpc/metadata"
)

const (
	address = "localhost:50051"
)

func authHeaderAdder(token *string) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req interface{},
		reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption) error {

		if len(*token) > 0 {
			ctx = md.AppendToOutgoingContext(ctx, "authorization", "bearer "+*token)
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func main() {
	token := ""
	conn, err := grpc.Dial(
		address,
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithChainUnaryInterceptor(grpc.UnaryClientInterceptor(authHeaderAdder(&token))),
	)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := auth.NewAuthClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	loginReply, err := c.AttemptLogin(ctx, &auth.LoginRequest{Username: "bob", Password: "secret"})

	if err != nil {
		log.Fatalf("server sucks: %v", err)
	}
	token = loginReply.AccessToken
	fmt.Println("Login Reply " + loginReply.AccessToken)

	refreshReply, err := c.RefreshToken(ctx, &auth.RefreshRequest{})
	fmt.Println(refreshReply == nil, err)
	if err != nil {
		log.Fatalf("server sucks: %v", err)
	}
	fmt.Println("Refresh Reply " + refreshReply.AccessToken)
}
