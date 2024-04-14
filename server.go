package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"song-recognition/shazam"
	"song-recognition/signal"
	"song-recognition/spotify"
	"song-recognition/utils"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media/oggwriter"

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

func main() {
	router := gin.New()

	server := socketio.NewServer(nil)

	server.OnConnect("/", func(socket socketio.Conn) error {
		socket.SetContext("")
		log.Println("CONNECTED: ", socket.ID())

		return nil
	})

	server.OnEvent("/", "initOffer", func(s socketio.Conn, initEncodedOffer string) {
		log.Println("initOffer: ", initEncodedOffer)

		peerConnection := signal.SetupWebRTC(initEncodedOffer)
		s.Emit("initAnswer", signal.Encode(*peerConnection.LocalDescription()))
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
				return
			}

			socket.Emit("albumStat", fmt.Sprintf("%v songs found in album.", len(tracksInAlbum)))

			totalTracksDownloaded, err := spotify.DlAlbum(spotifyURL, tmpSongDir)
			if err != nil {
				socket.Emit("downloadStatus", fmt.Sprintf("Failed to download album."))
				return
			}

			socket.Emit("downloadStatus", fmt.Sprintf("%d songs downloaded from album", totalTracksDownloaded))

		} else if strings.Contains(spotifyURL, "playlist") {
			tracksInPL, err := spotify.PlaylistInfo(spotifyURL)
			if err != nil {
				fmt.Println("log error: ", err)
				return
			}

			socket.Emit("playlistStat", fmt.Sprintf("%v songs found in playlist.", len(tracksInPL)))

			totalTracksDownloaded, err := spotify.DlPlaylist(spotifyURL, tmpSongDir)
			if err != nil {
				fmt.Println("log errorr: ", err)
				socket.Emit("downloadStatus", fmt.Sprintf("Failed to download playlist."))
				return
			}

			socket.Emit("downloadStatus", fmt.Sprintf("%d songs downloaded from playlist", totalTracksDownloaded))

		} else if strings.Contains(spotifyURL, "track") {
			trackInfo, err := spotify.TrackInfo(spotifyURL)
			if err != nil {
				fmt.Println("log error: ", err)
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
				socket.Emit("downloadStatus", fmt.Sprintf(
					"'%s' by '%s' already exists in the database (https://www.youtube.com/watch?v=%s)",
					trackInfo.Title, trackInfo.Artist, chunkTag["youtubeid"]))
				return
			}

			totalDownloads, err := spotify.DlSingleTrack(spotifyURL, tmpSongDir)
			if err != nil {
				socket.Emit("downloadStatus", fmt.Sprintf("Failed to download '%s' by '%s'", trackInfo.Title, trackInfo.Artist))
				return
			}

			if totalDownloads != 1 {
				socket.Emit("downloadStatus", fmt.Sprintf("'%s' by '%s' failed to download", trackInfo.Title, trackInfo.Artist))
			} else {
				socket.Emit("downloadStatus", fmt.Sprintf("'%s' by '%s' was downloaded", trackInfo.Title, trackInfo.Artist))
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
		err = ioutil.WriteFile("recorded_audio.ogg", decodedData, 0644)
		if err != nil {
			fmt.Println("Error: Failed to write file to disk:", err)
			return
		}

		fmt.Println("Audio saved successfully.")

		matches, err := shazam.Match(decodedData)
		if err != nil {
			fmt.Println("Error: Failed to match:", err)
			return
		}

		jsonData, err := json.Marshal(matches)

		if len(matches) > 5 {
			jsonData, err = json.Marshal(matches[:5])
		}

		if err != nil {
			fmt.Println("Log error: ", err)
			return
		}

		socket.Emit("matches", string(jsonData))

		fmt.Println("BLOB: ", matches)
	})

	server.OnEvent("/", "engage", func(s socketio.Conn, encodedOffer string) {
		log.Println("engage: ", encodedOffer)

		peerConnection := signal.SetupWebRTC(encodedOffer)

		// Allow us to receive 1 audio track
		if _, err := peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio); err != nil {
			panic(err)
		}

		// Set a handler for when a new remote track starts, this handler saves buffers to disk as
		// an Ogg file.
		oggFile, err := oggwriter.New("output.ogg", 48000, 1)
		if err != nil {
			panic(err)
		}

		peerConnection.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
			codec := track.Codec()
			if strings.EqualFold(codec.MimeType, webrtc.MimeTypeOpus) {
				// fmt.Println("Got Opus track, saving to disk as output.opus (44.1 kHz, 1 channel)")
				// signal.SaveToDisk(oggFile, track)

				matches, err := signal.MatchSampleAudio(track)
				if err != nil {
					panic(err)
				}

				jsonData, err := json.Marshal(matches)

				if len(matches) > 5 {
					jsonData, err = json.Marshal(matches[:5])
				}

				if err != nil {
					fmt.Println("Log error: ", err)
					return
				}

				fmt.Println(string(jsonData))

				s.Emit("matches", string(jsonData))
				peerConnection.Close()
			}
		})

		// Set the handler for ICE connection state
		// This will notify you when the peer has connected/disconnected
		peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
			fmt.Printf("Connection State has changed %s \n", connectionState.String())

			if connectionState == webrtc.ICEConnectionStateConnected {
				fmt.Println("Ctrl+C the remote client to stop the demo")
			} else if connectionState == webrtc.ICEConnectionStateFailed || connectionState == webrtc.ICEConnectionStateClosed {
				if closeErr := oggFile.Close(); closeErr != nil {
					panic(closeErr)
				}

				fmt.Println("Done writing media files")

				// Gracefully shutdown the peer connection
				if closeErr := peerConnection.Close(); closeErr != nil {
					panic(closeErr)
				}

				// os.Exit(0)
			}
		})

		// Emit answer in base64
		s.Emit("serverEngaged", signal.Encode(*peerConnection.LocalDescription()))
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
