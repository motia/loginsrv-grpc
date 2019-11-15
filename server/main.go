package main

import (
	"context"
	"log"
	"net"

	auth "github.com/motia/proto-auth/auth"

	"google.golang.org/grpc"
)

const (
	port = ":50051"
)

type server struct {
	auth.UnimplementedAuthServer
}

func (s *server) AttemptLogin(ctx context.Context, in *auth.LoginRequest) (*auth.LoginReply, error) {
	return &auth.LoginReply{AccessToken: "First token"}, nil
}

func (s *server) RefreshToken(ctx context.Context, in *auth.RefreshRequest) (*auth.LoginReply, error) {
	return &auth.LoginReply{AccessToken: "Refreshed token"}, nil
}

func (s *server) GetProfile(ctx context.Context, in *auth.ProfileRequest) (*auth.Profile, error) {
	return &auth.Profile{}, nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	auth.RegisterAuthServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
