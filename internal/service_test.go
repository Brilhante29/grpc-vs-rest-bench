package internal

import (
	"strings"
	"testing"

	pb "github.com/Brilhante29/grpc-vs-rest-bench/internal/proto/benchmark/v1"
)

func TestEchoLogic(t *testing.T) {
	l := NewEchoLogic()
	resp := l.Echo(&pb.EchoRequest{Message: "test", PayloadBytes: 4})
	if resp.Message != "testAAAA" {
		t.Fatalf("unexpected message: %q", resp.Message)
	}
	if resp.ServerTimestamp == 0 {
		t.Fatal("expected non-zero timestamp")
	}
}

func TestEchoLogicPadding(t *testing.T) {
	l := NewEchoLogic()
	resp := l.Echo(&pb.EchoRequest{Message: "ping", PayloadBytes: 256})
	expected := "ping" + strings.Repeat("A", 256)
	if resp.Message != expected {
		t.Fatalf("expected message length %d, got %d", len(expected), len(resp.Message))
	}
}

func TestEchoLogicEmpty(t *testing.T) {
	l := NewEchoLogic()
	resp := l.Echo(&pb.EchoRequest{Message: "", PayloadBytes: 0})
	if resp.Message != "" {
		t.Fatalf("expected empty, got %q", resp.Message)
	}
}
