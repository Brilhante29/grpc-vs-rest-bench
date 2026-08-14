package internal_test

import (
	"bytes"
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Brilhante29/grpc-vs-rest-bench/internal"
	pb "github.com/Brilhante29/grpc-vs-rest-bench/internal/proto/benchmark/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestRESTAndGRPCPreserveTheSameSemanticContract(t *testing.T) {
	logic := internal.NewEchoLogic()
	restServer := httptest.NewServer(internal.NewRESTHandler(logic))
	defer restServer.Close()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterBenchmarkServiceServer(grpcServer, internal.NewGRPCServer(logic))
	go func() { _ = grpcServer.Serve(listener) }()
	defer grpcServer.Stop()

	timeout := 5 * time.Second
	restClient := internal.NewRESTClient(strings.TrimPrefix(restServer.URL, "http://"), 1, timeout)
	grpcClient, err := internal.NewGRPCClient(listener.Addr().String(), timeout)
	if err != nil {
		t.Fatal(err)
	}
	defer grpcClient.Close()

	request := internal.EchoRequest{RequestID: "contract-1", Payload: strings.Repeat("same-payload-", 32)}
	restResponse, err := restClient.Echo(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	grpcResponse, err := grpcClient.Echo(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if restResponse != grpcResponse {
		t.Fatalf("transport parity failed: REST=%#v gRPC=%#v", restResponse, grpcResponse)
	}
	if restResponse.RequestID != request.RequestID || restResponse.Payload != request.Payload {
		t.Fatalf("response changed request semantics: %#v", restResponse)
	}
}

func TestRESTAndGRPCRejectMissingRequestID(t *testing.T) {
	logic := internal.NewEchoLogic()
	restServer := httptest.NewServer(internal.NewRESTHandler(logic))
	defer restServer.Close()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterBenchmarkServiceServer(grpcServer, internal.NewGRPCServer(logic))
	go func() { _ = grpcServer.Serve(listener) }()
	defer grpcServer.Stop()

	restResponse, err := http.Post(restServer.URL+"/api/v1/echo", "application/json", bytes.NewBufferString(`{"payload":"same"}`))
	if err != nil {
		t.Fatal(err)
	}
	defer restResponse.Body.Close()
	if restResponse.StatusCode != http.StatusBadRequest {
		t.Fatalf("REST status = %d, want 400", restResponse.StatusCode)
	}

	client, err := internal.NewGRPCClient(listener.Addr().String(), 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	_, err = client.Echo(context.Background(), internal.EchoRequest{Payload: "same"})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("gRPC status = %s, want InvalidArgument", status.Code(err))
	}
}
