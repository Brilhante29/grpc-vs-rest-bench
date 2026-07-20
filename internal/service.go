package internal

import (
	"strings"
	"time"

	pb "github.com/Brilhante29/grpc-vs-rest-bench/internal/proto/benchmark/v1"
)

type EchoLogic struct{}

func NewEchoLogic() *EchoLogic {
	return &EchoLogic{}
}

func (l *EchoLogic) Echo(req *pb.EchoRequest) *pb.EchoResponse {
	padding := ""
	if req.PayloadBytes > 0 {
		padding = strings.Repeat("A", int(req.PayloadBytes))
	}
	return &pb.EchoResponse{
		Message:         req.Message + padding,
		ServerTimestamp: time.Now().UnixNano(),
	}
}
