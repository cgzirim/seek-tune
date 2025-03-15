package utils

import (
	"math/rand"
	"os"
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

func GetEnv(key string, fallback ...string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	if len(fallback) > 0 {
		return fallback[0]
	}
	return ""
}
