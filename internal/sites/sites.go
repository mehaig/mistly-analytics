package sites

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mehaig/mistly-ingestor/internal/db"
)

type createRequest struct {
	Name   string `json:"name"`
	Domain string `json:"domain"`
}

func Create(w http.ResponseWriter, r *http.Request) {
	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	id, err := db.NewSiteID()
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	site, err := db.CreateSite(id, req.Name, req.Domain)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(site)
}

func List(w http.ResponseWriter, r *http.Request) {
	sites, err := db.ListSites()
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sites)
}

func Snippet(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	site, err := db.GetSite(id)
	if err == sql.ErrNoRows {
		http.Error(w, "site not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	base := serverURL(r)
	snippet := fmt.Sprintf(
		`<script src="%s/tracker.js" data-site-id="%s" defer></script>`,
		base, site.ID,
	)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, snippet)
}

func serverURL(r *http.Request) string {
	scheme := "https"
	if r.TLS == nil && r.Header.Get("X-Forwarded-Proto") != "https" {
		scheme = "http"
	}
	return scheme + "://" + r.Host
}

