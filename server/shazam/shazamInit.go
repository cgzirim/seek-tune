package shazam

import (
	"fmt"
	"song-recognition/db"
	"song-recognition/models"
	"song-recognition/utils"
	"sort"
)

type Match1 struct {
	SongID     uint32
	SongTitle  string
	SongArtist string
	YouTubeID  string
	Timestamp  uint32
	Coherency  float64
}

func Search(audioSamples []float64, audioDuration float64, sampleRate int) ([]Match1, error) {
	spectrogram, err := Spectrogram(audioSamples, sampleRate)
	if err != nil {
		return nil, fmt.Errorf("failed to get spectrogram of samples: %v", err)
	}

	peaks := ExtractPeaks(spectrogram, audioDuration)
	fingerprints := Fingerprint(peaks, utils.GenerateUniqueID())

	addresses := make([]uint32, 0, len(fingerprints))
	for address, _ := range fingerprints {
		addresses = append(addresses, address)
	}

	db, err := db.NewDBClient()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	couples, err := db.GetCouples(addresses)
	if err != nil {
		return nil, err
	}

	targetZones := targetZones(couples)
	fmt.Println("TargetZones: ", targetZones)
	matches := timeCoherency(fingerprints, targetZones)

	var matchList []Match1
	for songID, coherency := range matches {
		song, songExists, err := db.GetSongByID(songID)
		if err != nil || !songExists {
			return nil, err
		}

		timestamp := targetZones[songID][0]
		match := Match1{songID, song.Title, song.Artist, song.YouTubeID, timestamp, float64(coherency)}

		matchList = append(matchList, match)
	}

	sort.Slice(matchList, func(i, j int) bool {
		return matchList[i].Coherency > matchList[j].Coherency
	})

	return matchList, nil
}

func targetZones(m map[uint32][]models.Couple) map[uint32][]uint32 {
	songs := make(map[uint32]map[uint32]int)

	for _, couples := range m {
		for _, couple := range couples {
			if _, ok := songs[couple.SongID]; !ok {
				songs[couple.SongID] = make(map[uint32]int)
			}
			songs[couple.SongID][couple.AnchorTimeMs]++
		}
	}
	fmt.Println("couples: ", songs)

	for songID, anchorTimes := range songs {
		for msTime, count := range anchorTimes {
			if count < 5 {
				delete(songs[songID], msTime)
			}
		}
	}
	fmt.Println("anchorTimes: ", songs)

	targetZones := make(map[uint32][]uint32)
	for songID, anchorTimes := range songs {
		for anchorTime, _ := range anchorTimes {
			targetZones[songID] = append(targetZones[songID], anchorTime)
		}
	}

	return targetZones
}

func timeCoherency(record map[uint32]models.Couple, songs map[uint32][]uint32) map[uint32]int {
	// var threshold float64
	matches := make(map[uint32]int)

	for songID, songAnchorTimes := range songs {
		deltas := make(map[float64]int)
		for _, songAnchorTime := range songAnchorTimes {
			for _, recordAnchor := range record {
				recordAnchorTimeMs := float64(recordAnchor.AnchorTimeMs)
				delta := recordAnchorTimeMs - float64(songAnchorTime)
				deltas[delta]++
			}
		}

		// Find the maximum number of time-coherent notes
		var maxOccurrences int
		for _, occurrences := range deltas {
			if occurrences > maxOccurrences {
				maxOccurrences = occurrences
			}
		}

		matches[songID] = maxOccurrences
	}

	// Apply threshold for coherency
	/**
	for songID, coherency := range matches {
		if float64(coherency) < threshold*float64(len(record)) {
			delete(matches, songID) // Remove songs with insufficient coherency
		}
	}
	*/

	return matches
}
