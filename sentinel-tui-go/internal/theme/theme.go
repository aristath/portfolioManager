package theme

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	Name       string
	Primary    lipgloss.Color
	Secondary  lipgloss.Color
	Background lipgloss.Color
	Surface    lipgloss.Color
	Success    lipgloss.Color
	Error      lipgloss.Color
	Warning    lipgloss.Color
	Text       lipgloss.Color
}

var Themes = []Theme{
	{
		Name:       "Synthwave Sunset",
		Primary:    lipgloss.Color("#ff00ff"),
		Secondary:  lipgloss.Color("#00ffff"),
		Background: lipgloss.Color("#1a0b2e"),
		Surface:    lipgloss.Color("#2d1b4e"),
		Success:    lipgloss.Color("#00ff88"),
		Error:      lipgloss.Color("#ff4444"),
		Warning:    lipgloss.Color("#ffaa00"),
		Text:       lipgloss.Color("#ffffff"),
	},
	{
		Name:       "Cyberpunk Matrix",
		Primary:    lipgloss.Color("#00ff41"),
		Secondary:  lipgloss.Color("#ff00ff"),
		Background: lipgloss.Color("#0a0a0a"),
		Surface:    lipgloss.Color("#1a1a1a"),
		Success:    lipgloss.Color("#00ff41"),
		Error:      lipgloss.Color("#ff4444"),
		Warning:    lipgloss.Color("#ffaa00"),
		Text:       lipgloss.Color("#ffffff"),
	},
	{
		Name:       "Neon Tron",
		Primary:    lipgloss.Color("#00f3ff"),
		Secondary:  lipgloss.Color("#ff9d00"),
		Background: lipgloss.Color("#0d1b2a"),
		Surface:    lipgloss.Color("#1b2d45"),
		Success:    lipgloss.Color("#00ff88"),
		Error:      lipgloss.Color("#ff4444"),
		Warning:    lipgloss.Color("#ff9d00"),
		Text:       lipgloss.Color("#ffffff"),
	},
	{
		Name:       "Cyberpunk Night",
		Primary:    lipgloss.Color("#ff0099"),
		Secondary:  lipgloss.Color("#00d4ff"),
		Background: lipgloss.Color("#1a1a2e"),
		Surface:    lipgloss.Color("#2a2a4e"),
		Success:    lipgloss.Color("#00ff88"),
		Error:      lipgloss.Color("#ff4444"),
		Warning:    lipgloss.Color("#ffaa00"),
		Text:       lipgloss.Color("#ffffff"),
	},
}
