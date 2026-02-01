package ui

import (
	"fmt"
	"strings"

	figure "github.com/common-nighthawk/go-figure"

	"github.com/charmbracelet/lipgloss"

	"sentinel-tui-go/internal/theme"
)

func (m Model) View() string {
	if !m.ready {
		return "\n  Loading..."
	}
	t := theme.Themes[m.themeIndex]

	if m.showTable {
		return m.viewTable(t)
	}
	return m.viewMain(t)
}

func (m Model) viewMain(t theme.Theme) string {
	status := m.viewStatusBar(t)
	footer := m.viewFooter(t)

	page := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Background(t.Background)

	return page.Render(lipgloss.JoinVertical(lipgloss.Left,
		status,
		m.viewport.View(),
		footer,
	))
}

func (m *Model) rebuildContent() {
	t := theme.Themes[m.themeIndex]
	sections := []string{
		m.viewHero(t),
		m.viewMetrics(t),
	}
	if m.sparklineView != "" {
		sections = append(sections, m.sparklineView)
	}
	sections = append(sections, m.viewActions(t))
	m.viewport.SetContent(lipgloss.JoinVertical(lipgloss.Left, sections...))
}

// Status bar

func (m Model) viewStatusBar(t theme.Theme) string {
	bar := lipgloss.NewStyle().
		Width(m.width).
		Background(t.Surface).
		Foreground(t.Text).
		Padding(0, 1)

	dot := lipgloss.NewStyle().Foreground(t.Success).Render("\u25cf")
	status := "CONNECTED"
	if !m.connected {
		dot = lipgloss.NewStyle().Foreground(t.Error).Render("\u25cf")
		status = "DISCONNECTED"
	}

	mode := ""
	if m.tradingMode != "" {
		mode = "  \u2502  " + strings.ToUpper(m.tradingMode)
	}
	themeText := fmt.Sprintf("[%d/%d]", m.themeIndex+1, len(theme.Themes))

	return bar.Render(fmt.Sprintf(
		" %s SENTINEL  \u2502  %s%s  \u2502  %s",
		dot, status, mode, themeText,
	))
}

// Hero

func (m Model) viewHero(t theme.Theme) string {
	style := lipgloss.NewStyle().
		Foreground(t.Primary).
		Width(m.width).
		Align(lipgloss.Center).
		Padding(1, 2)

	if m.portfolio == nil && !m.connected {
		errStyle := lipgloss.NewStyle().
			Foreground(t.Error).
			Width(m.width).
			Align(lipgloss.Center).
			Padding(1, 2)
		return errStyle.Render(fmt.Sprintf("Cannot reach API at %s", m.apiURL))
	}

	value := m.heroValue
	if !m.animating && m.portfolio != nil {
		value = m.portfolio.TotalValueEUR
	}

	formatted := fmt.Sprintf("%.0f", value)
	if m.width < 60 {
		return style.Render(formatted)
	}

	fig := figure.NewFigure(formatted, "small", false)
	return style.Render(strings.Join(fig.Slicify(), "\n"))
}

// Metrics

func (m Model) viewMetrics(t theme.Theme) string {
	var pnlPct, cash, invested float64
	if m.portfolio != nil {
		cash = m.portfolio.TotalCashEUR
		invested = m.portfolio.TotalValueEUR - cash
	}
	if m.pnlHistory != nil {
		pnlPct = m.pnlHistory.Summary.PnLPercent
	}

	arrow := "\u25b2"
	pnlColor := t.Success
	if pnlPct < 0 {
		arrow = "\u25bc"
		pnlColor = t.Error
	}

	pnlStyle := lipgloss.NewStyle().Foreground(pnlColor)
	textStyle := lipgloss.NewStyle().Foreground(t.Text)

	line := fmt.Sprintf("%s  \u2502  %s  \u2502  %s",
		pnlStyle.Render(fmt.Sprintf("%s %+.2f%%", arrow, pnlPct)),
		textStyle.Render(fmt.Sprintf("Cash: %.0f", cash)),
		textStyle.Render(fmt.Sprintf("Invested: %.0f", invested)),
	)

	return lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center).
		Padding(0, 2).
		Render(line)
}

// Actions

func (m Model) viewActions(t theme.Theme) string {
	border := lipgloss.NewStyle().
		Width(m.width).
		BorderTop(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(t.Primary).
		Padding(1, 2)

	if len(m.recommendations) == 0 {
		return border.Render(
			lipgloss.NewStyle().Foreground(t.Secondary).Render("No recommendations"),
		)
	}

	headerStyle := lipgloss.NewStyle().Foreground(t.Primary)

	var lines []string
	lines = append(lines, headerStyle.Render("\u2500\u2500\u2500 NEXT ACTIONS \u2500\u2500\u2500"))
	lines = append(lines, "")

	var buyTotal, sellTotal float64

	limit := 3
	if limit > len(m.recommendations) {
		limit = len(m.recommendations)
	}

	for _, rec := range m.recommendations[:limit] {
		action := strings.ToUpper(rec.Action)
		cost := rec.Price * float64(rec.Quantity)
		text := fmt.Sprintf("%s %.0f %s", action, cost, rec.Symbol)

		fig := figure.NewFigure(text, "straight", false)
		figLines := fig.Slicify()

		color := t.Success
		if action == "SELL" {
			color = t.Secondary
		}
		actionStyle := lipgloss.NewStyle().Foreground(color)

		if action == "BUY" {
			buyTotal += cost
		} else {
			sellTotal += cost
		}

		for _, fl := range figLines {
			if strings.TrimSpace(fl) != "" {
				lines = append(lines, actionStyle.Render(fl))
			}
		}
		lines = append(lines, "")
	}

	// Summary
	var parts []string
	if buyTotal > 0 {
		s := lipgloss.NewStyle().Foreground(t.Success)
		parts = append(parts, s.Render(fmt.Sprintf("BUY %.0f", buyTotal)))
	}
	if sellTotal > 0 {
		s := lipgloss.NewStyle().Foreground(t.Secondary)
		parts = append(parts, s.Render(fmt.Sprintf("SELL %.0f", sellTotal)))
	}
	if len(parts) > 0 {
		lines = append(lines, strings.Join(parts, "  "))
	}

	return border.Render(strings.Join(lines, "\n"))
}

// Footer

func (m Model) viewFooter(t theme.Theme) string {
	return lipgloss.NewStyle().
		Width(m.width).
		Background(t.Surface).
		Foreground(t.Text).
		Padding(0, 1).
		Render("q: Quit  r: Refresh  c: Theme  t: Table  \u2191\u2193: Scroll")
}

// Table screen

func (m Model) viewTable(t theme.Theme) string {
	header := lipgloss.NewStyle().
		Width(m.width).
		Background(t.Surface).
		Foreground(t.Text).
		Padding(0, 1).
		Render(fmt.Sprintf(" %s SECURITIES  [dim]ESC to close[/]",
			lipgloss.NewStyle().Foreground(t.Primary).Render("\u25c0")))

	footer := lipgloss.NewStyle().
		Width(m.width).
		Background(t.Surface).
		Foreground(t.Text).
		Padding(0, 1).
		Render("esc/t: Back  \u2191\u2193: Navigate  q: Quit")

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		m.table.View(),
		footer,
	)
}
