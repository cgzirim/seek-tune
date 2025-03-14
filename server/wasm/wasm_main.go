//go:build js && wasm
// +build js,wasm

package main

import (
	"song-recognition/shazam"
	"song-recognition/utils"
	"syscall/js"
)

func generateFingerprint(this js.Value, args []js.Value) interface{} {
	if len(args) < 2 {
		return js.ValueOf(map[string]interface{}{
			"error": 1,
			"data":  "Expected audio array and sample rate",
		})
	}

	if args[0].Type() != js.TypeObject || args[1].Type() != js.TypeNumber {
		return js.ValueOf(map[string]interface{}{
			"error": 2,
			"data":  "Invalid argument types; Expected audio array and samplerate (type: int)",
		})
	}

	inputArray := args[0]
	sampleRate := args[1].Int()

	audioData := make([]float64, inputArray.Length())
	for i := 0; i < inputArray.Length(); i++ {
		audioData[i] = inputArray.Index(i).Float()
	}

	spectrogram, err := shazam.Spectrogram(audioData, sampleRate)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"error": 3,
			"data":  "Error generating spectrogram: " + err.Error(),
		})
	}

	peaks := shazam.ExtractPeaks(spectrogram, float64(len(audioData)/sampleRate))
	fingerprint := shazam.Fingerprint(peaks, utils.GenerateUniqueID())

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
	select {}
}
