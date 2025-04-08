package utils

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
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

func MoveFile(sourcePath string, destinationPath string) error {
	srcFile, err := os.Open(sourcePath)
	if err != nil {
		return err
	}

	destFile, err := os.Create(destinationPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return err
	}

	err = srcFile.Close()
	if err != nil {
		return err
	}

	err = os.Remove(sourcePath)
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
			val := uint8((sample + 1.0) * 127.5)
			byteData = append(byteData, byte(val))
		}
	case 16:
		for _, sample := range data {
			val := int16(sample * 32767.0)
			buf := make([]byte, 2)
			binary.LittleEndian.PutUint16(buf, uint16(val))
			byteData = append(byteData, buf...)
		}
	case 24:
		for _, sample := range data {
			val := int32(sample * 8388607.0)
			buf := make([]byte, 4)
			binary.LittleEndian.PutUint32(buf, uint32(val)<<8) // Shift by 8 bits to fit 24-bit
			byteData = append(byteData, buf[:3]...)
		}
	case 32:
		for _, sample := range data {
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
