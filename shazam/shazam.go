package shazam

import (
	"fmt"
	"math"
	"song-recognition/models"
	"song-recognition/utils"
	"sort"
)

type Match struct {
	SongID     uint32
	SongTitle  string
	SongArtist string
	YouTubeID  string
	Timestamp  uint32
	Score      float64
}

func FindMatches(audioSamples []float64, audioDuration float64, sampleRate int) ([]Match, error) {
	logger := utils.GetLogger()

	spectrogram, err := Spectrogram(audioSamples, sampleRate)
	if err != nil {
		return nil, fmt.Errorf("failed to get spectrogram of samples: %v", err)
	}

	peaks := ExtractPeaks(spectrogram, audioDuration)
	fingerprints := Fingerprint(peaks, utils.GenerateUniqueID())
	fmt.Println("peaks len: ", len(peaks))

	addresses := make([]uint32, 0, len(fingerprints))
	for address, _ := range fingerprints {
		addresses = append(addresses, address)
	}

	db, err := utils.NewDbClient()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	m, err := db.GetCouples(addresses)
	if err != nil {
		return nil, err
	}

	matches := map[uint32]map[uint32]models.Couple{}
	timestamps := map[uint32]uint32{}

	for address, couples := range m {
		for _, couple := range couples {

			if _, ok := matches[couple.SongID]; !ok {
				matches[couple.SongID] = map[uint32]models.Couple{}
				timestamps[couple.SongID] = couple.AnchorTimeMs
			}

			matches[couple.SongID][address] = couple
		}
	}

	scores := map[uint32]float64{}
	for songID, couples := range matches {
		song, songExists, err := db.GetSongByID(songID)
		if err != nil || !songExists {
			// log error
			fmt.Println("Continuing")
			continue
		}
		fmt.Printf("Song: %v, Scores:\n", song.Title)

		scores[songID] = matchScore(fingerprints, couples)
		fmt.Println("------------------------------------")
	}

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

		fmt.Printf("Song: %v, Score: %v\n", song.Title, points)
		fmt.Println("====================================")
		match := Match{songID, song.Title, song.Artist, song.YouTubeID, timestamps[songID], points}
		matchList = append(matchList, match)
	}

	sort.Slice(matchList, func(i, j int) bool {
		return matchList[i].Score > matchList[j].Score
	})

	fmt.Println("MatchList len: ", len(matchList))
	return matchList, nil
}

// MatchScore computes a match score between the two transformed audio samples (into a list of Key + TableValue)
func matchScore(sample, match map[uint32]models.Couple) float64 {
	// Will hold a list of points (time in the sample sound file, time in the matched database sound file)
	points := [2][]float64{}
	matches := 0.0
	for k, sampleValue := range sample {
		if matchValue, ok := match[k]; ok {
			points[0] = append(points[0], float64(sampleValue.AnchorTimeMs))
			points[1] = append(points[1], float64(matchValue.AnchorTimeMs))
			matches++
		}
	}
	corr := correlation(points[0], points[1])
	fmt.Printf("Score (%v * %v * %v): %v\n", corr, corr, matches, corr*corr*matches)
	return corr * corr * matches
}

// Correlation computes the correlation between 2 series of points
// the length used is x's
func correlation(x []float64, y []float64) float64 {
	n := len(x)
	meanX, meanY := Avg(x[:n]), Avg(y[:n])

	sXY := 0.0
	sX := 0.0
	sY := 0.0

	for i, xp := range x {
		dx := xp - meanX
		dy := y[i] - meanY

		sX += dx * dx
		sY += dy * dy

		sXY += dx * dy
	}

	if sX == 0 || sY == 0 {
		return 0
	}

	return sXY / (math.Sqrt(sX) * math.Sqrt(sY))
}

// Avg computes the average of the given array
func Avg(arr []float64) float64 {
	if len(arr) == 0 {
		return 0
	}

	sum := 0.0
	for _, v := range arr {
		sum += v
	}

	return sum / float64(len(arr))
}
