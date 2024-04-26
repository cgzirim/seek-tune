package shazam

import (
	"errors"
	"fmt"
	"math"
)

const (
	dspRatio         = 4
	lowPassFilter    = 5000.0 // 5kHz
	samplesPerWindow = 1024
)

func Spectrogram(samples []float64, channels, sampleRate int) [][]complex128 {
	lpf := NewLowPassFilter(lowPassFilter, float64(sampleRate))
	filteredSamples := lpf.Filter(samples)

	downsampledSamples, err := downsample(filteredSamples, dspRatio)
	if err != nil {
		fmt.Println("Couldn't downsample audio samples: ", err)
	}

	hopSize := samplesPerWindow / 32
	numOfWindows := len(downsampledSamples) / (samplesPerWindow - hopSize)
	spectrogram := make([][]complex128, numOfWindows)

	// Apply Hamming window function
	windowSize := len(samples)
	for i := 0; i < len(downsampledSamples); i++ {
		downsampledSamples[i] = 0.54 - 0.46*math.Cos(2*math.Pi*float64(i)/(float64(windowSize)-1))
	}

	// Perform STFT
	for i := 0; i < numOfWindows; i++ {
		start := i * hopSize
		end := start + samplesPerWindow
		if end > len(downsampledSamples) {
			end = len(downsampledSamples)
		}

		spec := make([]float64, samplesPerWindow)
		for j := start; j < end; j++ {
			spec[j-start] = downsampledSamples[j]
		}

		applyHammingWindow(spec)
		spectrogram[i] = FFT(spec)
	}

	return spectrogram
}

func applyHammingWindow(samples []float64) {
	windowSize := len(samples)

	for i := 0; i < windowSize; i++ {
		samples[i] *= 0.54 - 0.46*math.Cos(2*math.Pi*float64(i)/(float64(windowSize)-1))
	}
}

// Downsample downsamples a list of float64 values from 44100 Hz to a specified ratio by averaging groups of samples
func downsample(input []float64, ratio int) ([]float64, error) {
	// Ensure the ratio is valid and compatible with the input length
	if ratio <= 0 || len(input)%ratio != 0 {
		return nil, errors.New("invalid or incompatible ratio")
	}

	// Calculate the size of the output slice
	outputSize := len(input) / ratio

	// Create the output slice
	output := make([]float64, outputSize)

	// Iterate over the input and calculate averages for each group of samples
	for i := 0; i < outputSize; i++ {
		startIndex := i * ratio
		endIndex := startIndex + ratio
		sum := 0.0

		// Sum up the values in the current group of samples
		for j := startIndex; j < endIndex; j++ {
			sum += input[j]
		}

		// Calculate the average for the current group
		output[i] = sum / float64(ratio)
	}

	return output, nil
}
