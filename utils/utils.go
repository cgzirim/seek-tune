package utils

import (
	"math/rand"
	"time"
)

func GenerateUniqueID() uint32 {
	rand.Seed(time.Now().UnixNano())
	randomNumber := rand.Uint32()

	return randomNumber
}

func GenerateSongKey(songTitle, songArtist string) string {
	return songTitle + "---" + songArtist
}
