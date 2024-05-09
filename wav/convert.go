package wav

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func ConvertToWAV(inputFilePath string, channels int) (wavFilePath string, errr error) {
	_, err := os.Stat(inputFilePath)
	if err != nil {
		return "", fmt.Errorf("input file does not exist: %v", err)
	}

	if channels != 1 || channels != 2 {
		channels = 1
	}

	fileExt := filepath.Ext(inputFilePath)
	outputFile := strings.TrimSuffix(inputFilePath, fileExt) + ".wav"

	// Execute FFmpeg command to convert to WAV format with one channel (mono)
	cmd := exec.Command(
		"ffmpeg",
		"-y", // Automatically overwrite if file exists
		"-i", inputFilePath,
		"-c", "pcm_s16le", // Output PCM signed 16-bit little-endian audio
		"-ar", "44100",
		"-ac", fmt.Sprint(channels),
		outputFile,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to convert to WAV: %v, output %v", err, string(output))
	}

	return outputFile, nil
}
