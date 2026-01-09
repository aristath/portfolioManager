package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// All documentation section IDs from the Tradernet API website
var docSections = map[string][]string{
	"Authentication": {
		"auth-login",
		"auth-api",
		"auth-get-opq",
		"auth-get-sidinfo",
		"public-api-client",
		"python-sdk",
	},
	"Security Sessions": {
		"security-get-list",
		"open-security-session",
	},
	"Securities Management": {
		"quotes-get-lists",
		"quotes-add-list",
		"quotes-update-list",
		"quotes-delete-list",
		"quotes-make-list-selected",
		"quotes-add-list-ticker",
		"quotes-delete-list-ticker",
	},
	"Quotes & Market Data": {
		"market-status",
		"quotes-get-info",
		"get-options-by-mkt",
		"quotes-get-top-securities",
		"quotes-get-changes",
		"quotes-get",
		"quotes-orderbook",
		"quotes-get-hloc",
		"get-trades",
		"get-trades-history",
		"quotes-finder",
		"quotes-get-news",
		"securities",
		"check-allowed-ticker-and-ban-on-trade",
	},
	"Portfolio & Orders": {
		"portfolio-get-changes",
		"orders-get-current-history",
		"get-orders-history",
		"orders-send",
		"stop-loss",
		"orders-delete",
	},
	"Alerts & Requests": {
		"alerts-get-list",
		"alerts-add",
		"alerts-delete",
		"get-client-cps-history",
		"get-cps-files",
	},
	"Reports": {
		"broker-report",
		"broker-report-url",
		"depositary-report",
		"broker-depositary-report-url",
		"get-cashflows",
	},
	"Currencies & WebSocket": {
		"cross-rates-for-date",
		"currency",
		"websocket",
		"websocket-sessions",
		"websocket-portfolio",
		"websocket-orders",
		"websocket-markets",
	},
	"Miscellaneous": {
		"reception-types",
		"special-files-list",
		"mkt",
		"instruments",
		"cps-types-list",
		"anketa-fields",
		"passport-type",
		"order-statuses",
		"safety",
		"type-codes",
	},
}

const (
	baseURL = "https://tradernet.com/tradernet-api/get-example?id="
	docsDir = "./internal/clients/tradernet/docs"
)

func main() {
	// Create docs directory
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		fmt.Printf("Error creating docs directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Scraping Tradernet API Documentation...")
	fmt.Println("========================================\n")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	totalDocs := 0
	successCount := 0
	failedDocs := []string{}

	// Fetch each documentation page
	for category, sections := range docSections {
		fmt.Printf("Category: %s\n", category)

		// Create category directory
		categoryDir := filepath.Join(docsDir, sanitizeFilename(category))
		if err := os.MkdirAll(categoryDir, 0755); err != nil {
			fmt.Printf("  Error creating category directory: %v\n", err)
			continue
		}

		for _, sectionID := range sections {
			totalDocs++
			url := baseURL + sectionID
			filename := filepath.Join(categoryDir, sectionID+".html")

			fmt.Printf("  Fetching: %s... ", sectionID)

			// Create request with proper headers
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				fmt.Printf("FAILED (request error: %v)\n", err)
				failedDocs = append(failedDocs, sectionID)
				continue
			}

			// Add headers to mimic a browser
			req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
			req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
			req.Header.Set("Accept-Language", "en-US,en;q=0.9")
			req.Header.Set("Referer", "https://tradernet.com/tradernet-api/")

			// Fetch the documentation page
			resp, err := client.Do(req)
			if err != nil {
				fmt.Printf("FAILED (error: %v)\n", err)
				failedDocs = append(failedDocs, sectionID)
				continue
			}

			if resp.StatusCode != http.StatusOK {
				fmt.Printf("FAILED (status: %d)\n", resp.StatusCode)
				resp.Body.Close()
				failedDocs = append(failedDocs, sectionID)
				continue
			}

			// Read response body
			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				fmt.Printf("FAILED (read error: %v)\n", err)
				failedDocs = append(failedDocs, sectionID)
				continue
			}

			// Save to file
			if err := os.WriteFile(filename, body, 0644); err != nil {
				fmt.Printf("FAILED (write error: %v)\n", err)
				failedDocs = append(failedDocs, sectionID)
				continue
			}

			successCount++
			fmt.Println("OK")

			// Be nice to the server
			time.Sleep(500 * time.Millisecond)
		}
		fmt.Println()
	}

	// Create README index
	if err := createREADME(); err != nil {
		fmt.Printf("Error creating README: %v\n", err)
	}

	// Print summary
	fmt.Println("========================================")
	fmt.Printf("Total documents: %d\n", totalDocs)
	fmt.Printf("Successfully fetched: %d\n", successCount)
	fmt.Printf("Failed: %d\n", len(failedDocs))

	if len(failedDocs) > 0 {
		fmt.Println("\nFailed documents:")
		for _, doc := range failedDocs {
			fmt.Printf("  - %s\n", doc)
		}
	}

	fmt.Println("\nDocumentation saved to:", docsDir)
}

func sanitizeFilename(s string) string {
	// Replace spaces and special characters with underscores
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "&", "and")
	s = strings.ToLower(s)
	return s
}

func createREADME() error {
	readme := `# Tradernet API Documentation

This directory contains the complete Tradernet API documentation scraped from https://tradernet.com/tradernet-api/

## Documentation Structure

The documentation is organized into the following categories:

`

	// Add categories and their sections
	categoryOrder := []string{
		"Authentication",
		"Security Sessions",
		"Securities Management",
		"Quotes & Market Data",
		"Portfolio & Orders",
		"Alerts & Requests",
		"Reports",
		"Currencies & WebSocket",
		"Miscellaneous",
	}

	for _, category := range categoryOrder {
		sections := docSections[category]
		readme += fmt.Sprintf("### %s\n\n", category)

		categoryDir := sanitizeFilename(category)
		for _, sectionID := range sections {
			// Convert ID to readable title
			title := strings.ReplaceAll(sectionID, "-", " ")
			title = strings.Title(title)
			readme += fmt.Sprintf("- [%s](./%s/%s.html)\n", title, categoryDir, sectionID)
		}
		readme += "\n"
	}

	readme += `## Scraping

The documentation was scraped using the script at ` + "`scripts/scrape_tradernet_docs.go`" + `.

To re-scrape the documentation:

` + "```bash\ncd scripts\ngo run scrape_tradernet_docs.go\n```" + `

## Source

All documentation content is copyright Tradernet and sourced from their official API documentation website.

Last updated: ` + time.Now().Format("2006-01-02 15:04:05") + `
`

	readmePath := filepath.Join(docsDir, "README.md")
	return os.WriteFile(readmePath, []byte(readme), 0644)
}
