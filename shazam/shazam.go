package shazam

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"math/cmplx"
	"math/rand"
	"song-recognition/utils"
	"sort"
	"time"

	"github.com/mjibson/go-dsp/fft"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Constants
const (
	chunkSize    = 4096 // 4KB
	hopSize      = 128
	fuzzFactor   = 2
	bitDepth     = 2
	channels     = 1
	samplingRate = 44100
)

type ChunkTag struct {
	SongTitle  string
	SongArtist string
	YouTubeID  string
	TimeStamp  string
}

func Match(sampleAudio []byte) ([]primitive.M, error) {
	sampleChunks := Chunkify(sampleAudio)
	chunkFingerprints, _ := FingerprintChunks(sampleChunks, nil)

	db, err := utils.NewDbClient()
	if err != nil {
		return nil, fmt.Errorf("error connecting to DB: %d", err)
	}
	defer db.Close()

	var chunkTags = make(map[string]primitive.M)
	var songsTimestamps = make(map[string][]string)
	for _, chunkfgp := range chunkFingerprints {
		listOfChunkTags, err := db.GetChunkTags(chunkfgp)
		if err != nil {
			return nil, fmt.Errorf("error getting chunk data with fingerprint %d: %v", chunkfgp, err)
		}

		for _, chunkTag := range listOfChunkTags {
			timeStamp := fmt.Sprint(chunkTag["timestamp"])
			songKey := fmt.Sprintf("%s by %s", chunkTag["songtitle"], chunkTag["songartist"])

			if songsTimestamps[songKey] == nil {
				songsTimestamps[songKey] = []string{timeStamp}
				chunkTags[songKey] = chunkTag
			} else {
				songsTimestamps[songKey] = append(songsTimestamps[songKey], timeStamp)
			}
		}
	}

	maxMatchCount := 0
	var maxMatch string

	matches := make(map[string][]int)

	for songKey, timestamps := range songsTimestamps {
		differences, err := timeDifference(timestamps)
		if err != nil && err.Error() == "insufficient timestamps" {
			continue
		} else if err != nil {
			return nil, err
		}

		fmt.Printf("%s DIFFERENCES: %d\n", songKey, differences)
		if len(differences) >= 2 {
			matches[songKey] = differences
			if len(differences) > maxMatchCount {
				maxMatchCount = len(differences)
				maxMatch = songKey
			}
		}
	}

	sortedChunkTags := sortMatchesByTimeDifference(matches, chunkTags)

	fmt.Println("SORTED CHUNK TAGS: ", sortedChunkTags)
	fmt.Println("MATCHES: ", matches)
	fmt.Println("MATCH: ", maxMatch)
	fmt.Println()
	return sortedChunkTags, nil
}

func sortMatchesByTimeDifference(matches map[string][]int, chunkTags map[string]primitive.M) []primitive.M {
	type songDifferences struct {
		songKey     string
		differences []int
		sum         int
	}

	var kvPairs []songDifferences
	for songKey, differences := range matches {
		sum := 0
		for _, difference := range differences {
			sum += difference
		}
		kvPairs = append(kvPairs, songDifferences{songKey, differences, sum})
	}

	sort.Slice(kvPairs, func(i, j int) bool {
		return kvPairs[i].sum > kvPairs[j].sum
	})

	var sortedChunkTags []primitive.M
	for _, pair := range kvPairs {
		sortedChunkTags = append(sortedChunkTags, chunkTags[pair.songKey])
	}

	return sortedChunkTags
}

func timeDifference(timestamps []string) ([]int, error) {
	if len(timestamps) < 2 {
		return nil, fmt.Errorf("insufficient timestamps")
	}

	layout := "15:04:05"

	timestampsInSeconds := make([]int, len(timestamps))
	for i, ts := range timestamps {
		parsedTime, err := time.Parse(layout, ts)
		if err != nil {
			return nil, fmt.Errorf("error parsing timestamp %q: %w", ts, err)
		}
		hours := parsedTime.Hour()
		minutes := parsedTime.Minute()
		seconds := parsedTime.Second()
		timestampsInSeconds[i] = (hours * 3600) + (minutes * 60) + seconds
	}

	// sort.Ints(timestampsInSeconds)

	differences := []int{}

	for i := len(timestampsInSeconds) - 1; i >= 1; i-- {
		difference := timestampsInSeconds[i] - timestampsInSeconds[i-1]
		// maxSeconds = 15
		if difference > 0 && difference <= 15 {
			differences = append(differences, difference)
		}
	}

	return differences, nil
}

// Chunkify divides the input audio signal into chunks and calculates the Short-Time Fourier Transform (STFT) for each chunk.
// The function returns a 2D slice containing the STFT coefficients for each chunk.
func Chunkify(audio []byte) [][]complex128 {
	numWindows := len(audio) / (chunkSize - hopSize)
	chunks := make([][]complex128, numWindows)

	// Apply Hamming window function
	window := make([]float64, chunkSize)
	for i := range window {
		window[i] = 0.54 - 0.46*math.Cos(2*math.Pi*float64(i)/float64(chunkSize-1))
	}

	// Perform STFT
	for i := 0; i < numWindows; i++ {
		// Extract current chunk
		start := i * hopSize
		end := start + chunkSize
		if end > len(audio) {
			end = len(audio)
		}

		chunk := make([]complex128, chunkSize)
		for j := start; j < end; j++ {
			chunk[j-start] = complex(float64(audio[j])*window[j-start], 0)
		}

		// Compute FFT
		// chunks[i] = Fft(chunk)
		chunks[i] = fft.FFT(chunk)
	}

	return chunks
}

// FingerprintChunks processes a collection of audio data represented as chunks of complex numbers and
// generates fingerprints for each chunk based on the magnitude of frequency components within specific frequency ranges.
func FingerprintChunks(chunks [][]complex128, chunkTag *ChunkTag) ([]int64, map[int64]ChunkTag) {
	var fingerprintList []int64
	fingerprintMap := make(map[int64]ChunkTag)

	var chunksPerSecond int
	var chunkCount int
	var chunkTime time.Time

	if chunkTag != nil {
		// bytesPerSecond = (samplingRate * bitDepth * channels) / 8
		chunksPerSecond = (chunkSize - hopSize) / samplingRate
		chunksPerSecond = 9
		fmt.Println("CHUNKS PER SECOND: ", chunksPerSecond)
		// if chunkSize == 4096 {
		// 	chunksPerSecond = 10
		// }
		chunkCount = 0
		chunkTime = time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	for _, chunk := range chunks {
		if chunkTag != nil {
			chunkCount++
			if chunkCount == chunksPerSecond {
				chunkCount = 0
				chunkTime = chunkTime.Add(1 * time.Second)
				// fmt.Println(chunkTime.Format("15:04:05"))
			}
		}

		chunkMags := map[string]int{
			"20-60": 0, "60-250": 0, "250-500": 0,
			"500-2000": 0, "2000-4000": 0, "4000-8000": 0, "8000-20000": 0,
		}

		for _, frequency := range chunk {
			magnitude := int(cmplx.Abs(frequency))
			ranges := []struct{ min, max int }{{20, 60}, {60, 250}, {250, 500}, {500, 2000}, {2000, 4000}, {4000, 8000}, {8000, 20001}}

			for _, r := range ranges {
				if magnitude >= r.min && magnitude < r.max &&
					chunkMags[fmt.Sprintf("%d-%d", r.min, r.max)] < magnitude {
					chunkMags[fmt.Sprintf("%d-%d", r.min, r.max)] = magnitude
				}
			}
		}

		// fingerprint := fmt.Sprintf("%d-%d-%d-%d-%d-%d-%d",
		// 	chunkMags["20-60"],
		// 	chunkMags["60-250"],
		// 	chunkMags["250-500"],
		// 	chunkMags["500-2000"],
		// 	chunkMags["2000-4000"],
		// 	chunkMags["4000-8000"],
		// 	chunkMags["8000-20000"])

		// fmt.Println(fingerprint)

		points := [4]int64{
			int64(chunkMags["60-250"]),
			int64(chunkMags["250-500"]),
			int64(chunkMags["500-2000"]),
			int64(chunkMags["2000-4000"])}
		// key := hash1(points[:])
		// fmt.Printf("%s: %v\n", fingerprint, key)

		// points := [6]int64{
		// 	int64(chunkMags["20-60"]),
		// 	int64(chunkMags["60-250"]),
		// 	int64(chunkMags["250-500"]),
		// 	int64(chunkMags["500-2000"]),
		// 	int64(chunkMags["2000-4000"]),
		// 	int64(chunkMags["4000-8000"])}
		key := hash(points[:])

		if chunkTag != nil {
			newSampleTag := *chunkTag
			newSampleTag.TimeStamp = chunkTime.Format("15:04:05")
			fingerprintMap[key] = newSampleTag
		} else {
			fingerprintList = append(fingerprintList, key)
		}
	}

	return fingerprintList, fingerprintMap
}

func hash(values []int64) int64 {
	weight := 100
	var result int64
	for _, value := range values {
		result += (value - (value % fuzzFactor)) * int64(weight)
		weight = weight * weight
	}

	return result
}

func hash1(values []int64) int64 {
	p1, p2, p3, p4 := values[0], values[1], values[2], values[3]
	return (p4-(p4%fuzzFactor))*100000000 +
		(p3-(p3%fuzzFactor))*100000 +
		(p2-(p2%fuzzFactor))*100 +
		(p1 - (p1 % fuzzFactor))
}

func hash2(values []int64) int64 {
	for i := range values {
		values[i] += rand.Int63n(fuzzFactor) - fuzzFactor/2
	}

	var buf []byte
	for _, v := range values {
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(v))
		buf = append(buf, b...)
	}

	hash := sha256.Sum256(buf)

	return int64(binary.BigEndian.Uint64(hash[:8]))
}
