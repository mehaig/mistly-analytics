package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	"github.com/mehaig/mistly-ingestor/internal/db"
)

func main() {
	log.SetFlags(0)

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "init":
		cmdInit(os.Args[2:])
	case "sites":
		cmdSites(os.Args[2:])
	case "token":
		cmdToken()
	default:
		printUsage()
		os.Exit(1)
	}
}

// cmdInit connects, migrates, creates the first site, and prints the snippet.
func cmdInit(args []string) {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	name := fs.String("name", "", "site name (required)")
	domain := fs.String("domain", "", "site domain (optional)")
	serverURL := fs.String("server-url", "http://localhost:8080", "public URL of your Mistly server")
	fs.Usage = func() {
		fmt.Println("Usage: mistly init --name <name> [--domain <domain>] [--server-url <url>]")
		fs.PrintDefaults()
	}
	fs.Parse(args)

	if *name == "" {
		fmt.Fprintln(os.Stderr, "error: --name is required")
		os.Exit(1)
	}

	db.Connect()
	fmt.Println("Connected to database")

	db.Migrate()
	fmt.Println("Migrations complete")

	id, err := db.NewSiteID()
	if err != nil {
		log.Fatalf("generate site ID: %v", err)
	}

	site, err := db.CreateSite(id, *name, *domain)
	if err != nil {
		log.Fatalf("create site: %v", err)
	}

	fmt.Printf("\nSite created\n")
	fmt.Printf("  ID:   %s\n", site.ID)
	fmt.Printf("  Name: %s\n", site.Name)
	if site.Domain != "" {
		fmt.Printf("  Domain: %s\n", site.Domain)
	}

	fmt.Printf("\nAdd this to your website <head>:\n\n")
	fmt.Printf("  <script src=\"%s/tracker.js\" data-site-id=\"%s\" defer></script>\n\n", *serverURL, site.ID)
}

func cmdSites(args []string) {
	if len(args) == 0 {
		printSitesUsage()
		os.Exit(1)
	}

	switch args[0] {
	case "create":
		cmdSitesCreate(args[1:])
	case "list":
		cmdSitesList()
	case "snippet":
		cmdSitesSnippet(args[1:])
	default:
		printSitesUsage()
		os.Exit(1)
	}
}

func cmdSitesCreate(args []string) {
	fs := flag.NewFlagSet("sites create", flag.ExitOnError)
	name := fs.String("name", "", "site name (required)")
	domain := fs.String("domain", "", "site domain (optional)")
	serverURL := fs.String("server-url", "http://localhost:8080", "public URL of your Mistly server")
	fs.Usage = func() {
		fmt.Println("Usage: mistly sites create --name <name> [--domain <domain>] [--server-url <url>]")
		fs.PrintDefaults()
	}
	fs.Parse(args)

	if *name == "" {
		fmt.Fprintln(os.Stderr, "error: --name is required")
		os.Exit(1)
	}

	db.Connect()

	id, err := db.NewSiteID()
	if err != nil {
		log.Fatalf("generate site ID: %v", err)
	}

	site, err := db.CreateSite(id, *name, *domain)
	if err != nil {
		log.Fatalf("create site: %v", err)
	}

	fmt.Printf("ID:   %s\n", site.ID)
	fmt.Printf("Name: %s\n", site.Name)
	if site.Domain != "" {
		fmt.Printf("Domain: %s\n", site.Domain)
	}

	fmt.Printf("\nSnippet:\n\n")
	fmt.Printf("  <script src=\"%s/tracker.js\" data-site-id=\"%s\" defer></script>\n\n", *serverURL, site.ID)
}

func cmdSitesList() {
	db.Connect()

	sites, err := db.ListSites()
	if err != nil {
		log.Fatalf("list sites: %v", err)
	}

	if len(sites) == 0 {
		fmt.Println("No sites. Run `mistly init` or `mistly sites create` to add one.")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tDOMAIN\tCREATED")
	for _, s := range sites {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", s.ID, s.Name, s.Domain, s.CreatedAt.Format("2006-01-02"))
	}
	w.Flush()
}

func cmdSitesSnippet(args []string) {
	fs := flag.NewFlagSet("sites snippet", flag.ExitOnError)
	serverURL := fs.String("server-url", "http://localhost:8080", "public URL of your Mistly server")
	fs.Usage = func() {
		fmt.Println("Usage: mistly sites snippet <id> [--server-url <url>]")
		fs.PrintDefaults()
	}
	fs.Parse(args)

	if fs.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "error: site ID is required")
		fmt.Fprintln(os.Stderr, "usage: mistly sites snippet <id> [--server-url <url>]")
		os.Exit(1)
	}

	db.Connect()

	site, err := db.GetSite(fs.Arg(0))
	if err == sql.ErrNoRows {
		fmt.Fprintf(os.Stderr, "error: site %q not found\n", fs.Arg(0))
		os.Exit(1)
	}
	if err != nil {
		log.Fatalf("get site: %v", err)
	}

	fmt.Printf("<script src=\"%s/tracker.js\" data-site-id=\"%s\" defer></script>\n", *serverURL, site.ID)
}

func cmdToken() {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		log.Fatalf("generate token: %v", err)
	}
	fmt.Println(hex.EncodeToString(b))
}

func printUsage() {
	fmt.Print(`Usage: mistly <command> [flags]

Commands:
  init          Connect to the database, run migrations, and create your first site
  sites         Manage sites
  token         Generate a secure random token for ADMIN_TOKEN

Environment:
  DATABASE_URL  Postgres connection string (required)

Run mistly <command> --help for command-specific flags.
`)
}

func printSitesUsage() {
	fmt.Print(`Usage: mistly sites <subcommand> [flags]

Subcommands:
  create        Create a new site and print its tracker snippet
  list          List all sites
  snippet       Print the tracker snippet for an existing site
`)
}
