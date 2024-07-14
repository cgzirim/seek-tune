package utils

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"log/slog"
	"os"
	"song-recognition/models"
	"song-recognition/wav"
	"strings"
	"time"

	"github.com/mdobak/go-xerrors"
)

func DeleteFile(filePath string) error {
	if _, err := os.Stat(filePath); err == nil {
		if err := os.RemoveAll(filePath); err != nil {
			return err
		}
	}
	return nil
}

func CreateFolder(folderPath string) error {
	err := os.MkdirAll(folderPath, 0755)
	if err != nil {
		return err
	}
	return nil
}

func FloatsToBytes(data []float64, bitsPerSample int) ([]byte, error) {
	var byteData []byte

	switch bitsPerSample {
	case 8:
		for _, sample := range data {
			// Convert float to 8-bit unsigned integer
			val := uint8((sample + 1.0) * 127.5)
			byteData = append(byteData, byte(val))
		}
	case 16:
		for _, sample := range data {
			// Convert float to 16-bit signed integer
			val := int16(sample * 32767.0)
			buf := make([]byte, 2)
			binary.LittleEndian.PutUint16(buf, uint16(val))
			byteData = append(byteData, buf...)
		}
	case 24:
		for _, sample := range data {
			// Convert float to 24-bit signed integer
			val := int32(sample * 8388607.0)
			buf := make([]byte, 4)
			binary.LittleEndian.PutUint32(buf, uint32(val)<<8) // Shift by 8 bits to fit 24-bit
			byteData = append(byteData, buf[:3]...)
		}
	case 32:
		for _, sample := range data {
			// Convert float to 32-bit signed integer
			val := int32(sample * 2147483647.0)
			buf := make([]byte, 4)
			binary.LittleEndian.PutUint32(buf, uint32(val))
			byteData = append(byteData, buf...)
		}
	default:
		return nil, fmt.Errorf("unsupported bitsPerSample: %d", bitsPerSample)
	}

	return byteData, nil
}

func ProcessRecording(recData *models.RecordData, saveRecording bool) ([]float64, error) {
	decodedAudioData, err := base64.StdEncoding.DecodeString(recData.Audio)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	fileName := fmt.Sprintf("%04d_%02d_%02d_%02d_%02d_%02d.wav",
		now.Second(), now.Minute(), now.Hour(),
		now.Day(), now.Month(), now.Year(),
	)
	filePath := "tmp/" + fileName

	err = wav.WriteWavFile(filePath, decodedAudioData, recData.SampleRate, recData.Channels, recData.SampleSize)
	if err != nil {
		return nil, err
	}

	reformatedWavFile, err := wav.ReformatWAV(filePath, 1)
	if err != nil {
		return nil, err
	}

	wavInfo, _ := wav.ReadWavInfo(reformatedWavFile)
	samples, _ := wav.WavBytesToSamples(wavInfo.Data)

	if saveRecording {
		logger := GetLogger()
		ctx := context.Background()

		err := CreateFolder("recordings")
		if err != nil {
			err := xerrors.New(err)
			logger.ErrorContext(ctx, "Failed create folder.", slog.Any("error", err))
		}

		newFilePath := strings.Replace(reformatedWavFile, "tmp/", "recordings/", 1)
		err = os.Rename(reformatedWavFile, newFilePath)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to move file.", slog.Any("error", err))
		}
	}

	DeleteFile(fileName)
	DeleteFile(reformatedWavFile)

	return samples, nil
}
