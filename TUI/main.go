package main

import (
	"flag"
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"

	"sentinel-tui-go/internal/api"
	"sentinel-tui-go/internal/config"
	"sentinel-tui-go/internal/ui"
)

func main() {
	apiURL := flag.String("api-url", "http://localhost:8000", "Sentinel API URL")
	settingsFile := flag.String("settings-file", "settings.json", "Path to TUI settings JSON")
	maxWidth := flag.Int("max-width", 0, "Max columns (0 = no limit)")
	maxHeight := flag.Int("max-height", 0, "Max rows (0 = no limit)")
	flag.Parse()

	effectiveAPIURL := *apiURL
	if cfg, err := config.Load(*settingsFile); err == nil && cfg.APIURL != "" {
		effectiveAPIURL = cfg.APIURL
	}

	client := api.NewClient(effectiveAPIURL)
	m := ui.NewModel(client, effectiveAPIURL, *settingsFile, *maxWidth, *maxHeight)

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
