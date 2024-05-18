package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"song-recognition/shazam"
	"song-recognition/spotify"
	"song-recognition/utils"
	"song-recognition/wav"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mdobak/go-xerrors"

	socketio "github.com/googollee/go-socket.io"
)

const (
	SONGS_DIR = "songs"
)

func GinMiddleware(allowOrigin string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", allowOrigin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, Content-Length, X-CSRF-Token, Token, session, Origin, Host, Connection, Accept-Encoding, Accept-Language, X-Requested-With")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Request.Header.Del("Origin")

		c.Next()
	}
}

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
	Audio      string `json:"audio"`
	Channels   int    `json:"channels"`
	SampleRate int    `json:"sampleRate"`
	SampleSize int    `json:"sampleSize"`
}

func main() {
	router := gin.New()

	server := socketio.NewServer(nil)

	logger := utils.GetLogger()
	ctx := context.Background()

	server.OnConnect("/", func(socket socketio.Conn) error {
		socket.SetContext("")
		log.Println("CONNECTED: ", socket.ID())

		return nil
	})

	err := spotify.CreateFolder(SONGS_DIR)
	if err != nil {
		err := xerrors.New(err)
		logMsg := fmt.Sprintf("failed to create directory %v", SONGS_DIR)
		logger.ErrorContext(ctx, logMsg, slog.Any("error", err))
	}

	server.OnEvent("/", "totalSongs", func(socket socketio.Conn) {
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
	})

	server.OnEvent("/", "newDownload", func(socket socketio.Conn, spotifyURL string) {
		if len(spotifyURL) == 0 {
			logger.Debug("Spotify URL required.")
			return
		}

		splitURL := strings.Split(spotifyURL, "/")

		if len(splitURL) < 2 {
			logger.Debug("invalid Spotify URL.")
			return
		}

		spotifyID := splitURL[len(splitURL)-1]
		if strings.Contains(spotifyID, "?") {
			spotifyID = strings.Split(spotifyID, "?")[0]
		}

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

		return
	})

	server.OnEvent("/", "record", func(socket socketio.Conn, recordData string) {
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
		fmt.Println(channels, sampleRate, bitsPerSample)

		err = wav.WriteWavFile("blob.wav", decodedAudioData, sampleRate, channels, bitsPerSample)
		if err != nil {
			err := xerrors.New(err)
			logger.ErrorContext(ctx, "failed to write wav file.", slog.Any("error", err))
		}

		samples, err := wav.WavBytesToSamples(decodedAudioData)
		if err != nil {
			err := xerrors.New(err)
			logger.ErrorContext(ctx, "failed to convert decodedData to samples.", slog.Any("error", err))
		}

		matches, err := shazam.FindMatches(samples, 10.0, sampleRate)
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
	})

	server.OnError("/", func(s socketio.Conn, e error) {
		log.Println("meet error:", e)
	})

	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		log.Println("closed", reason)
	})

	go func() {
		if err := server.Serve(); err != nil {
			log.Fatalf("socketio listen error: %s\n", err)
		}
	}()
	defer server.Close()

	router.Use(GinMiddleware("http://localhost:3000"))
	router.GET("/socket.io/*any", gin.WrapH(server))
	router.POST("/socket.io/*any", gin.WrapH(server))

	if err := router.Run(":5000"); err != nil {
		log.Fatal("failed run app: ", err)
	}
}
