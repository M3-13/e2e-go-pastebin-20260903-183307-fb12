package main

import (
	"log"
	"net/http"
	"os"

	"pastebin/internal/api"
	"pastebin/internal/paste"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	s := paste.NewStore()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", api.Health)
	api.RegisterCreateRoutes(mux, s)
	api.RegisterReadRoutes(mux, s)
	api.RegisterDeleteRoutes(mux, s)

	handler := api.JSONErrorHandler(mux)

	log.Printf("pastebin API listening on port %s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
