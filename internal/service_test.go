package internal

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"
	"testing"
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

func TestEchoLogicRequiresRequestID(t *testing.T) {
	l := NewEchoLogic()
	_, err := l.Echo(context.Background(), EchoRequest{Payload: "payload"})
	if !errors.Is(err, ErrRequestIDRequired) {
		t.Fatalf("expected ErrRequestIDRequired, got %v", err)
	}
}
