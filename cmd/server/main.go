package main

import (
	_ "embed"
	"log"
	"net/http"
	"os"

	"github.com/mehaig/mistly-ingestor/internal/collect"
	"github.com/mehaig/mistly-ingestor/internal/db"
	"github.com/mehaig/mistly-ingestor/internal/sites"
)

//go:embed static/tracker.js
var trackerJS []byte

func main() {
	db.Connect()
	db.Migrate()

	mux := http.NewServeMux()
	mux.HandleFunc("/collect", collect.Handler)
	mux.HandleFunc("/tracker.js", serveTracker)
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("ok"))
	})

	mux.HandleFunc("POST /sites", sites.Create)
	mux.HandleFunc("GET /sites", sites.List)
	mux.HandleFunc("GET /sites/{id}/snippet", sites.Snippet)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Mistly listening on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}

func serveTracker(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.Write(trackerJS)
}
