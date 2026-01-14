package market_regime

import (
	"database/sql"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
)

// setupRegimeTestDB creates a test database with per-region schema
func setupRegimeTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	_, err = db.Exec(`
		CREATE TABLE market_regime_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			recorded_at INTEGER NOT NULL,
			region TEXT NOT NULL DEFAULT 'GLOBAL',
			raw_score REAL NOT NULL,
			smoothed_score REAL NOT NULL,
			discrete_regime TEXT NOT NULL,
			created_at INTEGER DEFAULT (strftime('%s', 'now'))
		)
	`)
	require.NoError(t, err)

	_, err = db.Exec(`CREATE INDEX idx_regime_history_region ON market_regime_history(region)`)
	require.NoError(t, err)

	return db
}

func TestRegimePersistence_RecordRegimeScoreForRegion(t *testing.T) {
	db := setupRegimeTestDB(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	rp := NewRegimePersistence(db, log)

	// Record US score
	err := rp.RecordRegimeScoreForRegion(RegionUS, MarketRegimeScore(0.3))
	require.NoError(t, err)

	// Record EU score
	err = rp.RecordRegimeScoreForRegion(RegionEU, MarketRegimeScore(-0.2))
	require.NoError(t, err)

	// Verify they're stored separately
	usScore, err := rp.GetCurrentRegimeScoreForRegion(RegionUS)
	require.NoError(t, err)
	assert.InDelta(t, 0.3, float64(usScore), 0.01)

	euScore, err := rp.GetCurrentRegimeScoreForRegion(RegionEU)
	require.NoError(t, err)
	assert.InDelta(t, -0.2, float64(euScore), 0.01)
}

func TestRegimePersistence_SmoothingPerRegion(t *testing.T) {
	db := setupRegimeTestDB(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	rp := NewRegimePersistence(db, log)

	// Record multiple US scores - smoothing should apply
	err := rp.RecordRegimeScoreForRegion(RegionUS, MarketRegimeScore(0.5))
	require.NoError(t, err)

	err = rp.RecordRegimeScoreForRegion(RegionUS, MarketRegimeScore(0.3))
	require.NoError(t, err)

	// Second score should be smoothed (not raw 0.3)
	usScore, err := rp.GetCurrentRegimeScoreForRegion(RegionUS)
	require.NoError(t, err)

	// With alpha=0.1: smoothed = 0.1 * 0.3 + 0.9 * 0.5 = 0.03 + 0.45 = 0.48
	assert.InDelta(t, 0.48, float64(usScore), 0.01)
}

func TestRegimePersistence_GetAllCurrentScores(t *testing.T) {
	db := setupRegimeTestDB(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	rp := NewRegimePersistence(db, log)

	// Record scores for all regions
	_ = rp.RecordRegimeScoreForRegion(RegionUS, MarketRegimeScore(0.4))
	_ = rp.RecordRegimeScoreForRegion(RegionEU, MarketRegimeScore(0.2))
	_ = rp.RecordRegimeScoreForRegion(RegionAsia, MarketRegimeScore(-0.1))

	// Get all current scores
	scores, err := rp.GetAllCurrentScores()
	require.NoError(t, err)

	assert.Len(t, scores, 3)
	assert.InDelta(t, 0.4, scores[RegionUS], 0.01)
	assert.InDelta(t, 0.2, scores[RegionEU], 0.01)
	assert.InDelta(t, -0.1, scores[RegionAsia], 0.01)
}

func TestRegimePersistence_RegionIsolation(t *testing.T) {
	db := setupRegimeTestDB(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	rp := NewRegimePersistence(db, log)

	// Record US score
	_ = rp.RecordRegimeScoreForRegion(RegionUS, MarketRegimeScore(0.5))
	_ = rp.RecordRegimeScoreForRegion(RegionUS, MarketRegimeScore(0.6))

	// Record EU score - should NOT affect US smoothing
	_ = rp.RecordRegimeScoreForRegion(RegionEU, MarketRegimeScore(-0.5))

	// US smoothing should only consider US history
	// First US: 0.5 (first entry, no smoothing)
	// Second US: 0.1 * 0.6 + 0.9 * 0.5 = 0.06 + 0.45 = 0.51
	usScore, err := rp.GetCurrentRegimeScoreForRegion(RegionUS)
	require.NoError(t, err)
	assert.InDelta(t, 0.51, float64(usScore), 0.01)

	// EU should be raw -0.5 (first entry)
	euScore, err := rp.GetCurrentRegimeScoreForRegion(RegionEU)
	require.NoError(t, err)
	assert.InDelta(t, -0.5, float64(euScore), 0.01)
}

func TestRegimePersistence_NoDataForRegion(t *testing.T) {
	db := setupRegimeTestDB(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	rp := NewRegimePersistence(db, log)

	// Query region with no data
	score, err := rp.GetCurrentRegimeScoreForRegion(RegionAsia)
	require.NoError(t, err)
	assert.Equal(t, NeutralScore, score)
}

func TestRegimePersistence_GetRegimeHistoryForRegion(t *testing.T) {
	db := setupRegimeTestDB(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	rp := NewRegimePersistence(db, log)

	// Record multiple scores for US
	_ = rp.RecordRegimeScoreForRegion(RegionUS, MarketRegimeScore(0.1))
	time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	_ = rp.RecordRegimeScoreForRegion(RegionUS, MarketRegimeScore(0.2))
	time.Sleep(10 * time.Millisecond)
	_ = rp.RecordRegimeScoreForRegion(RegionUS, MarketRegimeScore(0.3))

	// Record EU score (should not appear in US history)
	_ = rp.RecordRegimeScoreForRegion(RegionEU, MarketRegimeScore(-0.5))

	// Get US history
	history, err := rp.GetRegimeHistoryForRegion(RegionUS, 10)
	require.NoError(t, err)
	assert.Len(t, history, 3)

	// Should be ordered DESC (most recent first) - check raw score (not smoothed)
	// The raw score stored should be the actual input value
	assert.InDelta(t, 0.3, history[0].RawScore, 0.01) // Most recent raw score

	// Verify all entries are for US region
	for _, entry := range history {
		assert.Equal(t, RegionUS, entry.Region)
	}
}

func TestRegimePersistence_BackwardsCompatibility(t *testing.T) {
	db := setupRegimeTestDB(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	rp := NewRegimePersistence(db, log)

	// Record using old global method (should use GLOBAL region)
	err := rp.RecordRegimeScore(MarketRegimeScore(0.5))
	require.NoError(t, err)

	// Get using old method (should return GLOBAL score)
	score, err := rp.GetCurrentRegimeScore()
	require.NoError(t, err)
	assert.InDelta(t, 0.5, float64(score), 0.01)

	// Verify it was stored with GLOBAL region
	globalScore, err := rp.GetCurrentRegimeScoreForRegion("GLOBAL")
	require.NoError(t, err)
	assert.InDelta(t, 0.5, float64(globalScore), 0.01)
}
