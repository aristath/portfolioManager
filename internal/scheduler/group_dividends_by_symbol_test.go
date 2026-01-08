package scheduler

import (
	"testing"

	"github.com/aristath/sentinel/internal/modules/dividends"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestGroupDividendsBySymbolJob_Name(t *testing.T) {
	job := NewGroupDividendsBySymbolJob()
	assert.Equal(t, "group_dividends_by_symbol", job.Name())
}

func TestGroupDividendsBySymbolJob_Run_Success(t *testing.T) {
	job := NewGroupDividendsBySymbolJob()
	job.SetLogger(zerolog.New(nil).Level(zerolog.Disabled))

	dividends := []dividends.DividendRecord{
		{ID: 1, Symbol: "AAPL", AmountEUR: 10.0},
		{ID: 2, Symbol: "AAPL", AmountEUR: 20.0},
		{ID: 3, Symbol: "MSFT", AmountEUR: 15.0},
	}
	job.SetDividends(dividends)

	err := job.Run()
	assert.NoError(t, err)

	grouped := job.GetGroupedDividends()
	assert.Equal(t, 2, len(grouped))
	assert.Equal(t, 30.0, grouped["AAPL"].TotalAmount)
	assert.Equal(t, 15.0, grouped["MSFT"].TotalAmount)
	assert.Equal(t, 2, grouped["AAPL"].DividendCount)
	assert.Equal(t, 1, grouped["MSFT"].DividendCount)
}

func TestGroupDividendsBySymbolJob_Run_Empty(t *testing.T) {
	job := NewGroupDividendsBySymbolJob()
	job.SetLogger(zerolog.New(nil).Level(zerolog.Disabled))

	err := job.Run()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(job.GetGroupedDividends()))
}
