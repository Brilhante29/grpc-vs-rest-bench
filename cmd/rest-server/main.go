package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Brilhante29/grpc-vs-rest-bench/internal"
	pb "github.com/Brilhante29/grpc-vs-rest-bench/internal/proto/benchmark/v1"
	"github.com/go-chi/chi/v5"
)

type echoRequest struct {
	Message      string `json:"message"`
	PayloadBytes int32  `json:"payload_bytes"`
}

type echoResponse struct {
	Message         string `json:"message"`
	ServerTimestamp int64  `json:"server_timestamp"`
}

var logic = internal.NewEchoLogic()

func echoHandler(w http.ResponseWriter, r *http.Request) {
	var req echoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	pbReq := &pb.EchoRequest{
		Message:      req.Message,
		PayloadBytes: req.PayloadBytes,
	}
	pbResp := logic.Echo(pbReq)
	resp := echoResponse{
		Message:         pbResp.Message,
		ServerTimestamp: pbResp.ServerTimestamp,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r := chi.NewRouter()
	r.Post("/api/v1/echo", echoHandler)
	log.Printf("REST server listening on :%s", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), r); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
