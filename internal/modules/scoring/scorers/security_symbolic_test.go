package scorers

import (
	"database/sql"
	"testing"

	"github.com/aristath/sentinel/internal/modules/symbolic_regression"
	"github.com/aristath/sentinel/pkg/formulas"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestSecurityScorer_WithDiscoveredFormula(t *testing.T) {
	// Setup test database
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	// Create schema
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS discovered_formulas (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			formula_type TEXT NOT NULL,
			security_type TEXT NOT NULL,
			regime_range_min REAL,
			regime_range_max REAL,
			formula_expression TEXT NOT NULL,
			validation_metrics TEXT NOT NULL,
			fitness_score REAL NOT NULL,
			complexity INTEGER NOT NULL,
			training_examples_count INTEGER,
			discovered_at TEXT NOT NULL,
			is_active INTEGER DEFAULT 1,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP
		);
	`)
	require.NoError(t, err)

	// Insert discovered scoring formula
	// Formula: (long_term * 0.3) + (stability * 0.25) + (dividends * 0.2) + ...
	_, err = db.Exec(`
		INSERT INTO discovered_formulas (
			formula_type, security_type, formula_expression, validation_metrics,
			fitness_score, complexity, discovered_at, is_active
		) VALUES (
			'scoring', 'stock',
			'long_term * 0.3 + stability * 0.25 + dividends * 0.2 + opportunity * 0.12 + short_term * 0.08 + technicals * 0.05',
			'{"spearman": 0.75, "mae": 0.05}',
			0.25, 7, '2024-01-01', 1
		);
	`)
	require.NoError(t, err)

	// Create scorer with formula storage
	log := zerolog.Nop()
	formulaStorage := symbolic_regression.NewFormulaStorage(db, log)

	scorer := NewSecurityScorer()
	scorer.SetFormulaStorage(formulaStorage)

	// Create test input with group scores
	input := ScoreSecurityInput{
		ProductType: "EQUITY",
		DailyPrices: []float64{100, 101, 102, 103, 104},
		MonthlyPrices: []formulas.MonthlyPrice{
			{YearMonth: "2024-01", AvgAdjClose: 100},
			{YearMonth: "2024-02", AvgAdjClose: 101},
		},
		TargetAnnualReturn: 0.11,
			}

	// Score security
	result := scorer.ScoreSecurity(input)

	require.NotNil(t, result)
	assert.Greater(t, result.TotalScore, 0.0)
	assert.LessOrEqual(t, result.TotalScore, 1.0)

	// Verify group scores exist
	assert.NotNil(t, result.GroupScores)
	assert.Contains(t, result.GroupScores, "long_term")
	assert.Contains(t, result.GroupScores, "stability")
}

func TestSecurityScorer_FallbackWhenNoFormula(t *testing.T) {
	// Setup test database without discovered formula
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	log := zerolog.Nop()
	formulaStorage := symbolic_regression.NewFormulaStorage(db, log)

	scorer := NewSecurityScorer()
	scorer.SetFormulaStorage(formulaStorage)

	input := ScoreSecurityInput{
		ProductType: "EQUITY",
		DailyPrices: []float64{100, 101, 102},
		MonthlyPrices: []formulas.MonthlyPrice{
			{YearMonth: "2024-01", AvgAdjClose: 100},
		},
		TargetAnnualReturn: 0.11,
			}

	// Should fall back to static weights when no discovered formula exists
	result := scorer.ScoreSecurity(input)

	require.NotNil(t, result)
	assert.Greater(t, result.TotalScore, 0.0)
}
