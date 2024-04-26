package spotify

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"song-recognition/shazam"
	"song-recognition/utils"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/kkdai/youtube/v2"
)

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
	results := make(chan int, len(tracks))
	numCPUs := runtime.NumCPU()
	semaphore := make(chan struct{}, numCPUs)

	db, err := utils.NewDbClient()
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
			songExists, err := db.SongExists(trackCopy.Title, trackCopy.Artist, "")
			if err != nil {
				logMessage := fmt.Sprintln("error checking song existence: ", err)
				slog.Error(logMessage)
			}
			if songExists {
				logMessage := fmt.Sprintf("'%s' by '%s' already downloaded\n", trackCopy.Title, trackCopy.Artist)
				slog.Info(logMessage)
				return
			}

			ytID, err := GetYoutubeId(*trackCopy)
			if ytID == "" || err != nil {
				logMessage := fmt.Sprintf("Error (0): '%s' by '%s' could not be downloaded: %s\n", trackCopy.Title, trackCopy.Artist, err)
				slog.Error(logMessage)
				yellow.Printf(logMessage)
				return
			}

			// Check if YouTube ID exists
			ytIdExists, err := db.SongExists("", "", ytID)
			fmt.Printf("%s exists? = %v\n", ytID, ytIdExists)
			if err != nil {
				logMessage := fmt.Sprintln("error checking song existence: ", err)
				slog.Error(logMessage)
			}

			if ytIdExists { // try to get the YouTube ID again
				logMessage := fmt.Sprintf("YouTube ID exists. Trying again: %s\n", ytID)
				fmt.Println("WARN: ", logMessage)
				slog.Warn(logMessage)

				ytID, err = GetYoutubeId(*trackCopy)
				if ytID == "" || err != nil {
					logMessage := fmt.Sprintf("Error (1): '%s' by '%s' could not be downloaded: %s\n", trackCopy.Title, trackCopy.Artist, err)
					slog.Info(logMessage)
					yellow.Printf(logMessage)
					return
				}

				ytIdExists, err := db.SongExists("", "", ytID)
				if err != nil {
					logMessage := fmt.Sprintln("error checking song existence: ", err)
					slog.Error(logMessage)
				}
				if ytIdExists {
					logMessage := fmt.Sprintf("'%s' by '%s' could not be downloaded: YouTube ID (%s) exists\n", trackCopy.Title, trackCopy.Artist, ytID)
					slog.Error(logMessage)
					return
				}
			}

			trackCopy.Title, trackCopy.Artist = correctFilename(trackCopy.Title, trackCopy.Artist)
			fileName := fmt.Sprintf("%s - %s.m4a", trackCopy.Title, trackCopy.Artist)
			filePath := filepath.Join(path, fileName)

			err = getAudio(ytID, path, filePath)
			if err != nil {
				logMessage := fmt.Sprintf("Error (2): '%s' by '%s' could not be downloaded: %s\n", trackCopy.Title, trackCopy.Artist, err)
				yellow.Printf(logMessage)
				slog.Error(logMessage)
				return
			}

			err = processAndSaveSong(filePath, trackCopy.Title, trackCopy.Artist, ytID)
			if err != nil {
				yellow.Println("Error processing audio: ", err)
				logMessage := fmt.Sprintf("Failed to process song ('%s' by '%s'): %s\n", trackCopy.Title, trackCopy.Artist, err)
				slog.Error(logMessage)
				return
			}

			// Consider removing this and deleting the song file after processing
			if err := addTags(filePath, *trackCopy); err != nil {
				yellow.Println("Error adding tags: ", filePath)
				return
			}

			size, _ := GetFileSize(filePath)
			if size < 1 {
				DeleteFile(filePath)
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
func getAudio(id, path, filePath string) error {
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

	/* itag code: 140, container: m4a, content: audio, bitrate: 128k */
	/* change the FindByItag parameter to 139 if you want smaller files (but with a bitrate of 48k) */
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
	tempFile := file
	index := strings.Index(file, ".m4a")
	if index != -1 {
		result := tempFile[:index]       /* filename but with no extension ('/path/to/title - artist') */
		tempFile = result + "2" + ".m4a" /* just a temporary dumb name ('/path/to/title - artist2.m4a') */
	}

	cmd := exec.Command(
		"ffmpeg",
		"-i", file, /* /path/to/title - artist.m4a */
		"-c", "copy",
		"-metadata", fmt.Sprintf("album_artist=%s", track.Artist),
		"-metadata", fmt.Sprintf("title=%s", track.Title),
		"-metadata", fmt.Sprintf("artist=%s", track.Artist),
		"-metadata", fmt.Sprintf("album=%s", track.Album),
		tempFile, /* /path/to/title - artist2.m4a */
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("ERROR FROM CMD:", err)
		fmt.Println("FFMPEG Output:", string(out))
		return err
	}
	// if err := cmd.Run(); err != nil {
	// 	fmt.Println("ERROR FROM CMD: ", err)
	// 	return err
	// }

	/* removes '2' from file name */
	if err := os.Rename(tempFile, file); err != nil {
		return err
	}

	return nil
}

func processAndSaveSong(songFilePath, songTitle, songArtist, ytID string) error {
	db, err := utils.NewDbClient()
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.RegisterSong(songTitle, songArtist, ytID)
	if err != nil {
		return err
	}

	audioBytes, err := convertStereoToMono(songFilePath)
	if err != nil {
		return fmt.Errorf("error converting song to mono: %v", err)
	}

	chunkTag := shazam.ChunkTag{
		SongTitle:  songTitle,
		SongArtist: songArtist,
		YouTubeID:  ytID,
	}

	// Fingerprint song
	chunks := shazam.Chunkify(audioBytes)
	_, fingerprints := shazam.FingerprintChunks(chunks, &chunkTag)

	// Save fingerprints in DB
	for fgp, ctag := range fingerprints {
		err := db.InsertChunkTag(fgp, ctag)
		if err != nil {
			return err
		}
	}

	fmt.Println("Fingerprints saved in MongoDB successfully")
	return nil
}
