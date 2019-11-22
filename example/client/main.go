// Package main implements a client for Greeter service.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	loginsrv_grpc "github.com/motia/loginsrv-grpc"
	"google.golang.org/grpc"
)

const (
	address = "localhost:50051"
)

type tokenStorage struct {
	token *string
}

func (ts *tokenStorage) getToken() *string {
	return ts.token
}

func main() {
	ts := &tokenStorage{}
	tokenAdderInterceptor := grpc.UnaryClientInterceptor(
		loginsrv_grpc.NewClientTokenInterceptor(ts.getToken))
	conn, err := grpc.Dial(
		address,
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithChainUnaryInterceptor(tokenAdderInterceptor),
	)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := loginsrv_grpc.NewAuthClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	loginReply, err := c.AttemptLogin(ctx, &loginsrv_grpc.LoginRequest{Username: "bob", Password: "secret"})

	if err != nil {
		log.Fatalf("server error: %v", err)
	}
	ts.token = &loginReply.AccessToken
	fmt.Println("Login Reply " + loginReply.AccessToken)

	refreshReply, err := c.RefreshToken(ctx, &loginsrv_grpc.RefreshRequest{})
	fmt.Println(refreshReply == nil, err)
	if err != nil {
		log.Fatalf("server error: %v", err)
	}
	fmt.Println("Refresh Reply: " + refreshReply.AccessToken)

	profileReply, err := c.GetProfile(ctx, &loginsrv_grpc.ProfileRequest{})
	if err != nil {
		log.Fatalf("server error: %v", err)
	}
	fmt.Println("Profile Reply: sub=" + profileReply.Sub)
}
