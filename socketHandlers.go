package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"song-recognition/shazam"
	"song-recognition/spotify"
	"song-recognition/utils"
	"song-recognition/wav"
	"strings"

	socketio "github.com/googollee/go-socket.io"
	"github.com/mdobak/go-xerrors"
)

func downloadStatus(statusType, message string) string {
	data := map[string]interface{}{"type": statusType, "message": message}
	jsonData, err := json.Marshal(data)
	if err != nil {
		logger := utils.GetLogger()
		ctx := context.Background()
		err := xerrors.New(err)
		logger.ErrorContext(ctx, "failed to marshal data.", slog.Any("error", err))
		return ""
	}
	return string(jsonData)
}

type RecordData struct {
	Audio      string  `json:"audio"`
	Duration   float64 `json:"duration"`
	Channels   int     `json:"channels"`
	SampleRate int     `json:"sampleRate"`
	SampleSize int     `json:"sampleSize"`
}

func handleTotalSongs(socket socketio.Conn) {
	logger := utils.GetLogger()
	ctx := context.Background()

	db, err := utils.NewDbClient()
	if err != nil {
		err := xerrors.New(err)
		logger.ErrorContext(ctx, "error connecting to DB", slog.Any("error", err))
		return
	}
	defer db.Close()

	totalSongs, err := db.TotalSongs()
	if err != nil {
		err := xerrors.New(err)
		logger.ErrorContext(ctx, "Log error getting total songs", slog.Any("error", err))
		return
	}

	socket.Emit("totalSongs", totalSongs)
}

func handleSongDownload(socket socketio.Conn, spotifyURL string) {
	logger := utils.GetLogger()
	ctx := context.Background()

	// Handle album download
	if strings.Contains(spotifyURL, "album") {
		tracksInAlbum, err := spotify.AlbumInfo(spotifyURL)
		if err != nil {
			fmt.Println("log error: ", err)
			if len(err.Error()) <= 25 {
				socket.Emit("downloadStatus", downloadStatus("error", err.Error()))
				logger.Info(err.Error())
			} else {
				err := xerrors.New(err)
				logger.ErrorContext(ctx, "error getting album info", slog.Any("error", err))
			}
			return
		}

		statusMsg := fmt.Sprintf("%v songs found in album.", len(tracksInAlbum))
		socket.Emit("downloadStatus", downloadStatus("info", statusMsg))

		totalTracksDownloaded, err := spotify.DlAlbum(spotifyURL, SONGS_DIR)
		if err != nil {
			socket.Emit("downloadStatus", downloadStatus("error", "Couldn't to download album."))

			err := xerrors.New(err)
			logger.ErrorContext(ctx, "failed to download album.", slog.Any("error", err))
			return
		}

		statusMsg = fmt.Sprintf("%d songs downloaded from album", totalTracksDownloaded)
		socket.Emit("downloadStatus", downloadStatus("success", statusMsg))
	}

	// Handle playlist download
	if strings.Contains(spotifyURL, "playlist") {
		tracksInPL, err := spotify.PlaylistInfo(spotifyURL)
		if err != nil {
			if len(err.Error()) <= 25 {
				socket.Emit("downloadStatus", downloadStatus("error", err.Error()))
				logger.Info(err.Error())
			} else {
				err := xerrors.New(err)
				logger.ErrorContext(ctx, "error getting album info", slog.Any("error", err))
			}
			return
		}

		statusMsg := fmt.Sprintf("%v songs found in playlist.", len(tracksInPL))
		socket.Emit("downloadStatus", downloadStatus("info", statusMsg))

		totalTracksDownloaded, err := spotify.DlPlaylist(spotifyURL, SONGS_DIR)
		if err != nil {
			socket.Emit("downloadStatus", downloadStatus("error", "Couldn't download playlist."))

			err := xerrors.New(err)
			logger.ErrorContext(ctx, "failed to download playlist.", slog.Any("error", err))
			return
		}

		statusMsg = fmt.Sprintf("%d songs downloaded from playlist.", totalTracksDownloaded)
		socket.Emit("downloadStatus", downloadStatus("success", statusMsg))
	}

	// Handle track download
	if strings.Contains(spotifyURL, "track") {
		trackInfo, err := spotify.TrackInfo(spotifyURL)
		if err != nil {
			if len(err.Error()) <= 25 {
				socket.Emit("downloadStatus", downloadStatus("error", err.Error()))
				logger.Info(err.Error())
			} else {
				err := xerrors.New(err)
				logger.ErrorContext(ctx, "error getting album info", slog.Any("error", err))
			}
			return
		}

		// check if track already exist
		db, err := utils.NewDbClient()
		if err != nil {
			fmt.Errorf("Log - error connecting to DB: %d", err)
		}
		defer db.Close()

		song, songExists, err := db.GetSongByKey(utils.GenerateSongKey(trackInfo.Title, trackInfo.Artist))
		if err == nil {
			if songExists {
				statusMsg := fmt.Sprintf(
					"'%s' by '%s' already exists in the database (https://www.youtube.com/watch?v=%s)",
					song.Title, song.Artist, song.YouTubeID)

				socket.Emit("downloadStatus", downloadStatus("error", statusMsg))
				return
			}
		} else {
			err := xerrors.New(err)
			logger.ErrorContext(ctx, "failed to get song by key.", slog.Any("error", err))
		}

		totalDownloads, err := spotify.DlSingleTrack(spotifyURL, SONGS_DIR)
		if err != nil {
			if len(err.Error()) <= 25 {
				socket.Emit("downloadStatus", downloadStatus("error", err.Error()))
				logger.Info(err.Error())
			} else {
				err := xerrors.New(err)
				logger.ErrorContext(ctx, "error getting album info", slog.Any("error", err))
			}
			return
		}

		statusMsg := ""
		if totalDownloads != 1 {
			statusMsg = fmt.Sprintf("'%s' by '%s' failed to download", trackInfo.Title, trackInfo.Artist)
			socket.Emit("downloadStatus", downloadStatus("error", statusMsg))
		} else {
			statusMsg = fmt.Sprintf("'%s' by '%s' was downloaded", trackInfo.Title, trackInfo.Artist)
			socket.Emit("downloadStatus", downloadStatus("success", statusMsg))
		}
	}
}

func handleNewRecording(socket socketio.Conn, recordData string) {
	logger := utils.GetLogger()
	ctx := context.Background()

	var recData RecordData
	if err := json.Unmarshal([]byte(recordData), &recData); err != nil {
		err := xerrors.New(err)
		logger.ErrorContext(ctx, "Failed to unmarshal record data.", slog.Any("error", err))
		return
	}

	// Decode base64 data
	decodedAudioData, err := base64.StdEncoding.DecodeString(recData.Audio)
	if err != nil {
		err := xerrors.New(err)
		logger.ErrorContext(ctx, "failed to decode base64 data.", slog.Any("error", err))
		return
	}

	// Save the decoded data to a file
	channels := recData.Channels
	sampleRate := recData.SampleRate
	bitsPerSample := recData.SampleSize

	samples, err := wav.WavBytesToSamples(decodedAudioData)
	if err != nil {
		err := xerrors.New(err)
		logger.ErrorContext(ctx, "failed to convert decodedData to samples.", slog.Any("error", err))
	}

	/** this operation alters the audio, adding some level of bass to it.
	if sampleRate != 44100 {
		samples, err = shazam.Downsample(samples, sampleRate, 44100)
		if err != nil {
			err := xerrors.New(err)
			logger.ErrorContext(ctx, "failed to downsample.", slog.Any("error", err))
		}
		sampleRate = 44100
	}

	// Save recording
	recordingInBytes, err := utils.FloatsToBytes(samples, bitsPerSample)
	if err != nil {
		err := xerrors.New(err)
		logger.ErrorContext(ctx, "failed to convert bytes.", slog.Any("error", err))
	}
	decodedAudioData = recordingInBytes
	*/

	err = wav.WriteWavFile("blob.wav", decodedAudioData, sampleRate, channels, bitsPerSample)
	if err != nil {
		err := xerrors.New(err)
		logger.ErrorContext(ctx, "failed to write wav file.", slog.Any("error", err))
	}

	matches, err := shazam.FindMatchess(samples, recData.Duration, sampleRate)
	if err != nil {
		err := xerrors.New(err)
		logger.ErrorContext(ctx, "failed to get matches.", slog.Any("error", err))
	}
	fmt.Println("Matches! : ", matches)

	jsonData, err := json.Marshal(matches)
	if len(matches) > 10 {
		jsonData, _ = json.Marshal(matches[:10])
	}

	if err != nil {
		err := xerrors.New(err)
		logger.ErrorContext(ctx, "failed to marshal matches.", slog.Any("error", err))
		return
	}

	socket.Emit("matches", string(jsonData))
}
