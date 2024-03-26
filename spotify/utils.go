package spotify

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"strings"
)

func EncodeParam(s string) string {
	return url.QueryEscape(s)
}

func ToLowerCase(s string) string {
	var result string
	for _, char := range s {
		if char >= 'A' && char <= 'Z' {
			result += string(char + 32)
		} else {
			result += string(char)
		}
	}

	return result
}

func GetFileSize(file string) (int64, error) {
	fileInfo, err := os.Stat(file)
	if err != nil {
		return 0, err
	}

	size := int64(fileInfo.Size())
	return size, nil
}

func DeleteFile(filePath string) {
	if _, err := os.Stat(filePath); err == nil {
		if err := os.RemoveAll(filePath); err != nil {
			fmt.Println("Error deleting file:", err)
		}
	}
}

// Convert M4A file from stereo to mono
func ConvertM4aToMono(inputFile, outputFile string) ([]byte, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "stream=channels", "-of", "default=noprint_wrappers=1:nokey=1", inputFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error running ffprobe: %v, %v", err, string(output))
	}

	audioBytes, err := ioutil.ReadFile(inputFile)
	if err != nil {
		return nil, fmt.Errorf("error reading input file: %v", err)
	}

	channels := strings.TrimSpace(string(output))
	if channels != "1" {
		// Convert to mono
		cmd = exec.Command("ffmpeg", "-i", inputFile, "-af", "pan=mono|c0=c0", outputFile)
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("error running ffmpeg: %v", err)
		}

		// Resample to 8192 Hz
		// resampledFile := strings.TrimSuffix(outputFile, filepath.Ext(outputFile)) + "_resampled.m4a"
		// cmd = exec.Command("ffmpeg", "-i", outputFile, "-ar", "8192", resampledFile)
		// output, err = cmd.CombinedOutput()
		// if err := cmd.Run(); err != nil {
		// 	return nil, fmt.Errorf("error resampling: %v, %v", err, string(output))
		// }

		audioBytes, err = ioutil.ReadFile(outputFile)
		if err != nil {
			return nil, fmt.Errorf("error reading input file: %v", err)
		}
	}

	return audioBytes, nil
}
