package main

import (
	"log"
	"net/http"
	"os"
	"time"
)

const (
	listenAddr           = ":8080"
	defaultIrmaServerURL = "http://localhost:8088"
	readHeaderTimeout    = 10 * time.Second
)

func main() {
	irma := newIrmaClient(envOr("IRMA_SERVER_URL", defaultIrmaServerURL), os.Getenv("API_TOKEN"))

	srv := &http.Server{
		Addr:              listenAddr,
		Handler:           newHandler(irma),
		ReadHeaderTimeout: readHeaderTimeout,
	}

	log.Printf("listening on %s", listenAddr)
	log.Fatal(srv.ListenAndServe())
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
