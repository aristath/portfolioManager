package utils

import (
	"math"
)

// TemperamentParams defines parameters for temperament value transformation
type TemperamentParams struct {
	// Type is the temperament type identifier (e.g., "risk_tolerance")
	// Currently unused but reserved for future use
	Type string
	// Value is the raw input value (0.0 to 1.0)
	Value float64
	// Min is the minimum output value
	Min float64
	// Max is the maximum output value
	Max float64
	// Progression defines the transformation curve type
	// Valid values: linear, linear-reverse, exponential, exponential-reverse,
	// logarithmic, logarithmic-reverse, sigmoid, sigmoid-reverse
	Progression string
	// Base is the current neutral/default value (used when value = 0.5)
	Base float64
}

// GetTemperamentValue transforms a raw value (0.0-1.0) based on progression type
// and returns an adjusted value between min and max.
//
// The function ensures that when value = 0.5 (neutral), the result equals Base,
// allowing for smooth transitions from existing defaults.
//
// Examples:
//   - GetTemperamentValue({Value: 0.5, Min: 0.3, Max: 0.7, Progression: "linear", Base: 0.5})
//     returns 0.5 (neutral = base)
//   - GetTemperamentValue({Value: 0.0, Min: 0.3, Max: 0.7, Progression: "linear", Base: 0.5})
//     returns 0.3 (min, conservative)
//   - GetTemperamentValue({Value: 1.0, Min: 0.3, Max: 0.7, Progression: "linear", Base: 0.5})
//     returns 0.7 (max, aggressive)
func GetTemperamentValue(params TemperamentParams) float64 {
	// Clamp value to [0, 1]
	value := math.Max(0.0, math.Min(1.0, params.Value))

	// Calculate the progression value based on type (normalized to [0, 1])
	var progressionValue float64

	switch params.Progression {
	case "linear":
		progressionValue = value
	case "linear-reverse":
		progressionValue = 1.0 - value
	case "exponential":
		progressionValue = value * value
	case "exponential-reverse":
		progressionValue = (1.0 - value) * (1.0 - value)
	case "logarithmic":
		if value == 0.0 {
			progressionValue = 0.0
		} else {
			logMax := math.Log10(10.0)
			progressionValue = math.Log10(1.0 + 9.0*value) / logMax
		}
	case "logarithmic-reverse":
		if value == 1.0 {
			progressionValue = 0.0
		} else {
			logMax := math.Log10(10.0)
			progressionValue = math.Log10(1.0 + 9.0*(1.0-value)) / logMax
		}
	case "sigmoid":
		exponent := -12.0 * (value - 0.5)
		progressionValue = 1.0 / (1.0 + math.Exp(exponent))
	case "sigmoid-reverse":
		exponent := -12.0 * (value - 0.5)
		progressionValue = 1.0 - (1.0 / (1.0 + math.Exp(exponent)))
	default:
		progressionValue = value
	}

	// We need to ensure three anchor points:
	// - value = 0.0 → result = Min (most conservative)
	// - value = 0.5 → result = Base (neutral)
	// - value = 1.0 → result = Max (most aggressive)
	//
	// The progressionValue gives us a normalized curve [0,1].
	// We'll use piecewise linear interpolation between the three anchor points,
	// but use the progression curve to determine the interpolation weight.

	var result float64

	// Calculate progression values at key anchor points
	var progAtZero, progAtHalf, progAtOne float64
	switch params.Progression {
	case "linear":
		progAtZero, progAtHalf, progAtOne = 0.0, 0.5, 1.0
	case "linear-reverse":
		progAtZero, progAtHalf, progAtOne = 1.0, 0.5, 0.0
	case "exponential":
		progAtZero, progAtHalf, progAtOne = 0.0, 0.25, 1.0
	case "exponential-reverse":
		progAtZero, progAtHalf, progAtOne = 1.0, 0.25, 0.0
	case "logarithmic":
		logMax := math.Log10(10.0)
		progAtZero = 0.0
		progAtHalf = math.Log10(1.0 + 9.0*0.5) / logMax
		progAtOne = 1.0
	case "logarithmic-reverse":
		logMax := math.Log10(10.0)
		progAtZero = 1.0
		progAtHalf = math.Log10(1.0 + 9.0*0.5) / logMax
		progAtOne = 0.0
	case "sigmoid":
		progAtZero = 1.0 / (1.0 + math.Exp(6.0))
		progAtHalf = 0.5
		progAtOne = 1.0 / (1.0 + math.Exp(-6.0))
	case "sigmoid-reverse":
		progAtZero = 1.0 - (1.0 / (1.0 + math.Exp(6.0)))
		progAtHalf = 0.5
		progAtOne = 1.0 - (1.0 / (1.0 + math.Exp(-6.0)))
	default:
		progAtZero, progAtHalf, progAtOne = 0.0, 0.5, 1.0
	}

	// Handle reverse progressions where progression values are decreasing
	isReverse := progAtZero > progAtHalf

	if value <= 0.5 {
		// For [0, 0.5]: map to output range
		// Normal: [Min, Base], Reverse: [Max, Base] (swapped)
		var outputStart, outputEnd float64
		if isReverse {
			outputStart, outputEnd = params.Max, params.Base // Reverse: value=0.0 → Max
		} else {
			outputStart, outputEnd = params.Min, params.Base // Normal: value=0.0 → Min
		}

		var t float64
		if isReverse {
			// For reverse: progression goes from high to low
			// progAtZero (high) should map to outputStart (Max), progAtHalf should map to outputEnd (Base)
			if math.Abs(progAtZero-progAtHalf) < 0.0001 {
				t = value / 0.5
			} else {
				t = (progAtZero - progressionValue) / (progAtZero - progAtHalf)
				t = math.Max(0.0, math.Min(1.0, t))
			}
		} else {
			// For normal: progression goes from low to high
			if math.Abs(progAtHalf-progAtZero) < 0.0001 {
				t = value / 0.5
			} else {
				t = (progressionValue - progAtZero) / (progAtHalf - progAtZero)
				t = math.Max(0.0, math.Min(1.0, t))
			}
		}
		result = outputStart + t*(outputEnd-outputStart)
	} else {
		// For [0.5, 1.0]: map to output range
		// Normal: [Base, Max], Reverse: [Base, Min] (swapped)
		var outputStart, outputEnd float64
		if isReverse {
			outputStart, outputEnd = params.Base, params.Min // Reverse: value=1.0 → Min
		} else {
			outputStart, outputEnd = params.Base, params.Max // Normal: value=1.0 → Max
		}

		var t float64
		if isReverse {
			// For reverse: progAtHalf should map to Base, progAtOne (low) should map to outputEnd (Min)
			if math.Abs(progAtHalf-progAtOne) < 0.0001 {
				t = (value - 0.5) / 0.5
			} else {
				t = (progAtHalf - progressionValue) / (progAtHalf - progAtOne)
				t = math.Max(0.0, math.Min(1.0, t))
			}
		} else {
			// For normal: progAtHalf should map to Base, progAtOne (high) should map to Max
			if math.Abs(progAtOne-progAtHalf) < 0.0001 {
				t = (value - 0.5) / 0.5
			} else {
				t = (progressionValue - progAtHalf) / (progAtOne - progAtHalf)
				t = math.Max(0.0, math.Min(1.0, t))
			}
		}
		result = outputStart + t*(outputEnd-outputStart)
	}

	// Clamp to bounds for safety
	result = math.Max(params.Min, math.Min(params.Max, result))

	return result
}
