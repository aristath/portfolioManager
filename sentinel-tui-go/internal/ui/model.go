package ui

import (
	"sort"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/harmonica"

	"sentinel-tui-go/internal/api"
	"sentinel-tui-go/internal/theme"
)

type Model struct {
	client *api.Client
	apiURL string

	// Data
	connected       bool
	tradingMode     string
	portfolio       *api.Portfolio
	pnlHistory      *api.PnLHistory
	recommendations []api.Recommendation
	securities      []api.Security

	// UI state
	themeIndex int
	showTable  bool
	width      int
	height     int
	ready      bool

	// Animation
	heroTarget   float64
	heroValue    float64
	heroVelocity float64
	spring       harmonica.Spring
	animating    bool

	// Components
	viewport      viewport.Model
	table         table.Model
	sparklineView string
}

// Messages

type healthMsg struct {
	health api.Health
	err    error
}

type portfolioMsg struct {
	portfolio api.Portfolio
	err       error
}

type pnlMsg struct {
	history api.PnLHistory
	err     error
}

type recsMsg struct {
	recs []api.Recommendation
	err  error
}

type securitiesMsg struct {
	securities []api.Security
	err        error
}

type tickMsg time.Time

func NewModel(client *api.Client, apiURL string) Model {
	return Model{
		client: client,
		apiURL: apiURL,
		spring: harmonica.NewSpring(harmonica.FPS(60), 6.0, 1.0),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(fetchAll(m.client)...)
}

// Commands

func fetchAll(c *api.Client) []tea.Cmd {
	return []tea.Cmd{
		fetchHealth(c),
		fetchPortfolio(c),
		fetchPnL(c),
		fetchRecs(c),
		fetchSecurities(c),
	}
}

func fetchHealth(c *api.Client) tea.Cmd {
	return func() tea.Msg {
		h, err := c.Health()
		return healthMsg{h, err}
	}
}

func fetchPortfolio(c *api.Client) tea.Cmd {
	return func() tea.Msg {
		p, err := c.Portfolio()
		return portfolioMsg{p, err}
	}
}

func fetchPnL(c *api.Client) tea.Cmd {
	return func() tea.Msg {
		h, err := c.PnLHistory("1M")
		return pnlMsg{h, err}
	}
}

func fetchRecs(c *api.Client) tea.Cmd {
	return func() tea.Msg {
		r, err := c.Recommendations()
		return recsMsg{r, err}
	}
}

func fetchSecurities(c *api.Client) tea.Cmd {
	return func() tea.Msg {
		s, err := c.Unified()
		if err == nil {
			sort.Slice(s, func(i, j int) bool {
				return s[i].ValueEUR > s[j].ValueEUR
			})
		}
		return securitiesMsg{s, err}
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second/60, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// keep compiler happy — theme is used in view.go/update.go
var _ = theme.Themes
