package scheduler

import (
	"github.com/aristath/sentinel/internal/modules/dividends"
	"github.com/rs/zerolog"
)

// SymbolDividendInfoForGroup holds aggregated dividend information for a symbol
type SymbolDividendInfoForGroup struct {
	Dividends     []dividends.DividendRecord
	DividendIDs   []int
	TotalAmount   float64
	DividendCount int
}

// GroupDividendsBySymbolJob groups dividends by symbol and sums amounts
type GroupDividendsBySymbolJob struct {
	log              zerolog.Logger
	dividendRecords  []dividends.DividendRecord
	groupedDividends map[string]SymbolDividendInfoForGroup
}

// NewGroupDividendsBySymbolJob creates a new GroupDividendsBySymbolJob
func NewGroupDividendsBySymbolJob() *GroupDividendsBySymbolJob {
	return &GroupDividendsBySymbolJob{
		log:              zerolog.Nop(),
		groupedDividends: make(map[string]SymbolDividendInfoForGroup),
	}
}

// SetLogger sets the logger for the job
func (j *GroupDividendsBySymbolJob) SetLogger(log zerolog.Logger) {
	j.log = log
}

// SetDividends sets the dividends to group
func (j *GroupDividendsBySymbolJob) SetDividends(dividends []dividends.DividendRecord) {
	j.dividendRecords = dividends
}

// GetGroupedDividends returns the grouped dividends
func (j *GroupDividendsBySymbolJob) GetGroupedDividends() map[string]SymbolDividendInfoForGroup {
	return j.groupedDividends
}

// Name returns the job name
func (j *GroupDividendsBySymbolJob) Name() string {
	return "group_dividends_by_symbol"
}

// Run executes the group dividends by symbol job
func (j *GroupDividendsBySymbolJob) Run() error {
	if len(j.dividendRecords) == 0 {
		j.log.Info().Msg("No dividends to group")
		return nil
	}

	grouped := make(map[string]SymbolDividendInfoForGroup)

	for _, dividend := range j.dividendRecords {
		info, exists := grouped[dividend.Symbol]
		if !exists {
			info = SymbolDividendInfoForGroup{
				Dividends:   []dividends.DividendRecord{},
				DividendIDs: []int{},
			}
		}

		info.Dividends = append(info.Dividends, dividend)
		info.DividendIDs = append(info.DividendIDs, dividend.ID)
		info.TotalAmount += dividend.AmountEUR
		info.DividendCount++

		grouped[dividend.Symbol] = info
	}

	j.groupedDividends = grouped

	j.log.Info().
		Int("symbols", len(grouped)).
		Int("total_dividends", len(j.dividendRecords)).
		Msg("Successfully grouped dividends by symbol")

	return nil
}
