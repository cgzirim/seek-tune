package spotify

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"song-recognition/db"
	"song-recognition/shazam"
	"song-recognition/utils"
	"song-recognition/wav"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/kkdai/youtube/v2"
	"github.com/mdobak/go-xerrors"
)

const DELETE_SONG_FILE = false

var yellow = color.New(color.FgYellow)

func DlSingleTrack(url, savePath string) (int, error) {
	trackInfo, err := TrackInfo(url)
	if err != nil {
		return 0, err
	}

	fmt.Println("Getting track info...")
	time.Sleep(500 * time.Millisecond)
	track := []Track{*trackInfo}

	fmt.Println("Now, downloading track...")
	totalTracksDownloaded, err := dlTrack(track, savePath)
	if err != nil {
		return 0, err
	}

	return totalTracksDownloaded, nil
}

func DlPlaylist(url, savePath string) (int, error) {
	tracks, err := PlaylistInfo(url)
	if err != nil {
		return 0, err
	}

	time.Sleep(1 * time.Second)
	fmt.Println("Now, downloading playlist...")
	totalTracksDownloaded, err := dlTrack(tracks, savePath)
	if err != nil {
		return 0, err
	}

	return totalTracksDownloaded, nil
}

func DlAlbum(url, savePath string) (int, error) {
	tracks, err := AlbumInfo(url)
	if err != nil {
		return 0, err
	}

	time.Sleep(1 * time.Second)
	fmt.Println("Now, downloading album...")
	totalTracksDownloaded, err := dlTrack(tracks, savePath)
	if err != nil {
		return 0, err
	}

	return totalTracksDownloaded, nil
}

func dlTrack(tracks []Track, path string) (int, error) {
	var wg sync.WaitGroup
	var downloadedTracks []string
	var totalTracks int
	logger := utils.GetLogger()
	results := make(chan int, len(tracks))
	numCPUs := runtime.NumCPU()
	semaphore := make(chan struct{}, numCPUs)

	ctx := context.Background()

	db, err := db.NewDBClient()
	if err != nil {
		return 0, err
	}
	defer db.Close()

	for _, t := range tracks {
		wg.Add(1)
		go func(track Track) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() {
				<-semaphore
			}()

			trackCopy := &Track{
				Album:    track.Album,
				Artist:   track.Artist,
				Artists:  track.Artists,
				Duration: track.Duration,
				Title:    track.Title,
			}

			// check if song exists
			keyExists, err := SongKeyExists(utils.GenerateSongKey(trackCopy.Title, trackCopy.Artist))
			if err != nil {
				err := xerrors.New(err)
				logger.ErrorContext(ctx, "error checking song existence", slog.Any("error", err))
			}
			if keyExists {
				logMessage := fmt.Sprintf("'%s' by '%s' already exits.", trackCopy.Title, trackCopy.Artist)
				logger.Info(logMessage)
				return
			}

			ytID, err := getYTID(trackCopy)
			if ytID == "" || err != nil {
				logMessage := fmt.Sprintf("'%s' by '%s' could not be downloaded", trackCopy.Title, trackCopy.Artist)
				logger.ErrorContext(ctx, logMessage, slog.Any("error", xerrors.New(err)))
				return
			}

			trackCopy.Title, trackCopy.Artist = correctFilename(trackCopy.Title, trackCopy.Artist)
			fileName := fmt.Sprintf("%s - %s", trackCopy.Title, trackCopy.Artist)
			filePath := filepath.Join(path, fileName+".m4a")

			err = downloadYTaudio(ytID, path, filePath)
			if err != nil {
				logMessage := fmt.Sprintf("'%s' by '%s' could not be downloaded", trackCopy.Title, trackCopy.Artist)
				logger.ErrorContext(ctx, logMessage, slog.Any("error", xerrors.New(err)))
				return
			}

			err = ProcessAndSaveSong(filePath, trackCopy.Title, trackCopy.Artist, ytID)
			if err != nil {
				logMessage := fmt.Sprintf("Failed to process song ('%s' by '%s')", trackCopy.Title, trackCopy.Artist)
				logger.ErrorContext(ctx, logMessage, slog.Any("error", xerrors.New(err)))
				return
			}

			utils.DeleteFile(filepath.Join(path, fileName+".m4a"))

			wavFilePath := filepath.Join(path, fileName+".wav")

			if err := addTags(wavFilePath, *trackCopy); err != nil {
				logMessage := fmt.Sprintf("Error adding tags: %s", filePath+".wav")
				logger.ErrorContext(ctx, logMessage, slog.Any("error", xerrors.New(err)))

				return
			}

			if DELETE_SONG_FILE {
				utils.DeleteFile(wavFilePath)
			}

			fmt.Printf("'%s' by '%s' was downloaded\n", track.Title, track.Artist)
			downloadedTracks = append(downloadedTracks, fmt.Sprintf("%s, %s", track.Title, track.Artist))
			results <- 1
		}(t)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for range results {
		totalTracks++
	}

	fmt.Println("Total tracks downloaded:", totalTracks)
	return totalTracks, nil

}

/* github.com/kkdai/youtube */
func downloadYTaudio(id, path, filePath string) error {
	dir, err := os.Stat(path)
	if err != nil {
		panic(err)
	}

	if !dir.IsDir() {
		return errors.New("the path is not valid (not a dir)")
	}

	client := youtube.Client{}
	video, err := client.GetVideo(id)
	if err != nil {
		return err
	}

	/*
		itag code: 140, container: m4a, content: audio, bitrate: 128k
		change the FindByItag parameter to 139 if you want smaller files (but with a bitrate of 48k)
		https://gist.github.com/sidneys/7095afe4da4ae58694d128b1034e01e2
	*/
	formats := video.Formats.Itag(140)

	/* in some cases, when attempting to download the audio
	using the library github.com/kkdai/youtube,
	the download fails (and shows the file size as 0 bytes)
	until the second or third attempt. */
	var fileSize int64
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}

	for fileSize == 0 {
		stream, _, err := client.GetStream(video, &formats[0])
		if err != nil {
			return err
		}

		if _, err = io.Copy(file, stream); err != nil {
			return err
		}

		fileSize, _ = GetFileSize(filePath)
	}
	defer file.Close()

	return nil
}

func addTags(file string, track Track) error {
	// Create a temporary file name by appending "2" before the extension
	tempFile := file
	index := strings.Index(file, ".wav")
	if index != -1 {
		baseName := tempFile[:index]       // Filename without extension ('/path/to/title - artist')
		tempFile = baseName + "2" + ".wav" // Temporary filename ('/path/to/title - artist2.wav')
	}

	// Execute FFmpeg command to add metadata tags
	cmd := exec.Command(
		"ffmpeg",
		"-i", file, // Input file path
		"-c", "copy",
		"-metadata", fmt.Sprintf("album_artist=%s", track.Artist),
		"-metadata", fmt.Sprintf("title=%s", track.Title),
		"-metadata", fmt.Sprintf("artist=%s", track.Artist),
		"-metadata", fmt.Sprintf("album=%s", track.Album),
		tempFile, // Output file path (temporary)
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to add tags: %v, output: %s", err, string(out))
	}

	// Rename the temporary file to the original filename
	if err := os.Rename(tempFile, file); err != nil {
		return fmt.Errorf("failed to rename file: %v", err)
	}

	return nil
}

func ProcessAndSaveSong(songFilePath, songTitle, songArtist, ytID string) error {
	dbclient, err := db.NewDBClient()
	if err != nil {
		return err
	}
	defer dbclient.Close()

	wavFilePath, err := wav.ConvertToWAV(songFilePath, 1)
	if err != nil {
		return err
	}

	wavInfo, err := wav.ReadWavInfo(wavFilePath)
	if err != nil {
		return err
	}

	samples, err := wav.WavBytesToSamples(wavInfo.Data)
	if err != nil {
		return fmt.Errorf("error converting wav bytes to float64: %v", err)
	}

	spectro, err := shazam.Spectrogram(samples, wavInfo.SampleRate)
	if err != nil {
		return fmt.Errorf("error creating spectrogram: %v", err)
	}

	songID, err := dbclient.RegisterSong(songTitle, songArtist, ytID)
	if err != nil {
		return err
	}

	peaks := shazam.ExtractPeaks(spectro, wavInfo.Duration)
	fingerprints := shazam.Fingerprint(peaks, songID)

	err = dbclient.StoreFingerprints(fingerprints)
	if err != nil {
		dbclient.DeleteSongByID(songID)
		return fmt.Errorf("error to storing fingerprint: %v", err)
	}

	fmt.Printf("Fingerprint for %v by %v saved in DB successfully\n", songTitle, songArtist)
	return nil
}

func getYTID(trackCopy *Track) (string, error) {
	ytID, err := GetYoutubeId(*trackCopy)
	if ytID == "" || err != nil {
		return "", err
	}

	// Check if YouTube ID exists
	ytidExists, err := YtIDExists(ytID)
	if err != nil {
		return "", fmt.Errorf("error checking YT ID existence: %v", err)
	}

	if ytidExists { // try to get the YouTube ID again
		logMessage := fmt.Sprintf("YouTube ID (%s) exists. Trying again...\n", ytID)
		fmt.Println("WARN: ", logMessage)
		slog.Warn(logMessage)

		ytID, err = GetYoutubeId(*trackCopy)
		if ytID == "" || err != nil {
			return "", err
		}

		ytidExists, err = YtIDExists(ytID)
		if err != nil {
			return "", fmt.Errorf("error checking YT ID existence: %v", err)
		}

		if ytidExists {
			return "", fmt.Errorf("youTube ID (%s) exists", ytID)
		}
	}

	return ytID, nil
}
