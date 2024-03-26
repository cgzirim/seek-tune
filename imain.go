package main

import (
	"fmt"
	"path/filepath"
	"song-recognition/shazam"
	"song-recognition/spotify"
	"song-recognition/utils"
	"strings"
)

func matchSong(songPath string) error {
	m4aFileMono := strings.TrimSuffix(songPath, filepath.Ext(songPath)) + "_mono.m4a"
	audioBytes, err := spotify.ConvertM4aToMono(songPath, m4aFileMono)
	if err != nil {
		return fmt.Errorf("error converting M4A file to mono: %v", err)
	}

	chunks := shazam.Chunkify(audioBytes)
	fingerpints, _ := shazam.FingerprintChunks(chunks, nil)

	for _, fingerprint := range fingerpints {
		db, err := utils.NewDbClient()
		if err != nil {
			return fmt.Errorf("error connecting to DB: %d", err)
		}
		chunkData, err := db.GetChunkData(fingerprint)
		if err != nil {
			return fmt.Errorf("error retrieving chunk data: %d", err)
		}
		fmt.Println("CHUNK DATA: ", chunkData)
	}

	return nil
}

func main() {
	// Example usage
	// Open the MP3 file
	// mp3FilePath := "spotifydown.com - These Are The Days.mp3"
	// signal.Process_and_SaveSong(mp3FilePath, "These Are The Days", "lauren Daigle")

	// https://open.spotify.com/track/3vnKyPnHMunE1bMXYQHFHU?si=34a43de5712c4331 - heaven has come
	// https://open.spotify.com/track/6h2vZPWSWsRJ0ps91epUgT?si=7ac5c26041014ea4 - What's going on
	// https://open.spotify.com/track/7zwSMMJkrRJNvxFO9w42nA?si=fa7cef0f7bd14904 - we raise a sound Nosa and 121SELAH
	// https://open.spotify.com/track/52WA7y6ACfdHbzIii6M9iA?si=8aa26d3974394645 - these are the days
	// https://open.spotify.com/track/3ddxe0WYUpNPtSnHgQOad5?si=8c1665c5b1384e9e - I still have faith in you

	spotify.DlSingleTrack("https://open.spotify.com/track/3vnKyPnHMunE1bMXYQHFHU?si=34a43de5712c4331",
		"/home/chigozirim/Documents/my-docs/song-recognition/songs/")

	spotify.DlSingleTrack("https://open.spotify.com/track/6h2vZPWSWsRJ0ps91epUgT?si=7ac5c26041014ea4",
		"/home/chigozirim/Documents/my-docs/song-recognition/songs/")

	spotify.DlSingleTrack("https://open.spotify.com/track/7zwSMMJkrRJNvxFO9w42nA?si=fa7cef0f7bd14904",
		"/home/chigozirim/Documents/my-docs/song-recognition/songs/")

	spotify.DlSingleTrack("https://open.spotify.com/track/52WA7y6ACfdHbzIii6M9iA?si=8aa26d3974394645",
		"/home/chigozirim/Documents/my-docs/song-recognition/songs/")
	spotify.DlSingleTrack("https://open.spotify.com/track/3ddxe0WYUpNPtSnHgQOad5?si=8c1665c5b1384e9e",
		"/home/chigozirim/Documents/my-docs/song-recognition/songs/")

	// spotify.DlPlaylist("https://open.spotify.com/playlist/7EAqBCOVkDZcbccjxZmgjp?si=bbc07260fb784861",
	// 	"/home/chigozirim/Documents/my-docs/song-recognition/songs/")

	// AJR Mix
	// spotify.DlPlaylist("https://open.spotify.com/playlist/37i9dQZF1EIZjJcbmXVBoA?si=35d7d4ba237147cf",
	// 	"/home/chigozirim/Documents/my-docs/song-recognition/songs/")

	// err := matchSong("/home/chigozirim/Documents/my-docs/song-recognition/songs/We Raise A Sound - Nosa.m4a")
	// if err != nil {
	// 	fmt.Println("error matching song: ", err)
	// 	return
	// }
}
