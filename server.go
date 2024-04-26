package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"song-recognition/shazam"
	"song-recognition/signal"
	"song-recognition/spotify"
	"song-recognition/utils"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pion/webrtc/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"

	socketio "github.com/googollee/go-socket.io"
)

const (
	tmpSongDir = "/home/chigozirim/Documents/my-docs/song-recognition/songs/"
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

type DownloadStatus struct {
	Type    string
	Message string
}

func downloadStatus(msgType, message string) string {
	data := map[string]interface{}{"type": msgType, "message": message}
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return ""
	}
	return string(jsonData)
}

func main() {
	router := gin.New()

	server := socketio.NewServer(nil)

	server.OnConnect("/", func(socket socketio.Conn) error {
		socket.SetContext("")
		log.Println("CONNECTED: ", socket.ID())

		return nil
	})

	server.OnEvent("/", "totalSongs", func(socket socketio.Conn) {
		db, err := utils.NewDbClient()
		if err != nil {
			log.Printf("Error connecting to DB: %v", err)
			return
		}
		defer db.Close()

		totalSongs, err := db.TotalSongs()
		if err != nil {
			log.Println("Log error getting total songs count:", err)
			return
		}

		socket.Emit("totalSongs", totalSongs)
	})

	server.OnEvent("/", "newDownload", func(socket socketio.Conn, spotifyURL string) {
		if len(spotifyURL) == 0 {
			fmt.Println("=> Spotify URL required.")
			return
		}

		splitURL := strings.Split(spotifyURL, "/")

		if len(splitURL) < 2 {
			fmt.Println("=> Please enter the url copied from the spotify client.")
			return
		}

		spotifyID := splitURL[len(splitURL)-1]
		if strings.Contains(spotifyID, "?") {
			spotifyID = strings.Split(spotifyID, "?")[0]
		}

		if strings.Contains(spotifyURL, "album") {
			tracksInAlbum, err := spotify.AlbumInfo(spotifyURL)
			if err != nil {
				fmt.Println("log error: ", err)
				if len(err.Error()) <= 25 {
					socket.Emit("downloadStatus", downloadStatus("error", err.Error()))
				}
				return
			}

			statusMsg := fmt.Sprintf("%v songs found in album.", len(tracksInAlbum))
			socket.Emit("downloadStatus", downloadStatus("info", statusMsg))

			totalTracksDownloaded, err := spotify.DlAlbum(spotifyURL, tmpSongDir)
			if err != nil {
				socket.Emit("downloadStatus", downloadStatus("error", "Couldn't to download album."))
				return
			}

			statusMsg = fmt.Sprintf("%d songs downloaded from album", totalTracksDownloaded)
			socket.Emit("downloadStatus", downloadStatus("success", statusMsg))

		} else if strings.Contains(spotifyURL, "playlist") {
			tracksInPL, err := spotify.PlaylistInfo(spotifyURL)
			if err != nil {
				fmt.Println("log error: ", err)
				if len(err.Error()) <= 25 {
					socket.Emit("downloadStatus", downloadStatus("error", err.Error()))
				}
				return
			}

			statusMsg := fmt.Sprintf("%v songs found in playlist.", len(tracksInPL))
			socket.Emit("downloadStatus", downloadStatus("info", statusMsg))

			totalTracksDownloaded, err := spotify.DlPlaylist(spotifyURL, tmpSongDir)
			if err != nil {
				fmt.Println("log errorr: ", err)
				socket.Emit("downloadStatus", downloadStatus("error", "Couldn't download playlist."))
				return
			}

			statusMsg = fmt.Sprintf("%d songs downloaded from playlist.", totalTracksDownloaded)
			socket.Emit("downloadStatus", downloadStatus("success", statusMsg))

		} else if strings.Contains(spotifyURL, "track") {
			trackInfo, err := spotify.TrackInfo(spotifyURL)
			if err != nil {
				fmt.Println("log error: ", err)
				if len(err.Error()) <= 25 {
					socket.Emit("downloadStatus", downloadStatus("error", err.Error()))
				}
				return
			}

			// check if track already exist
			db, err := utils.NewDbClient()
			if err != nil {
				fmt.Errorf("Log - error connecting to DB: %d", err)
			}
			defer db.Close()

			chunkTag, err := db.GetChunkTagForSong(trackInfo.Title, trackInfo.Artist)
			if err != nil {
				fmt.Println("chunkTag error: ", err)
			}

			if chunkTag != nil {
				statusMsg := fmt.Sprintf(
					"'%s' by '%s' already exists in the database (https://www.youtube.com/watch?v=%s)",
					trackInfo.Title, trackInfo.Artist, chunkTag["youtubeid"])

				fmt.Println("Emitting1")

				socket.Emit("downloadStatus", downloadStatus("error", statusMsg))
				return
			}

			totalDownloads, err := spotify.DlSingleTrack(spotifyURL, tmpSongDir)
			if err != nil {
				statusMsg := fmt.Sprintf("Couldn't download '%s' by '%s'", trackInfo.Title, trackInfo.Artist)
				fmt.Println("Emitting2")
				socket.Emit("downloadStatus", downloadStatus("error", statusMsg))
				return
			}

			statusMsg := ""
			if totalDownloads != 1 {
				statusMsg = fmt.Sprintf("'%s' by '%s' failed to download", trackInfo.Title, trackInfo.Artist)
				fmt.Println("Emitting2")
				socket.Emit("downloadStatus", downloadStatus("error", statusMsg))
			} else {
				statusMsg = fmt.Sprintf("'%s' by '%s' was downloaded", trackInfo.Title, trackInfo.Artist)
				fmt.Println("Emitting3")
				socket.Emit("downloadStatus", downloadStatus("success", statusMsg))
			}

		} else {
			fmt.Println("=> Only Spotify Album/Playlist/Track URL's are supported.")
			return
		}
	})

	server.OnEvent("/", "blob", func(socket socketio.Conn, base64data string) {
		// Decode base64 data
		decodedData, err := base64.StdEncoding.DecodeString(base64data)
		if err != nil {
			fmt.Println("Error: Failed to decode base64 data:", err)
			return
		}

		// Save the decoded data to a file
		sampleRate := 44100
		channels := 1
		bitsPerSample := 16

		err = utils.WriteWavFile("blob.wav", decodedData, sampleRate, channels, bitsPerSample)
		if err != nil {
			fmt.Println("Error: Failed to write wav file: ", err)
		}

		fmt.Println("Audio saved successfully.")

		matches, err := shazam.FindMatches(decodedData)
		if err != nil {
			fmt.Println("Error: Failed to match:", err)
			return
		}

		var matchesChunkTags []primitive.M
		for _, match := range matches {
			matchesChunkTags = append(matchesChunkTags, match.ChunkTag)
		}

		jsonData, err := json.Marshal(matchesChunkTags)

		if len(matchesChunkTags) > 5 {
			jsonData, err = json.Marshal(matchesChunkTags[:5])
		}

		if err != nil {
			fmt.Println("Log error: ", err)
			return
		}

		socket.Emit("matches", string(jsonData))
	})

	server.OnEvent("/", "engage", func(socket socketio.Conn, encodedOffer string) {
		log.Println("Offer received from client ", socket.ID())

		peerConnection := signal.SetupWebRTC(encodedOffer)

		// Allow us to receive 1 audio track
		if _, err := peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio); err != nil {
			fmt.Println("AAAAA")
			panic(err)
		}

		peerConnection.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
			codec := track.Codec()
			if strings.EqualFold(codec.MimeType, webrtc.MimeTypeOpus) {
				fmt.Println("Getting tracks")
				matches, err := signal.MatchSampleAudio(track)
				if err != nil {
					fmt.Println("CCCCC")
					fmt.Println("Error getting matches: ", err)
					return
				}

				jsonData, err := json.Marshal(matches)
				if err != nil {
					fmt.Println("Log error: ", err)
					return
				}

				if len(matches) > 5 {
					jsonData, err = json.Marshal(matches[:5])
				}

				if err != nil {
					fmt.Println("Log error: ", err)
					return
				}

				fmt.Println(string(jsonData))

				socket.Emit("matches", string(jsonData))
				// peerConnection.Close()
			}
		})

		// Set the handler for ICE connection state
		// This will notify you when the peer has connected/disconnected
		peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
			fmt.Printf("Connection State has changed %s \n", connectionState.String())

			if connectionState == webrtc.ICEConnectionStateConnected {
				fmt.Println("WebRTC Connected. Client: ", socket.ID())
			} else if connectionState == webrtc.ICEConnectionStateFailed || connectionState == webrtc.ICEConnectionStateClosed {

				if connectionState == webrtc.ICEConnectionStateFailed {
					fmt.Println("WebRTC connection failed. Client: ", socket.ID())
					socket.Emit("failedToEngage", "")
				}

				// Gracefully shutdown the peer connection
				if closeErr := peerConnection.Close(); closeErr != nil {
					fmt.Println("Gracefully shutdown the peer connection")
					panic(closeErr)
				}

				// os.Exit(0)
			}
		})

		// Emit answer in base64
		socket.Emit("serverEngaged", signal.Encode(*peerConnection.LocalDescription()))
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
