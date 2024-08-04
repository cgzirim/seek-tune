package shazam

import (
	"fmt"
	"math"
	"song-recognition/db"
	"song-recognition/utils"
	"sort"
	"time"
)

type Match struct {
	SongID     uint32
	SongTitle  string
	SongArtist string
	YouTubeID  string
	Timestamp  uint32
	Score      float64
}

// FindMatches processes the audio samples and finds matches in the database
func FindMatches(audioSamples []float64, audioDuration float64, sampleRate int) ([]Match, time.Duration, error) {
	startTime := time.Now()
	logger := utils.GetLogger()

	spectrogram, err := Spectrogram(audioSamples, sampleRate)
	if err != nil {
		return nil, time.Since(startTime), fmt.Errorf("failed to get spectrogram of samples: %v", err)
	}

	peaks := ExtractPeaks(spectrogram, audioDuration)
	fingerprints := Fingerprint(peaks, utils.GenerateUniqueID())

	addresses := make([]uint32, 0, len(fingerprints))
	for address := range fingerprints {
		addresses = append(addresses, address)
	}

	db, err := db.NewDBClient()
	if err != nil {
		return nil, time.Since(startTime), err
	}
	defer db.Close()

	m, err := db.GetCouples(addresses)
	if err != nil {
		return nil, time.Since(startTime), err
	}

	matches := map[uint32][][2]uint32{} // songID -> [(sampleTime, dbTime)]
	timestamps := map[uint32][]uint32{}

	for address, couples := range m {
		for _, couple := range couples {
			matches[couple.SongID] = append(matches[couple.SongID], [2]uint32{fingerprints[address].AnchorTimeMs, couple.AnchorTimeMs})
			timestamps[couple.SongID] = append(timestamps[couple.SongID], couple.AnchorTimeMs)
		}
	}

	scores := analyzeRelativeTiming(matches)

	var matchList []Match
	for songID, points := range scores {
		song, songExists, err := db.GetSongByID(songID)
		if !songExists {
			logger.Info(fmt.Sprintf("song with ID (%v) doesn't exist", songID))
			continue
		}
		if err != nil {
			logger.Info(fmt.Sprintf("failed to get song by ID (%v): %v", songID, err))
			continue
		}

		sort.Slice(timestamps[songID], func(i, j int) bool {
			return timestamps[songID][i] < timestamps[songID][j]
		})

		match := Match{songID, song.Title, song.Artist, song.YouTubeID, timestamps[songID][0], points}
		matchList = append(matchList, match)
	}

	sort.Slice(matchList, func(i, j int) bool {
		return matchList[i].Score > matchList[j].Score
	})

	return matchList, time.Since(startTime), nil
}

// AnalyzeRelativeTiming checks for consistent relative timing and returns a score
func analyzeRelativeTiming(matches map[uint32][][2]uint32) map[uint32]float64 {
	scores := make(map[uint32]float64)
	for songID, times := range matches {
		count := 0
		for i := 0; i < len(times); i++ {
			for j := i + 1; j < len(times); j++ {
				sampleDiff := math.Abs(float64(times[i][0] - times[j][0]))
				dbDiff := math.Abs(float64(times[i][1] - times[j][1]))
				if math.Abs(sampleDiff-dbDiff) < 100 { // Allow some tolerance
					count++
				}
			}
		}
		scores[songID] = float64(count)
	}
	return scores
}
