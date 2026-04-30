package collect

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/mehaig/mistly-ingestor/internal/db"
	"github.com/mehaig/mistly-ingestor/internal/ua"
)

type request struct {
	SiteID       string `json:"site_id"`
	URL          string `json:"url"`
	Referrer     string `json:"referrer"`
	PageTitle    string `json:"page_title"`
	ScreenWidth  int    `json:"screen_width"`
	ScreenHeight int    `json:"screen_height"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if origin == "" {
		origin = "*"
	}
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	if req.SiteID == "" || req.URL == "" {
		http.Error(w, "site_id and url are required", http.StatusBadRequest)
		return
	}

	exists, err := db.SiteExists(req.SiteID)
	if err != nil {
		log.Printf("check site: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if !exists {
		http.Error(w, "unknown site", http.StatusForbidden)
		return
	}

	userAgent := r.Header.Get("User-Agent")
	ip := clientIP(r)
	info := ua.Parse(userAgent)

	date := time.Now().UTC().Format("2006-01-02")
	sum := sha256.Sum256([]byte(ip + "|" + userAgent + "|" + req.SiteID + "|" + date))
	sessionID := hex.EncodeToString(sum[:])

	utmSource, utmMedium, utmCampaign := parseUTM(req.URL)

	pv := db.Pageview{
		SiteID:       req.SiteID,
		URL:          req.URL,
		Referrer:     req.Referrer,
		Browser:      info.Browser,
		OS:           info.OS,
		Device:       info.Device,
		SessionID:    sessionID,
		Language:     parseLanguage(r.Header.Get("Accept-Language")),
		PageTitle:    req.PageTitle,
		ScreenWidth:  req.ScreenWidth,
		ScreenHeight: req.ScreenHeight,
		UTMSource:    utmSource,
		UTMMedium:    utmMedium,
		UTMCampaign:  utmCampaign,
	}

	if err := db.InsertPageview(pv); err != nil {
		log.Printf("insert pageview: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if comma := strings.IndexByte(xff, ','); comma >= 0 {
			return strings.TrimSpace(xff[:comma])
		}
		return strings.TrimSpace(xff)
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// parseLanguage extracts the primary language tag from an Accept-Language header.
// "en-US,en;q=0.9,ro;q=0.8" → "en-US"
func parseLanguage(header string) string {
	if header == "" {
		return ""
	}
	lang := strings.SplitN(header, ",", 2)[0]
	lang = strings.SplitN(lang, ";", 2)[0]
	return strings.TrimSpace(lang)
}

func parseUTM(rawURL string) (source, medium, campaign string) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return
	}
	q := u.Query()
	return q.Get("utm_source"), q.Get("utm_medium"), q.Get("utm_campaign")
}
