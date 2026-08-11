package internal

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
)

const ContractVersion = "echo.v1"

var ErrRequestIDRequired = errors.New("request_id is required")

type EchoRequest struct {
	RequestID string
	Payload   string
}

type EchoResponse struct {
	RequestID     string
	Payload       string
	PayloadSHA256 string
}

type Echoer interface {
	Echo(context.Context, EchoRequest) (EchoResponse, error)
}

type EchoLogic struct{}

func NewEchoLogic() *EchoLogic {
	return &EchoLogic{}
}

func (l *EchoLogic) Echo(_ context.Context, req EchoRequest) (EchoResponse, error) {
	if req.RequestID == "" {
		return EchoResponse{}, ErrRequestIDRequired
	}
	digest := sha256.Sum256([]byte(req.Payload))
	return EchoResponse{
		RequestID:     req.RequestID,
		Payload:       req.Payload,
		PayloadSHA256: fmt.Sprintf("%x", digest),
	}, nil
}
