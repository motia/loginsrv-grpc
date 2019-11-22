package main

import (
	"log"
	"net"

	auth "github.com/motia/loginsrv-grpc"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"

	"google.golang.org/grpc"
)

const (
	port = ":50051"
)

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	server := auth.NewLoginSrvServer("http://localhost:8080")

	s := grpc.NewServer(
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			// grpc_ctxtags.StreamServerInterceptor(),
			// grpc_opentracing.StreamServerInterceptor(),
			// grpc_prometheus.StreamServerInterceptor,
			// grpc_zap.StreamServerInterceptor(zapLogger),
			grpc_auth.StreamServerInterceptor(server.Authenticate),
			grpc_recovery.StreamServerInterceptor(),
		)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			// grpc_ctxtags.UnaryServerInterceptor(),
			// grpc_opentracing.UnaryServerInterceptor(),
			// grpc_prometheus.UnaryServerInterceptor,
			// grpc_zap.UnaryServerInterceptor(zapLogger),
			grpc_auth.UnaryServerInterceptor(server.Authenticate),
			grpc_recovery.UnaryServerInterceptor(),
		)),
	)
	auth.RegisterAuthServer(s, server)
	log.Println("Auth sever started")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
