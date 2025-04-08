package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"log/slog"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"song-recognition/db"
	"song-recognition/shazam"
	"song-recognition/spotify"
	"song-recognition/utils"
	"song-recognition/wav"
	"strconv"
	"strings"

	"github.com/fatih/color"
	socketio "github.com/googollee/go-socket.io"
	"github.com/googollee/go-socket.io/engineio"
	"github.com/googollee/go-socket.io/engineio/transport"
	"github.com/googollee/go-socket.io/engineio/transport/polling"
	"github.com/googollee/go-socket.io/engineio/transport/websocket"
	"github.com/mdobak/go-xerrors"
)

const (
	SONGS_DIR = "songs"
)

var yellow = color.New(color.FgYellow)

func find(filePath string) {
	wavInfo, err := wav.ReadWavInfo(filePath)
	if err != nil {
		yellow.Println("Error reading wave info:", err)
		return
	}

	samples, err := wav.WavBytesToSamples(wavInfo.Data)
	if err != nil {
		yellow.Println("Error converting to samples:", err)
		return
	}

	matches, searchDuration, err := shazam.FindMatches(samples, wavInfo.Duration, wavInfo.SampleRate)
	if err != nil {
		yellow.Println("Error finding matches:", err)
		return
	}

	if len(matches) == 0 {
		fmt.Println("\nNo match found.")
		fmt.Printf("\nSearch took: %s\n", searchDuration)
		return
	}

	msg := "Matches:"
	topMatches := matches
	if len(matches) >= 20 {
		msg = "Top 20 matches:"
		topMatches = matches[:20]
	}

	fmt.Println(msg)
	for _, match := range topMatches {
		fmt.Printf("\t- %s by %s, score: %.2f\n",
			match.SongTitle, match.SongArtist, match.Score)
	}

	fmt.Printf("\nSearch took: %s\n", searchDuration)
	topMatch := topMatches[0]
	fmt.Printf("\nFinal prediction: %s by %s , score: %.2f\n",
		topMatch.SongTitle, topMatch.SongArtist, topMatch.Score)
}

func download(spotifyURL string) {
	err := utils.CreateFolder(SONGS_DIR)
	if err != nil {
		err := xerrors.New(err)
		logger := utils.GetLogger()
		ctx := context.Background()
		logMsg := fmt.Sprintf("failed to create directory %v", SONGS_DIR)
		logger.ErrorContext(ctx, logMsg, slog.Any("error", err))
	}

	if strings.Contains(spotifyURL, "album") {
		_, err := spotify.DlAlbum(spotifyURL, SONGS_DIR)
		if err != nil {
			yellow.Println("Error: ", err)
		}
	}

	if strings.Contains(spotifyURL, "playlist") {
		_, err := spotify.DlPlaylist(spotifyURL, SONGS_DIR)
		if err != nil {
			yellow.Println("Error: ", err)
		}
	}

	if strings.Contains(spotifyURL, "track") {
		_, err := spotify.DlSingleTrack(spotifyURL, SONGS_DIR)
		if err != nil {
			yellow.Println("Error: ", err)
		}
	}
}

func serve(protocol, port string) {
	protocol = strings.ToLower(protocol)
	var allowOriginFunc = func(r *http.Request) bool {
		return true
	}

	server := socketio.NewServer(&engineio.Options{
		Transports: []transport.Transport{
			&polling.Transport{
				CheckOrigin: allowOriginFunc,
			},
			&websocket.Transport{
				CheckOrigin: allowOriginFunc,
			},
		},
	})

	server.OnConnect("/", func(socket socketio.Conn) error {
		socket.SetContext("")
		log.Println("CONNECTED: ", socket.ID())

		return nil
	})

	server.OnEvent("/", "totalSongs", handleTotalSongs)
	server.OnEvent("/", "newDownload", handleSongDownload)
	server.OnEvent("/", "newRecording", handleNewRecording)
	server.OnEvent("/", "newFingerprint", handleNewFingerprint)

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

	serveHTTPS := protocol == "https"

	serveHTTP(server, serveHTTPS, port)
}

func serveHTTP(socketServer *socketio.Server, serveHTTPS bool, port string) {
	http.Handle("/socket.io/", socketServer)
	http.Handle("/", http.FileServer(http.Dir("static")))

	if serveHTTPS {
		httpsAddr := ":" + port
		httpsServer := &http.Server{
			Addr: httpsAddr,
			TLSConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
			Handler: socketServer,
		}

		cert_key_default := "/etc/letsencrypt/live/localport.online/privkey.pem"
		cert_file_default := "/etc/letsencrypt/live/localport.online/fullchain.pem"

		cert_key := utils.GetEnv("CERT_KEY", cert_key_default)
		cert_file := utils.GetEnv("CERT_FILE", cert_file_default)
		if cert_key == "" || cert_file == "" {
			log.Fatal("Missing cert")
		}

		log.Printf("Starting HTTPS server on %s\n", httpsAddr)
		if err := httpsServer.ListenAndServeTLS(cert_file, cert_key); err != nil {
			log.Fatalf("HTTPS server ListenAndServeTLS: %v", err)
		}
	}

	log.Printf("Starting HTTP server on port %v", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}
}

func erase(songsDir string) {
	logger := utils.GetLogger()
	ctx := context.Background()

	// wipe db
	dbClient, err := db.NewDBClient()
	if err != nil {
		msg := fmt.Sprintf("Error creating DB client: %v\n", err)
		logger.ErrorContext(ctx, msg, slog.Any("error", err))
	}

	err = dbClient.DeleteCollection("fingerprints")
	if err != nil {
		msg := fmt.Sprintf("Error deleting collection: %v\n", err)
		logger.ErrorContext(ctx, msg, slog.Any("error", err))
	}

	err = dbClient.DeleteCollection("songs")
	if err != nil {
		msg := fmt.Sprintf("Error deleting collection: %v\n", err)
		logger.ErrorContext(ctx, msg, slog.Any("error", err))
	}

	// delete song files
	err = filepath.Walk(songsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			ext := filepath.Ext(path)
			if ext == ".wav" || ext == ".m4a" {
				err := os.Remove(path)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		msg := fmt.Sprintf("Error walking through directory %s: %v\n", songsDir, err)
		logger.ErrorContext(ctx, msg, slog.Any("error", err))
	}

	fmt.Println("Erase complete")
}

func save(path string, force bool) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		fmt.Printf("Error stating path %v: %v\n", path, err)
		return
	}

	if fileInfo.IsDir() {
		err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
			if err != nil {
				fmt.Printf("Error walking the path %v: %v\n", filePath, err)
				return err
			}
			// Process only files, skip directories
			if !info.IsDir() {
				err := saveSong(filePath, force)
				if err != nil {
					fmt.Printf("Error saving song (%v): %v\n", filePath, err)
				}
			}
			return nil
		})
		if err != nil {
			fmt.Printf("Error walking the directory %v: %v\n", path, err)
		}
	} else {
		err := saveSong(path, force)
		if err != nil {
			fmt.Printf("Error saving song (%v): %v\n", path, err)
		}
	}
}

func saveSong(filePath string, force bool) error {
	metadata, err := wav.GetMetadata(filePath)
	if err != nil {
		return err
	}

	durationFloat, err := strconv.ParseFloat(metadata.Format.Duration, 64)
	if err != nil {
		return fmt.Errorf("failed to parse duration to float: %v", err)
	}

	tags := metadata.Format.Tags
	track := &spotify.Track{
		Album:    tags["album"],
		Artist:   tags["artist"],
		Title:    tags["title"],
		Duration: int(math.Round(durationFloat)),
	}

	ytID, err := spotify.GetYoutubeId(*track)
	if err != nil && !force {
		return fmt.Errorf("failed to get YouTube ID for song: %v", err)
	}

	fileName := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	if track.Title == "" {
		// If title is empty, use the file name
		track.Title = fileName
	}

	if track.Artist == "" {
		return fmt.Errorf("no artist found in metadata")
	}

	err = spotify.ProcessAndSaveSong(filePath, track.Title, track.Artist, ytID)
	if err != nil {
		return fmt.Errorf("failed to process or save song: %v", err)
	}

	// Move song in wav format to songs directory
	wavFile := fileName + ".wav"
	sourcePath := filepath.Join(filepath.Dir(filePath), wavFile)
	newFilePath := filepath.Join(SONGS_DIR, wavFile)
	err = utils.MoveFile(sourcePath, newFilePath)
	if err != nil {
		return fmt.Errorf("failed to rename temporary file to output file: %v", err)
	}

	return nil
}
