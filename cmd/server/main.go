package main

import (
	_ "embed"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/mehaig/mistly-ingestor/internal/collect"
	"github.com/mehaig/mistly-ingestor/internal/db"
	"github.com/mehaig/mistly-ingestor/internal/sites"
)

//go:embed static/tracker.js
var trackerJS []byte

func main() {
	db.Connect()
	db.Migrate()

	if os.Getenv("ADMIN_TOKEN") == "" {
		log.Println("warning: ADMIN_TOKEN not set — /sites endpoints are unprotected")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/collect", collect.Handler)
	mux.HandleFunc("/tracker.js", serveTracker)
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("ok"))
	})

	mux.HandleFunc("POST /sites", adminOnly(sites.Create))
	mux.HandleFunc("GET /sites", adminOnly(sites.List))
	mux.HandleFunc("GET /sites/{id}/snippet", adminOnly(sites.Snippet))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Mistly listening on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}

func adminOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := os.Getenv("ADMIN_TOKEN")
		if token == "" {
			next(w, r)
			return
		}
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") || strings.TrimPrefix(auth, "Bearer ") != token {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

func serveTracker(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.Write(trackerJS)
}
