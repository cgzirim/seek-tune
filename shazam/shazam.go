package shazam

import (
	"fmt"
	"math"
	"song-recognition/models"
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

func FindMatches(audioSamples []float64, audioDuration float64, sampleRate int) ([]Match, time.Duration, error) {
	startTime := time.Now()
	logger := utils.GetLogger()

	spectrogram, err := Spectrogram(audioSamples, sampleRate)
	if err != nil {
		return nil, time.Since(startTime), fmt.Errorf("failed to get spectrogram of samples: %v", err)
	}

	peaks := ExtractPeaks(spectrogram, audioDuration)
	fingerprints := Fingerprint(peaks, utils.GenerateUniqueID())

	var sampleCouples []models.Couple
	addresses := make([]uint32, 0, len(fingerprints))
	for address := range fingerprints {
		addresses = append(addresses, address)
		sampleCouples = append(sampleCouples, fingerprints[address])
	}

	db, err := utils.NewDbClient()
	if err != nil {
		return nil, time.Since(startTime), err
	}
	defer db.Close()

	couplesMap, err := db.GetCouples(addresses)
	if err != nil {
		return nil, time.Since(startTime), err
	}

	// Count occurrences of each couple to derive potential target zones
	coupleCounts := make(map[uint32]map[uint32]int)
	for _, couples := range couplesMap {
		for _, couple := range couples {
			key := (couple.SongID << 32) | uint32(couple.AnchorTimeMs)
			if _, exists := coupleCounts[couple.SongID]; !exists {
				coupleCounts[couple.SongID] = make(map[uint32]int)
			}
			coupleCounts[couple.SongID][key]++
		}
	}

	// Filter target zones with targets (couples) meeting or exceeding the threshold
	threshold := 4
	filteredCouples := make(map[uint32][]models.Couple)
	for songID, counts := range coupleCounts {
		for key, count := range counts {
			if count >= threshold {
				filteredCouples[songID] = append(filteredCouples[songID], models.Couple{
					AnchorTimeMs: key & 0xFFFFFFFF,
					SongID:       songID,
				})
			}
		}
	}

	// Score matches by calculating mean absolute difference
	var matches []Match
	for songID, songCouples := range filteredCouples {
		song, songExists, err := db.GetSongByID(songID)
		if err != nil {
			logger.Info(fmt.Sprintf("failed to get song by ID (%v): %v", songID, err))
			continue
		}
		if !songExists {
			logger.Info(fmt.Sprintf("song with ID (%v) doesn't exist", songID))
			continue
		}

		m_a_d := meanAbsoluteDifference(songCouples, sampleCouples)

		tstamp := songCouples[len(songCouples)-1].AnchorTimeMs
		match := Match{songID, song.Title, song.Artist, song.YouTubeID, tstamp, m_a_d}
		matches = append(matches, match)
	}

	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Score > matches[j].Score
	})

	// TODO: hanld case when there's no match for cmdHandlers

	return matches, time.Since(startTime), nil
}

func meanAbsoluteDifference(A, B []models.Couple) float64 {
	minLen := len(A)
	if len(B) < minLen {
		minLen = len(B)
	}

	var sumDiff float64
	for i := 0; i < minLen; i++ {
		diff := math.Abs(float64(A[i].AnchorTimeMs - B[i].AnchorTimeMs))
		sumDiff += diff
	}

	meanAbsDiff := sumDiff / float64(minLen)
	return meanAbsDiff
}

// Function to calculate Dynamic Time Warping distance
func dynamicTimeWarping(A, B []models.Couple) float64 {
	lenA := len(A)
	lenB := len(B)

	// Create a 2D array to store DTW distances
	dtw := make([][]float64, lenA+1)
	for i := range dtw {
		dtw[i] = make([]float64, lenB+1)
		for j := range dtw[i] {
			dtw[i][j] = math.Inf(1)
		}
	}
	dtw[0][0] = 0

	for i := 1; i <= lenA; i++ {
		for j := 1; j <= lenB; j++ {
			cost := math.Abs(float64(A[i-1].AnchorTimeMs - B[j-1].AnchorTimeMs))
			dtw[i][j] = cost + math.Min(math.Min(dtw[i-1][j], dtw[i][j-1]), dtw[i-1][j-1])
		}
	}

	return dtw[lenA][lenB]
}
