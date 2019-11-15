// Package main implements a client for Greeter service.
package main

import (
	"context"
	"log"
	"time"

	auth "github.com/motia/proto-auth/auth"
	"google.golang.org/grpc"
)

const (
	address     = "localhost:50051"
	defaultName = "world"
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := auth.NewAuthClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err = c.GetProfile(ctx, &auth.ProfileRequest{})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Println("GOT USER")
}
