package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"song-recognition/spotify"
	"song-recognition/utils"
	"strconv"
	"strings"

	"github.com/mdobak/go-xerrors"

	socketio "github.com/googollee/go-socket.io"
	"github.com/googollee/go-socket.io/engineio"
	"github.com/googollee/go-socket.io/engineio/transport"
	"github.com/googollee/go-socket.io/engineio/transport/polling"
	"github.com/googollee/go-socket.io/engineio/transport/websocket"
)

const (
	SONGS_DIR = "songs"
)

var allowOriginFunc = func(r *http.Request) bool {
	return true
}

func main() {

	err := spotify.CreateFolder(SONGS_DIR)
	if err != nil {
		err := xerrors.New(err)
		logger := utils.GetLogger()
		ctx := context.Background()
		logMsg := fmt.Sprintf("failed to create directory %v", SONGS_DIR)
		logger.ErrorContext(ctx, logMsg, slog.Any("error", err))
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

	SERVE_HTTPS := strings.ToLower(utils.GetEnv("SERVE_HTTPS", "true"))
	serveHTTPS, err := strconv.ParseBool(SERVE_HTTPS)
	if err != nil {
		log.Fatalf("Error converting string to bool: %v", err)
	}

	serveHTTP(server, serveHTTPS)
}

func serveHTTP(socketServer *socketio.Server, serveHTTPS bool) {
	http.Handle("/socket.io/", socketServer)

	if serveHTTPS {
		httpsAddr := ":443"
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

	log.Printf("Starting HTTP server on port 5000")
	if err := http.ListenAndServe(":5000", nil); err != nil {
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}
}
