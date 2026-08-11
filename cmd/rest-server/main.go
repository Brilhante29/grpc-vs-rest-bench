package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Brilhante29/grpc-vs-rest-bench/internal"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	handler := internal.NewRESTHandler(internal.NewEchoLogic())
	log.Printf("REST server listening on :%s", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), handler); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
