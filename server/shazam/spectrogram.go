package shazam

import (
	"errors"
	"fmt"
	"math"
	"math/cmplx"
)

const (
	dspRatio    = 4
	freqBinSize = 1024
	maxFreq     = 5000.0 // 5kHz
	hopSize     = freqBinSize / 32
)

func Spectrogram(sample []float64, sampleRate int) ([][]complex128, error) {
	filteredSample := LowPassFilter(maxFreq, float64(sampleRate), sample)

	downsampledSample, err := Downsample(filteredSample, sampleRate, sampleRate/dspRatio)
	if err != nil {
		return nil, fmt.Errorf("couldn't downsample audio sample: %v", err)
	}

	numOfWindows := len(downsampledSample) / (freqBinSize - hopSize)
	spectrogram := make([][]complex128, numOfWindows)

	window := make([]float64, freqBinSize)
	for i := range window {
		window[i] = 0.54 - 0.46*math.Cos(2*math.Pi*float64(i)/(float64(freqBinSize)-1))
	}

	// Perform STFT
	for i := 0; i < numOfWindows; i++ {
		start := i * hopSize
		end := start + freqBinSize
		if end > len(downsampledSample) {
			end = len(downsampledSample)
		}

		bin := make([]float64, freqBinSize)
		copy(bin, downsampledSample[start:end])

		// Apply Hamming window
		for j := range window {
			bin[j] *= window[j]
		}

		spectrogram[i] = FFT(bin)
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

type Peak struct {
	Time float64
	Freq complex128
}

// ExtractPeaks analyzes a spectrogram and extracts significant peaks in the frequency domain over time.
func ExtractPeaks(spectrogram [][]complex128, audioDuration float64) []Peak {
	if len(spectrogram) < 1 {
		return []Peak{}
	}

	type maxies struct {
		maxMag  float64
		maxFreq complex128
		freqIdx int
	}

	bands := []struct{ min, max int }{{0, 10}, {10, 20}, {20, 40}, {40, 80}, {80, 160}, {160, 512}}

	var peaks []Peak
	binDuration := audioDuration / float64(len(spectrogram))

	for binIdx, bin := range spectrogram {
		var maxMags []float64
		var maxFreqs []complex128
		var freqIndices []float64

		binBandMaxies := []maxies{}
		for _, band := range bands {
			var maxx maxies
			var maxMag float64
			for idx, freq := range bin[band.min:band.max] {
				magnitude := cmplx.Abs(freq)
				if magnitude > maxMag {
					maxMag = magnitude
					freqIdx := band.min + idx
					maxx = maxies{magnitude, freq, freqIdx}
				}
			}
			binBandMaxies = append(binBandMaxies, maxx)
		}

		for _, value := range binBandMaxies {
			maxMags = append(maxMags, value.maxMag)
			maxFreqs = append(maxFreqs, value.maxFreq)
			freqIndices = append(freqIndices, float64(value.freqIdx))
		}

		// Calculate the average magnitude
		var maxMagsSum float64
		for _, max := range maxMags {
			maxMagsSum += max
		}
		avg := maxMagsSum / float64(len(maxFreqs)) // * coefficient

		// Add peaks that exceed the average magnitude
		for i, value := range maxMags {
			if value > avg {
				peakTimeInBin := freqIndices[i] * binDuration / float64(len(bin))

				// Calculate the absolute time of the peak
				peakTime := float64(binIdx)*binDuration + peakTimeInBin

				peaks = append(peaks, Peak{Time: peakTime, Freq: maxFreqs[i]})
			}
		}
	}

	return peaks
}
