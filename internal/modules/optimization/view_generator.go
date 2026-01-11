package optimization

import (
	"fmt"

	"github.com/rs/zerolog"
)

// ViewGenerator converts security scores into Black-Litterman views.
type ViewGenerator struct {
	log zerolog.Logger
}

// NewViewGenerator creates a new view generator.
func NewViewGenerator(log zerolog.Logger) *ViewGenerator {
	if log.GetLevel() == zerolog.Disabled {
		log = zerolog.Nop()
	}
	return &ViewGenerator{
		log: log.With().Str("component", "view_generator").Logger(),
	}
}

// GenerateViewsFromScores converts security scores to BL views.
// High scores (>0.8) = positive view, Low scores (<0.5) = negative view.
// Maps are keyed by ISIN.
func (vg *ViewGenerator) GenerateViewsFromScores(
	scores map[string]float64,
	expectedReturns map[string]float64,
) ([]View, error) {
	views := make([]View, 0)

	for isin, score := range scores {
		expectedReturn, hasReturn := expectedReturns[isin]
		if !hasReturn {
			continue
		}

		// Generate view based on score
		if score > 0.8 {
			// High score: positive view (outperform by score-based amount)
			outperformance := (score - 0.5) * 0.10 // Scale to 0-3% outperformance
			viewReturn := expectedReturn + outperformance
			confidence := score // Use score as confidence

			views = append(views, View{
				Type:       "absolute",
				ISIN:       isin,
				Return:     viewReturn,
				Confidence: confidence,
			})
		} else if score < 0.5 {
			// Low score: negative view (underperform)
			underperformance := (0.5 - score) * 0.10 // Scale to 0-5% underperformance
			viewReturn := expectedReturn - underperformance
			confidence := 1.0 - score // Higher confidence for very low scores

			views = append(views, View{
				Type:       "absolute",
				ISIN:       isin,
				Return:     viewReturn,
				Confidence: confidence,
			})
		}
		// Scores between 0.5 and 0.8: no view (use equilibrium)
	}

	return views, nil
}

// CalculateViewUncertainty calculates uncertainty for a view based on score confidence.
func (vg *ViewGenerator) CalculateViewUncertainty(score float64, baseUncertainty float64) float64 {
	// Uncertainty decreases with score confidence
	// High scores (high confidence) = low uncertainty
	uncertainty := baseUncertainty * (1.0 - score*0.5) // Scale uncertainty by confidence

	// Clamp to reasonable range
	if uncertainty < 0.01 {
		uncertainty = 0.01
	}
	if uncertainty > 0.5 {
		uncertainty = 0.5
	}

	return uncertainty
}

// CreateViewMatrix creates the P matrix for BL formula from views.
func (vg *ViewGenerator) CreateViewMatrix(views []View, symbols []string) ([][]float64, error) {
	if len(views) == 0 {
		return nil, fmt.Errorf("views cannot be empty")
	}

	m := len(views)
	n := len(symbols)
	P := make([][]float64, m)

	for i, view := range views {
		P[i] = make([]float64, n)

		if view.Type == "absolute" {
			// Absolute view: single security
			for j, isin := range symbols {
				if isin == view.ISIN {
					P[i][j] = 1.0
					break
				}
			}
		} else if view.Type == "relative" {
			// Relative view: ISIN1 outperforms ISIN2
			for j, isin := range symbols {
				if isin == view.ISIN1 {
					P[i][j] = 1.0
				} else if isin == view.ISIN2 {
					P[i][j] = -1.0
				}
			}
		}
	}

	return P, nil
}
