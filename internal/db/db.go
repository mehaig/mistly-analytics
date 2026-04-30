package db

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

var conn *sql.DB

type Site struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Domain    string    `json:"domain"`
	CreatedAt time.Time `json:"created_at"`
}

type Pageview struct {
	SiteID       string
	URL          string
	Referrer     string
	Browser      string
	OS           string
	Device       string
	SessionID    string
	Language     string
	PageTitle    string
	ScreenWidth  int
	ScreenHeight int
	UTMSource    string
	UTMMedium    string
	UTMCampaign  string
}

const schema = `
CREATE TABLE IF NOT EXISTS sites (
    id         TEXT        PRIMARY KEY,
    name       TEXT        NOT NULL,
    domain     TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS pageviews (
    id            BIGSERIAL   PRIMARY KEY,
    site_id       TEXT        NOT NULL,
    url           TEXT        NOT NULL,
    referrer      TEXT,
    browser       TEXT,
    os            TEXT,
    device        TEXT,
    session_id    TEXT,
    language      TEXT,
    page_title    TEXT,
    screen_width  INT,
    screen_height INT,
    utm_source    TEXT,
    utm_medium    TEXT,
    utm_campaign  TEXT,
    created_at    TIMESTAMPTZ DEFAULT NOW()
);

ALTER TABLE pageviews ADD COLUMN IF NOT EXISTS language      TEXT;
ALTER TABLE pageviews ADD COLUMN IF NOT EXISTS page_title   TEXT;
ALTER TABLE pageviews ADD COLUMN IF NOT EXISTS screen_width  INT;
ALTER TABLE pageviews ADD COLUMN IF NOT EXISTS screen_height INT;
ALTER TABLE pageviews ADD COLUMN IF NOT EXISTS utm_source   TEXT;
ALTER TABLE pageviews ADD COLUMN IF NOT EXISTS utm_medium   TEXT;
ALTER TABLE pageviews ADD COLUMN IF NOT EXISTS utm_campaign TEXT;

CREATE INDEX IF NOT EXISTS idx_pageviews_site_id    ON pageviews(site_id);
CREATE INDEX IF NOT EXISTS idx_pageviews_created_at ON pageviews(created_at);
`

func NewSiteID() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func Connect() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("ping database: %v", err)
	}

	conn = db
}

func Migrate() {
	if _, err := conn.Exec(schema); err != nil {
		log.Fatalf("migrate: %v", err)
	}
}

func CreateSite(id, name, domain string) (Site, error) {
	var s Site
	err := conn.QueryRow(
		`INSERT INTO sites (id, name, domain) VALUES ($1, $2, NULLIF($3, ''))
		 RETURNING id, name, COALESCE(domain, ''), created_at`,
		id, name, domain,
	).Scan(&s.ID, &s.Name, &s.Domain, &s.CreatedAt)
	return s, err
}

func ListSites() ([]Site, error) {
	rows, err := conn.Query(
		`SELECT id, name, COALESCE(domain, ''), created_at FROM sites ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sites := []Site{}
	for rows.Next() {
		var s Site
		if err := rows.Scan(&s.ID, &s.Name, &s.Domain, &s.CreatedAt); err != nil {
			return nil, err
		}
		sites = append(sites, s)
	}
	return sites, rows.Err()
}

func GetSite(id string) (Site, error) {
	var s Site
	err := conn.QueryRow(
		`SELECT id, name, COALESCE(domain, ''), created_at FROM sites WHERE id = $1`,
		id,
	).Scan(&s.ID, &s.Name, &s.Domain, &s.CreatedAt)
	return s, err
}

func SiteExists(id string) (bool, error) {
	var exists bool
	err := conn.QueryRow(`SELECT EXISTS(SELECT 1 FROM sites WHERE id = $1)`, id).Scan(&exists)
	return exists, err
}

func InsertPageview(p Pageview) error {
	_, err := conn.Exec(
		`INSERT INTO pageviews
		    (site_id, url, referrer, browser, os, device, session_id,
		     language, page_title, screen_width, screen_height,
		     utm_source, utm_medium, utm_campaign)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9,
		         NULLIF($10, 0), NULLIF($11, 0),
		         $12, $13, $14)`,
		p.SiteID, p.URL, p.Referrer, p.Browser, p.OS, p.Device, p.SessionID,
		p.Language, p.PageTitle, p.ScreenWidth, p.ScreenHeight,
		p.UTMSource, p.UTMMedium, p.UTMCampaign,
	)
	return err
}
