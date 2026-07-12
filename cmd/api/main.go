package main

import (
	"log"
	"net/http"
	"os"
	"risk-api/internal/handlers"
	"risk-api/internal/store"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Initializing embedded in-memory database store...")

	s, err := store.NewMemoryStore()
	if err != nil {
		log.Fatalf("Failed to initialize memory store: %v", err)
	}

	mux := http.NewServeMux()
	h := handlers.NewHandler(s)
	h.RegisterRoutes(mux)

	// Request logger middleware
	loggingMux := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		mux.ServeHTTP(w, r)
	})

	log.Printf("Server listening on port %s...", port)
	if err := http.ListenAndServe(":"+port, loggingMux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
