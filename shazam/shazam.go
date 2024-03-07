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
)

// Constants
const (
	chunkSize    = 4096 // 4KB
	fuzzFactor   = 2
	bitDepth     = 2
	channels     = 1
	samplingRate = 44100
)

// AudioInfo contains details about the audio data.
type AudioInfo struct {
	SongName     string
	SongArtist   string
	BitDepth     int
	Channels     int
	SamplingRate int
	TimeStamp    string // TimeStamp for the chunk
}

func Match(sampleAudio []byte) (string, error) {
	sampleChunks := Chunkify(sampleAudio)
	chunkFingerprints, _ := FingerprintChunks(sampleChunks, nil)

	db, err := utils.NewDbClient()
	if err != nil {
		return "", fmt.Errorf("error connecting to DB: %d", err)
	}
	defer db.Close()

	var results = make(map[string][]string)
	for _, chunkfgp := range chunkFingerprints {
		listOfChunkData, err := db.GetChunkData(chunkfgp)
		if err != nil {
			return "", fmt.Errorf("error getting chunk data with fingerpring %d: %v", chunkfgp, err)
		}

		for _, chunkData := range listOfChunkData {
			timeStamp := fmt.Sprint(chunkData["timestamp"])
			songKey := fmt.Sprintf("%s by %s", chunkData["songname"], chunkData["songartist"])

			if results[songKey] == nil {
				results[songKey] = []string{timeStamp}
			} else {
				results[songKey] = append(results[songKey], timeStamp)
			}
		}
	}

	fmt.Println("Results: ", results)

	maxMatchCount := 0
	var maxMatch string

	for songKey, timestamps := range results {
		differences, err := timeDifference(timestamps)
		if err != nil && err.Error() == "insufficient timestamps" {
			continue
		} else if err != nil {
			return "", err
		}

		fmt.Printf("%s DIFFERENCES: %d\n", songKey, differences)
		if len(differences) >= 2 {
			if len(differences) > maxMatchCount {
				maxMatchCount = len(differences)
				maxMatch = songKey
			}
		}
	}

	fmt.Println("MATCH: ", maxMatch)
	return "", nil
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

	sort.Ints(timestampsInSeconds)
	fmt.Println("timeStampsInSeconds: ", timestampsInSeconds)

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
	const hopSize = 32
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
		chunks[i] = Fft(chunk)
	}

	return chunks
}

// FingerprintChunks processes a collection of audio data represented as chunks of complex numbers and
// generates fingerprints for each chunk based on the magnitude of frequency components within specific frequency ranges.
func FingerprintChunks(chunks [][]complex128, audioInfo *AudioInfo) ([]int64, map[int64]AudioInfo) {
	var fingerprintList []int64
	fingerprintMap := make(map[int64]AudioInfo)

	var bytesPerSecond, chunksPerSecond int
	var chunkCount int
	var chunkTime time.Time

	if audioInfo != nil {
		bytesPerSecond = (samplingRate * bitDepth * channels) / 8
		chunksPerSecond = bytesPerSecond / chunkSize
		// if chunkSize == 4096 {
		// 	chunksPerSecond = 10
		// }
		chunkCount = 0
		chunkTime = time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	for _, chunk := range chunks {
		if audioInfo != nil {
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
		key := hash1(points[:])
		// fmt.Printf("%s: %v\n", fingerprint, key)

		if audioInfo != nil {
			newAudioInfo := *audioInfo
			newAudioInfo.TimeStamp = chunkTime.Format("15:04:05")
			fingerprintMap[key] = newAudioInfo
		} else {
			fingerprintList = append(fingerprintList, key)
		}
	}

	return fingerprintList, fingerprintMap
}

func hash(values []int64) int64 {
	if len(values) != 7 {
		return 0 // Handle invalid input length
	}

	var result int64
	for i := 0; i < len(values); i++ {
		roundedValue := values[i] - (values[i] % fuzzFactor)
		weight := int64(math.Pow10(len(values) - i - 1))
		result += roundedValue * weight
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
