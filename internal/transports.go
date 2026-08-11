package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	pb "github.com/Brilhante29/grpc-vs-rest-bench/internal/proto/benchmark/v1"
	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type restRequest struct {
	RequestID string `json:"request_id"`
	Payload   string `json:"payload"`
}

type restResponse struct {
	RequestID     string `json:"request_id"`
	Payload       string `json:"payload"`
	PayloadSHA256 string `json:"payload_sha256"`
}

func NewRESTHandler(logic Echoer) http.Handler {
	router := chi.NewRouter()
	router.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })
	router.Post("/api/v1/echo", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var req restRequest
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<20)).Decode(&req); err != nil {
			http.Error(w, "invalid JSON request", http.StatusBadRequest)
			return
		}
		resp, err := logic.Echo(r.Context(), EchoRequest{RequestID: req.RequestID, Payload: req.Payload})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(restResponse(resp))
	})
	return router
}

type RESTClient struct {
	baseURL string
	client  *http.Client
}

func NewRESTClient(address string, concurrency int, timeout time.Duration) *RESTClient {
	transport := &http.Transport{
		MaxIdleConns: concurrency, MaxIdleConnsPerHost: concurrency,
		MaxConnsPerHost: concurrency, IdleConnTimeout: 90 * time.Second,
	}
	return &RESTClient{
		baseURL: "http://" + strings.TrimSuffix(address, "/"),
		client:  &http.Client{Transport: transport, Timeout: timeout},
	}
}

func (c *RESTClient) Echo(ctx context.Context, request EchoRequest) (EchoResponse, error) {
	body, err := json.Marshal(restRequest(request))
	if err != nil {
		return EchoResponse{}, fmt.Errorf("marshal REST request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/echo", bytes.NewReader(body))
	if err != nil {
		return EchoResponse{}, fmt.Errorf("create REST request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return EchoResponse{}, fmt.Errorf("REST request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return EchoResponse{}, fmt.Errorf("REST status: %s", resp.Status)
	}
	var decoded restResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return EchoResponse{}, fmt.Errorf("decode REST response: %w", err)
	}
	return EchoResponse(decoded), nil
}

type GRPCServer struct {
	pb.UnimplementedBenchmarkServiceServer
	logic Echoer
}

func NewGRPCServer(logic Echoer) *GRPCServer { return &GRPCServer{logic: logic} }

func (s *GRPCServer) Echo(ctx context.Context, req *pb.EchoRequest) (*pb.EchoResponse, error) {
	resp, err := s.logic.Echo(ctx, EchoRequest{RequestID: req.GetRequestId(), Payload: req.GetPayload()})
	if err != nil {
		return nil, err
	}
	return &pb.EchoResponse{RequestId: resp.RequestID, Payload: resp.Payload, PayloadSha256: resp.PayloadSHA256}, nil
}

type GRPCClient struct {
	conn   *grpc.ClientConn
	client pb.BenchmarkServiceClient
}

func NewGRPCClient(address string, timeout time.Duration) (*GRPCClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	conn, err := grpc.DialContext(ctx, address, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		return nil, fmt.Errorf("dial gRPC server: %w", err)
	}
	return &GRPCClient{conn: conn, client: pb.NewBenchmarkServiceClient(conn)}, nil
}

func (c *GRPCClient) Echo(ctx context.Context, request EchoRequest) (EchoResponse, error) {
	resp, err := c.client.Echo(ctx, &pb.EchoRequest{RequestId: request.RequestID, Payload: request.Payload})
	if err != nil {
		return EchoResponse{}, err
	}
	return EchoResponse{RequestID: resp.GetRequestId(), Payload: resp.GetPayload(), PayloadSHA256: resp.GetPayloadSha256()}, nil
}

func (c *GRPCClient) Close() error { return c.conn.Close() }
