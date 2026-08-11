package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/Brilhante29/grpc-vs-rest-bench/internal"
	pb "github.com/Brilhante29/grpc-vs-rest-bench/internal/proto/benchmark/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "50051"
	}
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterBenchmarkServiceServer(s, internal.NewGRPCServer(internal.NewEchoLogic()))
	reflection.Register(s)
	log.Printf("gRPC server listening on :%s", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
