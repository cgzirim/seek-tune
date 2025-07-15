package shazam

import (
	"fmt"
	"song-recognition/models"
	"song-recognition/utils"
	"song-recognition/wav"
)

const (
	maxFreqBits    = 9
	maxDeltaBits   = 14
	targetZoneSize = 5
)

// Fingerprint generates fingerprints from a list of peaks and stores them in an array.
// Each fingerprint consists of an address and a couple.
// The address is a hash. The couple contains the anchor time and the song ID.
func Fingerprint(peaks []Peak, songID uint32) map[uint32]models.Couple {
	fingerprints := map[uint32]models.Couple{}

	for i, anchor := range peaks {
		for j := i + 1; j < len(peaks) && j <= i+targetZoneSize; j++ {
			target := peaks[j]

			address := createAddress(anchor, target)
			anchorTimeMs := uint32(anchor.Time * 1000)

			fingerprints[address] = models.Couple{anchorTimeMs, songID}
		}
	}

	return fingerprints
}

// createAddress generates a unique address for a pair of anchor and target points.
// The address is a 32-bit integer where certain bits represent the frequency of
// the anchor and target points, and other bits represent the time difference (delta time)
// between them. This function combines these components into a single address (a hash).
func createAddress(anchor, target Peak) uint32 {
	anchorFreq := int(real(anchor.Freq))
	targetFreq := int(real(target.Freq))
	deltaMs := uint32((target.Time - anchor.Time) * 1000)

	// Combine the frequency of the anchor, target, and delta time into a 32-bit address
	address := uint32(anchorFreq<<23) | uint32(targetFreq<<14) | deltaMs

	return address
}

func FingerprintAudio(songFilePath string, songID uint32) (map[uint32]models.Couple, error) {
	wavFilePath, err := wav.ConvertToWAV(songFilePath)
	if err != nil {
		return nil, fmt.Errorf("error converting input file to WAV: %v", err)
	}

	wavInfo, err := wav.ReadWavInfo(wavFilePath)
	if err != nil {
		return nil, fmt.Errorf("error reading WAV info: %v", err)
	}

	fingerprint := make(map[uint32]models.Couple)

	spectro, err := Spectrogram(wavInfo.LeftChannelSamples, wavInfo.SampleRate)
	if err != nil {
		return nil, fmt.Errorf("error creating spectrogram: %v", err)
	}

	peaks := ExtractPeaks(spectro, wavInfo.Duration)
	utils.ExtendMap(fingerprint, Fingerprint(peaks, songID))

	if wavInfo.Channels == 2 {
		spectro, err = Spectrogram(wavInfo.RightChannelSamples, wavInfo.SampleRate)
		if err != nil {
			return nil, fmt.Errorf("error creating spectrogram for right channel: %v", err)
		}

		peaks = ExtractPeaks(spectro, wavInfo.Duration)
		utils.ExtendMap(fingerprint, Fingerprint(peaks, songID))
	}

	return fingerprint, nil
}
