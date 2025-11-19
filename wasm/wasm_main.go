//go:build js && wasm
// +build js,wasm

package main

import (
	"song-recognition/models"
	"song-recognition/shazam"
	"song-recognition/utils"
	"syscall/js"
)

// generateFingerprint takes audio data from the frontend and generates fingerprints
// Arguments: [audioArray, sampleRate, channels]
// Returns: { error: number, data: fingerprintArray or error message }
func generateFingerprint(this js.Value, args []js.Value) interface{} {
	if len(args) < 3 {
		return js.ValueOf(map[string]interface{}{
			"error": 1,
			"data":  "Expected audio array, sample rate, and number of channels",
		})
	}

	if args[0].Type() != js.TypeObject || args[1].Type() != js.TypeNumber {
		return js.ValueOf(map[string]interface{}{
			"error": 2,
			"data":  "Invalid argument types; Expected audio array and samplerate (type: int)",
		})
	}

	channels := args[2].Int()
	if args[2].Type() != js.TypeNumber || (channels != 1 && channels != 2) {
		return js.ValueOf(map[string]interface{}{
			"error": 2,
			"data":  "Invalid number of channels; expected 1 or 2",
		})
	}

	inputArray := args[0]
	sampleRate := args[1].Int()

	audioData := make([]float64, inputArray.Length())
	for i := 0; i < inputArray.Length(); i++ {
		audioData[i] = inputArray.Index(i).Float()
	}

	fingerprint := make(map[uint32]models.Couple)
	var leftChannel, rightChannel []float64

	if channels == 1 {
		leftChannel = audioData
		spectrogram, err := shazam.Spectrogram(audioData, sampleRate)
		if err != nil {
			return js.ValueOf(map[string]interface{}{
				"error": 3,
				"data":  "Error generating spectrogram: " + err.Error(),
			})
		}
		peaks := shazam.ExtractPeaks(spectrogram, float64(len(audioData))/float64(sampleRate), sampleRate)
		fingerprint = shazam.Fingerprint(peaks, utils.GenerateUniqueID())
	} else {
		for i := 0; i < len(audioData); i += 2 {
			leftChannel = append(leftChannel, audioData[i])
			rightChannel = append(rightChannel, audioData[i+1])
		}

		// LEFT
		spectrogram, err := shazam.Spectrogram(leftChannel, sampleRate)
		if err != nil {
			return js.ValueOf(map[string]interface{}{
				"error": 3,
				"data":  "Error generating spectrogram: " + err.Error(),
			})
		}
		peaks := shazam.ExtractPeaks(spectrogram, float64(len(leftChannel))/float64(sampleRate), sampleRate)
		utils.ExtendMap(fingerprint, shazam.Fingerprint(peaks, utils.GenerateUniqueID()))

		// RIGHT
		spectrogram, err = shazam.Spectrogram(rightChannel, sampleRate)
		if err != nil {
			return js.ValueOf(map[string]interface{}{
				"error": 3,
				"data":  "Error generating spectrogram: " + err.Error(),
			})
		}
		peaks = shazam.ExtractPeaks(spectrogram, float64(len(rightChannel))/float64(sampleRate), sampleRate)
		utils.ExtendMap(fingerprint, shazam.Fingerprint(peaks, utils.GenerateUniqueID()))
	}

	fingerprintArray := []interface{}{}
	for address, couple := range fingerprint {
		entry := map[string]interface{}{
			"address":    address,
			"anchorTime": couple.AnchorTimeMs,
		}
		fingerprintArray = append(fingerprintArray, entry)
	}

	return js.ValueOf(map[string]interface{}{
		"error": 0,
		"data":  fingerprintArray,
	})
}

func main() {
	js.Global().Set("generateFingerprint", js.FuncOf(generateFingerprint))

	js.Global().Call("dispatchEvent", js.Global().Get("Event").New("wasmReady"))

	select {}
}
