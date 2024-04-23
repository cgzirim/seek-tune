package shazam

import (
	"math"
)

// LowPassFilter implements a simple first-order low-pass filter
type LowPassFilter struct {
	alpha float64 // Filter coefficient
	yPrev float64 // Previous output value
}

// NewLowPassFilter creates a new LowPassFilter with the specified cutoff frequency and sample rate
func NewLowPassFilter(cutoffFrequency, sampleRate float64) *LowPassFilter {
	// Calculate filter coefficient (alpha) based on cutoff frequency and sample rate
	alpha := 1.0 - math.Exp(-2.0*math.Pi*cutoffFrequency/sampleRate)

	return &LowPassFilter{
		alpha: alpha,
		yPrev: 0,
	}
}

// Filter filters the input signal using the low-pass filter and returns the filtered output
func (lpf *LowPassFilter) Filter(input []float64) []float64 {
	filtered := make([]float64, len(input))

	for i, x := range input {
		// Update filter output using the single-pole low-pass filter equation
		output := lpf.alpha*x + (1-lpf.alpha)*lpf.yPrev
		lpf.yPrev = output

		filtered[i] = output
	}

	return filtered
}
