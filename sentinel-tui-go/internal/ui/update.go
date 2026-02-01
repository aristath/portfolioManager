package ui

import (
	"fmt"
	"math"

	"github.com/NimbleMarkets/ntcharts/sparkline"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"sentinel-tui-go/internal/theme"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport = viewport.New(m.width, m.height-2) // status + footer
		m.ready = true
		m.rebuildSparkline()

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, keys.Refresh):
			return m, tea.Batch(fetchAll(m.client)...)
		case key.Matches(msg, keys.Theme):
			m.themeIndex = (m.themeIndex + 1) % len(theme.Themes)
			m.rebuildSparkline()
		case key.Matches(msg, keys.Table):
			m.showTable = !m.showTable
		case key.Matches(msg, keys.Back):
			if m.showTable {
				m.showTable = false
			}
		}

	case healthMsg:
		if msg.err != nil {
			m.connected = false
		} else {
			m.connected = true
			m.tradingMode = msg.health.TradingMode
		}

	case portfolioMsg:
		if msg.err == nil {
			m.portfolio = &msg.portfolio
			target := msg.portfolio.TotalValueEUR
			if m.heroTarget != target {
				m.heroTarget = target
				m.animating = true
				cmds = append(cmds, tickCmd())
			}
		}

	case pnlMsg:
		if msg.err == nil {
			m.pnlHistory = &msg.history
			m.rebuildSparkline()
		}

	case recsMsg:
		if msg.err == nil {
			m.recommendations = msg.recs
		}

	case securitiesMsg:
		if msg.err == nil {
			m.securities = msg.securities
			m.rebuildTable()
		}

	case tickMsg:
		if m.animating {
			m.heroValue, m.heroVelocity = m.spring.Update(
				m.heroValue, m.heroVelocity, m.heroTarget,
			)
			if math.Abs(m.heroValue-m.heroTarget) < 1 &&
				math.Abs(m.heroVelocity) < 0.1 {
				m.heroValue = m.heroTarget
				m.animating = false
			} else {
				cmds = append(cmds, tickCmd())
			}
		}
	}

	// Rebuild viewport content
	if m.ready && !m.showTable {
		m.rebuildContent()
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	// Forward to table when visible
	if m.showTable {
		var cmd tea.Cmd
		m.table, cmd = m.table.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) rebuildSparkline() {
	if m.pnlHistory == nil || len(m.pnlHistory.Snapshots) == 0 || m.width < 14 {
		m.sparklineView = ""
		return
	}
	t := theme.Themes[m.themeIndex]
	data := make([]float64, len(m.pnlHistory.Snapshots))
	for i, s := range m.pnlHistory.Snapshots {
		data[i] = s.PnLPct
	}
	w := m.width - 4
	style := lipgloss.NewStyle().Foreground(t.Primary)
	sl := sparkline.New(w, 1, sparkline.WithStyle(style))
	sl.PushAll(data)
	sl.DrawBraille()
	m.sparklineView = sl.View()
}

func (m *Model) rebuildTable() {
	t := theme.Themes[m.themeIndex]
	columns := []table.Column{
		{Title: "Symbol", Width: 14},
		{Title: "Name", Width: 30},
		{Title: "Value (\u20ac)", Width: 12},
		{Title: "P/L (%)", Width: 10},
		{Title: "Score", Width: 8},
	}

	var rows []table.Row
	for _, sec := range m.securities {
		var valStr, plStr, scoreStr string
		if sec.HasPosition {
			valStr = fmt.Sprintf("%.2f", sec.ValueEUR)
			plStr = fmt.Sprintf("%+.2f%%", sec.ProfitPct)
		} else {
			valStr = "-"
			plStr = "-"
		}
		if sec.PlannerScore > 0 {
			scoreStr = fmt.Sprintf("%.1f", sec.PlannerScore)
		} else {
			scoreStr = "-"
		}
		rows = append(rows, table.Row{sec.Symbol, sec.Name, valStr, plStr, scoreStr})
	}

	h := m.height - 3
	if h < 5 {
		h = 5
	}
	m.table = table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(h),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.Foreground(t.Primary).Bold(true)
	s.Selected = s.Selected.Foreground(t.Background).Background(t.Primary)
	m.table.SetStyles(s)
}
