package utils

import (
	"encoding/binary"
	"fmt"
)

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
