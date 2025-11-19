package shazam

import (
	"errors"
	"fmt"
	"math"
	"math/cmplx"
)

const (
	dspRatio   = 4
	windowSize = 1024
	maxFreq    = 5000.0         // 5kHz
	hopSize    = windowSize / 2 // 50% overlap for better time-frequency resolution
	windowType = "hanning"      // choices: "hanning" or "hamming"
)

func Spectrogram(sample []float64, sampleRate int) ([][]float64, error) {
	filteredSample := LowPassFilter(maxFreq, float64(sampleRate), sample)

	downsampledSample, err := Downsample(filteredSample, sampleRate, sampleRate/dspRatio)
	if err != nil {
		return nil, fmt.Errorf("couldn't downsample audio sample: %v", err)
	}

	window := make([]float64, windowSize)
	for i := range window {
		theta := 2 * math.Pi * float64(i) / float64(windowSize-1)
		switch windowType {
		case "hamming":
			window[i] = 0.54 - 0.46*math.Cos(theta)
		default: // Hanning window
			window[i] = 0.5 - 0.5*math.Cos(theta)
		}
	}

	// Initialize spectrogram slice
	spectrogram := make([][]float64, 0)

	// Perform STFT
	for start := 0; start+windowSize <= len(downsampledSample); start += hopSize {
		end := start + windowSize

		frame := make([]float64, windowSize)
		copy(frame, downsampledSample[start:end])

		// Apply window
		for j := range window {
			frame[j] *= window[j]
		}

		// Perform FFT
		fftResult := FFT(frame)

		// Convert complex spectrum to magnitude spectrum
		magnitude := make([]float64, len(fftResult)/2)
		for j := range magnitude {
			magnitude[j] = cmplx.Abs(fftResult[j])
		}

		spectrogram = append(spectrogram, magnitude)
	}

	return spectrogram, nil
}

// LowPassFilter is a first-order low-pass filter that attenuates high
// frequencies above the cutoffFrequency.
// It uses the transfer function H(s) = 1 / (1 + sRC), where RC is the time constant.
func LowPassFilter(cutoffFrequency, sampleRate float64, input []float64) []float64 {
	rc := 1.0 / (2 * math.Pi * cutoffFrequency)
	dt := 1.0 / sampleRate
	alpha := dt / (rc + dt)

	filteredSignal := make([]float64, len(input))
	var prevOutput float64 = 0

	for i, x := range input {
		if i == 0 {
			filteredSignal[i] = x * alpha
		} else {

			filteredSignal[i] = alpha*x + (1-alpha)*prevOutput
		}
		prevOutput = filteredSignal[i]
	}
	return filteredSignal
}

// Downsample downsamples the input audio from originalSampleRate to targetSampleRate
func Downsample(input []float64, originalSampleRate, targetSampleRate int) ([]float64, error) {
	if targetSampleRate <= 0 || originalSampleRate <= 0 {
		return nil, errors.New("sample rates must be positive")
	}
	if targetSampleRate > originalSampleRate {
		return nil, errors.New("target sample rate must be less than or equal to original sample rate")
	}

	ratio := originalSampleRate / targetSampleRate
	if ratio <= 0 {
		return nil, errors.New("invalid ratio calculated from sample rates")
	}

	var resampled []float64
	for i := 0; i < len(input); i += ratio {
		end := i + ratio
		if end > len(input) {
			end = len(input)
		}

		sum := 0.0
		for j := i; j < end; j++ {
			sum += input[j]
		}
		avg := sum / float64(end-i)
		resampled = append(resampled, avg)
	}

	return resampled, nil
}

// Peak represents a significant point in the spectrogram.
type Peak struct {
	Freq float64 // Frequency in Hz
	Time float64 // Time in seconds
}

// ExtractPeaks analyzes a spectrogram and extracts significant peaks in the frequency domain over time.
func ExtractPeaks(spectrogram [][]float64, audioDuration float64, sampleRate int) []Peak {
	if len(spectrogram) < 1 {
		return []Peak{}
	}

	type maxies struct {
		maxMag  float64
		freqIdx int
	}

	bands := []struct{ min, max int }{
		{0, 10}, {10, 20}, {20, 40}, {40, 80}, {80, 160}, {160, 512},
	}

	var peaks []Peak
	frameDuration := audioDuration / float64(len(spectrogram))

	// Calculate frequency resolution (Hz per bin)
	effectiveSampleRate := float64(sampleRate) / float64(dspRatio)
	freqResolution := effectiveSampleRate / float64(windowSize)

	for frameIdx, frame := range spectrogram {
		var maxMags []float64
		var freqIndices []int

		binBandMaxies := []maxies{}
		for _, band := range bands {
			var maxx maxies
			var maxMag float64
			for idx, mag := range frame[band.min:band.max] {
				if mag > maxMag {
					maxMag = mag
					freqIdx := band.min + idx
					maxx = maxies{mag, freqIdx}
				}
			}
			binBandMaxies = append(binBandMaxies, maxx)
		}

		for _, value := range binBandMaxies {
			maxMags = append(maxMags, value.maxMag)
			freqIndices = append(freqIndices, value.freqIdx)
		}

		// Calculate the average magnitude
		var maxMagsSum float64
		for _, max := range maxMags {
			maxMagsSum += max
		}
		avg := maxMagsSum / float64(len(maxMags))

		// Add peaks that exceed the average magnitude
		for i, value := range maxMags {
			if value > avg {
				peakTime := float64(frameIdx) * frameDuration
				peakFreq := float64(freqIndices[i]) * freqResolution

				peaks = append(peaks, Peak{Time: peakTime, Freq: peakFreq})
			}
		}
	}

	return peaks
}
