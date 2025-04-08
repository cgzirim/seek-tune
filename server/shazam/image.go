package shazam

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"math/cmplx"
	"os"
)

// ConvertSpectrogramToImage converts a spectrogram to a heat map image
func SpectrogramToImage(spectrogram [][]complex128, outputPath string) error {
	numWindows := len(spectrogram)
	numFreqBins := len(spectrogram[0])

	img := image.NewGray(image.Rect(0, 0, numFreqBins, numWindows))

	// Scale the values in the spectrogram to the range [0, 255]
	maxMagnitude := 0.0
	for i := 0; i < numWindows; i++ {
		for j := 0; j < numFreqBins; j++ {
			magnitude := cmplx.Abs(spectrogram[i][j])
			if magnitude > maxMagnitude {
				maxMagnitude = magnitude
			}
		}
	}

	// Convert spectrogram values to pixel intensities
	for i := 0; i < numWindows; i++ {
		for j := 0; j < numFreqBins; j++ {
			magnitude := cmplx.Abs(spectrogram[i][j])
			intensity := uint8(math.Floor(255 * (magnitude / maxMagnitude)))
			img.SetGray(j, i, color.Gray{Y: intensity})
		}
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	err = png.Encode(file, img)
	if err != nil {
		return err
	}

	return nil
}
