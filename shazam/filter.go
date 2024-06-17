package shazam

import (
	"math"
)

// LowPassFilter is a first-order low-pass filter using H(p) = 1 / (1 + pRC)
type LowPassFilter struct {
	alpha float64 // Filter coefficient
	yPrev float64 // Previous output value
}

// NewLowPassFilter creates a new low-pass filter
func NewLowPassFilter(cutoffFrequency, sampleRate float64) *LowPassFilter {
	rc := 1.0 / (2 * math.Pi * cutoffFrequency)
	dt := 1.0 / sampleRate
	alpha := dt / (rc + dt)
	return &LowPassFilter{
		alpha: alpha,
		yPrev: 0,
	}
}

// Filter processes the input signal through the low-pass filter
func (lpf *LowPassFilter) Filter(input []float64) []float64 {
	filtered := make([]float64, len(input))
	for i, x := range input {
		if i == 0 {
			filtered[i] = x * lpf.alpha
		} else {
			filtered[i] = lpf.alpha*x + (1-lpf.alpha)*lpf.yPrev
		}
		lpf.yPrev = filtered[i]
	}
	return filtered
}
