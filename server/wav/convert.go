package wav

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"song-recognition/utils"
	"strings"
)

// ConvertToWAV converts an input audio file to WAV format with specified channels.
func ConvertToWAV(inputFilePath string, channels int) (wavFilePath string, err error) {
	_, err = os.Stat(inputFilePath)
	if err != nil {
		return "", fmt.Errorf("input file does not exist: %v", err)
	}

	if channels < 1 || channels > 2 {
		channels = 1
	}

	fileExt := filepath.Ext(inputFilePath)
	outputFile := strings.TrimSuffix(inputFilePath, fileExt) + ".wav"

	// Output file may already exists. If it does FFmpeg will fail as
	// it cannot edit existing files in-place. Use a temporary file.
	tmpFile := filepath.Join(filepath.Dir(outputFile), "tmp_"+filepath.Base(outputFile))
	defer os.Remove(tmpFile)

	cmd := exec.Command(
		"ffmpeg",
		"-y",
		"-i", inputFilePath,
		"-c", "pcm_s16le",
		"-ar", "44100",
		"-ac", fmt.Sprint(channels),
		tmpFile,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to convert to WAV: %v, output %v", err, string(output))
	}

	// Rename the temporary file to the output file
	err = utils.MoveFile(tmpFile, outputFile)
	if err != nil {
		return "", fmt.Errorf("failed to rename temporary file to output file: %v", err)
	}

	return outputFile, nil
}

// ReformatWAV converts a given WAV file to the specified number of channels,
// either mono (1 channel) or stereo (2 channels).
func ReformatWAV(inputFilePath string, channels int) (reformatedFilePath string, errr error) {
	if channels < 1 || channels > 2 {
		channels = 1
	}

	fileExt := filepath.Ext(inputFilePath)
	outputFile := strings.TrimSuffix(inputFilePath, fileExt) + "rfm.wav"

	cmd := exec.Command(
		"ffmpeg",
		"-y",
		"-i", inputFilePath,
		"-c", "pcm_s16le",
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
