package spotify

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"song-recognition/db"
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

func SongKeyExists(key string) (bool, error) {
	db, err := db.NewDBClient()
	if err != nil {
		return false, err
	}
	defer db.Close()

	_, songExists, err := db.GetSongByKey(key)
	if err != nil {
		return false, err
	}

	return songExists, nil
}

func YtIDExists(ytID string) (bool, error) {
	db, err := db.NewDBClient()
	if err != nil {
		return false, err
	}
	defer db.Close()

	_, songExits, err := db.GetSongByYTID(ytID)
	if err != nil {
		return false, err
	}

	return songExits, nil
}

/* fixes some invalid file names (windows is the capricious one) */
func correctFilename(title, artist string) (string, string) {
	if runtime.GOOS == "windows" {
		invalidChars := []byte{'<', '>', '<', ':', '"', '\\', '/', '|', '?', '*'}
		for _, invalidChar := range invalidChars {
			title = strings.ReplaceAll(title, string(invalidChar), "")
			artist = strings.ReplaceAll(artist, string(invalidChar), "")
		}
	} else {
		title = strings.ReplaceAll(title, "/", "\\")
		artist = strings.ReplaceAll(artist, "/", "\\")
	}

	return title, artist
}

func convertStereoToMono(stereoFilePath string) ([]byte, error) {
	fileExt := filepath.Ext(stereoFilePath)
	monoFilePath := strings.TrimSuffix(stereoFilePath, fileExt) + "_mono" + fileExt
	defer os.Remove(monoFilePath)

	// Check the number of channels in the stereo audio
	cmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "stream=channels", "-of", "default=noprint_wrappers=1:nokey=1", stereoFilePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error getting number of channels: %v, %v", err, string(output))
	}
	channels := strings.TrimSpace(string(output))

	audioBytes, err := ioutil.ReadFile(stereoFilePath)
	if err != nil {
		return nil, fmt.Errorf("error reading stereo file: %v", err)
	}

	if channels != "1" {
		// Convert stereo to mono and downsample by 44100/2
		cmd = exec.Command("ffmpeg", "-i", stereoFilePath, "-af", "pan=mono|c0=c0", monoFilePath)
		// cmd = exec.Command("ffmpeg", "-i", stereoFilePath, "-af", "pan=mono|c0=c0", "-ar", "22050", monoFilePath)
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("error converting stereo to mono: %v", err)
		}

		audioBytes, err = ioutil.ReadFile(monoFilePath)
		if err != nil {
			return nil, fmt.Errorf("error reading mono file: %v", err)
		}
	}

	return audioBytes, nil
}
