package allocation

import "time"

// AllocationTarget represents target allocation for geography or industry
// Faithful translation from Python: app/modules/allocation/domain/models.py
type AllocationTarget struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Type      string    `json:"type"`
	Name      string    `json:"name"`
	ID        int64     `json:"id"`
	TargetPct float64   `json:"target_pct"`
}

// Note: ConcentrationAlert has been moved to domain/interfaces.go
// It is now available as domain.ConcentrationAlert
// Type alias is in interfaces.go for backward compatibility

// AllocationInfo represents allocation status for display
type AllocationInfo struct {
	Name         string  `json:"name"`
	TargetPct    float64 `json:"target_pct"`
	CurrentPct   float64 `json:"current_pct"`
	CurrentValue float64 `json:"current_value"`
	Deviation    float64 `json:"deviation"`
}

// DeviationInfo represents allocation deviation for a group
type DeviationInfo struct {
	Status    string  `json:"status"`
	Deviation float64 `json:"deviation"`
	Need      float64 `json:"need"`
}

// Concentration limit constants
// Faithful translation from Python: app/modules/scoring/domain/constants.py
const (
	MaxGeographyConcentration = 0.35 // 35% max per geography
	MaxSectorConcentration    = 0.30 // 30% max per sector
	MaxPositionConcentration  = 0.15 // 15% max per position

	GeographyAlertThreshold = 0.28 // Alert at 28% (80% of 35%)
	SectorAlertThreshold    = 0.24 // Alert at 24% (80% of 30%)
	PositionAlertThreshold  = 0.12 // Alert at 12% (80% of 15%)
)
