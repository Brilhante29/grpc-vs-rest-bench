package internal

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"
	"testing"

	pb "github.com/Brilhante29/grpc-vs-rest-bench/internal/proto/benchmark/v1"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestEchoLogic(t *testing.T) {
	l := NewEchoLogic()
	payload := strings.Repeat("A", 256)
	resp, err := l.Echo(context.Background(), EchoRequest{RequestID: "request-1", Payload: payload})
	if err != nil {
		t.Fatal(err)
	}
	if resp.RequestID != "request-1" || resp.Payload != payload {
		t.Fatalf("unexpected response: %#v", resp)
	}
	digest := sha256.Sum256([]byte(payload))
	if resp.PayloadSHA256 != fmt.Sprintf("%x", digest) {
		t.Fatalf("unexpected payload digest: %s", resp.PayloadSHA256)
	}
}

func TestGeneratedProtobufMatchesEchoContract(t *testing.T) {
	request := pb.File_proto_benchmark_v1_service_proto.Messages().ByName(protoreflect.Name("EchoRequest"))
	response := pb.File_proto_benchmark_v1_service_proto.Messages().ByName(protoreflect.Name("EchoResponse"))
	if request == nil || request.Fields().ByName("request_id") == nil || request.Fields().ByName("payload") == nil {
		t.Fatal("generated EchoRequest does not match the versioned proto contract")
	}
	if response == nil || response.Fields().ByName("request_id") == nil || response.Fields().ByName("payload") == nil || response.Fields().ByName("payload_sha256") == nil {
		t.Fatal("generated EchoResponse does not match the versioned proto contract")
	}
}

func TestEchoLogicRequiresRequestID(t *testing.T) {
	l := NewEchoLogic()
	_, err := l.Echo(context.Background(), EchoRequest{Payload: "payload"})
	if !errors.Is(err, ErrRequestIDRequired) {
		t.Fatalf("expected ErrRequestIDRequired, got %v", err)
	}
}
