package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/Brilhante29/grpc-vs-rest-bench/internal"
	pb "github.com/Brilhante29/grpc-vs-rest-bench/internal/proto/benchmark/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type grpcServer struct {
	pb.UnimplementedBenchmarkServiceServer
	logic *internal.EchoLogic
}

func (s *grpcServer) Echo(ctx context.Context, req *pb.EchoRequest) (*pb.EchoResponse, error) {
	return s.logic.Echo(req), nil
}

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
	pb.RegisterBenchmarkServiceServer(s, &grpcServer{logic: internal.NewEchoLogic()})
	reflection.Register(s)
	log.Printf("gRPC server listening on :%s", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
