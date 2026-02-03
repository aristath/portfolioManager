package theme

import (
	"fmt"
	"image/color"
	"math"
	"strings"

	"charm.land/lipgloss/v2"
)

// Theme holds the semantic color palette for the entire TUI.
type Theme struct {
	Base    color.Color
	Surface color.Color
	Overlay color.Color
	Border  color.Color
	Muted   color.Color
	Text    color.Color
	Subtext color.Color
	Primary color.Color
	Accent  color.Color
	Success color.Color
	Warning color.Color
	Error   color.Color
	Info    color.Color
}

// Default theme uses Charmbracelet's CharmTone palette from Crush.
var Default = Theme{
	Base:    lipgloss.Color("#201F26"), // Pepper
	Surface: lipgloss.Color("#2D2C35"), // BBQ
	Overlay: lipgloss.Color("#3A3943"), // Charcoal
	Border:  lipgloss.Color("#4D4C57"), // Iron
	Muted:   lipgloss.Color("#858392"), // Squid
	Text:    lipgloss.Color("#DFDBDD"), // Ash
	Subtext: lipgloss.Color("#BFBCC8"), // Smoke
	Primary: lipgloss.Color("#6B50FF"), // Charple
	Accent:  lipgloss.Color("#FF60FF"), // Dolly
	Success: lipgloss.Color("#00FFB2"), // Julep
	Warning: lipgloss.Color("#FFD300"),
	Error:   lipgloss.Color("#E94090"),
	Info:    lipgloss.Color("#00CED1"),
}

// GradientText applies a horizontal color gradient across each line of text.
func GradientText(text string, from, to color.Color) string {
	fr, fg, fb := colorToRGB(from)
	tr, tg, tb := colorToRGB(to)

	lines := strings.Split(text, "\n")
	var result []string

	for _, line := range lines {
		runes := []rune(line)
		n := len(runes)
		if n == 0 {
			result = append(result, "")
			continue
		}

		var sb strings.Builder
		for i, r := range runes {
			t := 0.0
			if n > 1 {
				t = float64(i) / float64(n-1)
			}
			cr := uint8(math.Round(float64(fr) + t*float64(int(tr)-int(fr))))
			cg := uint8(math.Round(float64(fg) + t*float64(int(tg)-int(fg))))
			cb := uint8(math.Round(float64(fb) + t*float64(int(tb)-int(fb))))

			c := lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", cr, cg, cb))
			sb.WriteString(lipgloss.NewStyle().Foreground(c).Render(string(r)))
		}
		result = append(result, sb.String())
	}
	return strings.Join(result, "\n")
}

func colorToRGB(c color.Color) (uint8, uint8, uint8) {
	r, g, b, _ := c.RGBA()
	return uint8(r >> 8), uint8(g >> 8), uint8(b >> 8)
}
