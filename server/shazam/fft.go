package shazam

import (
	"math"
)

// FFT computes the Fast Fourier Transform (FFT) of the input data,
// converting the signal from the time domain to the frequency domain.
// For better understanding, refer to this video: https://www.youtube.com/watch?v=spUNpyF58BY
func FFT(input []float64) []complex128 {
	complexArray := make([]complex128, len(input))
	for i, v := range input {
		complexArray[i] = complex(v, 0)
	}

	fftResult := make([]complex128, len(complexArray))
	copy(fftResult, complexArray)
	return recursiveFFT(fftResult)
}

func recursiveFFT(complexArray []complex128) []complex128 {
	N := len(complexArray)
	if N <= 1 {
		return complexArray
	}

	even := make([]complex128, N/2)
	odd := make([]complex128, N/2)
	for i := 0; i < N/2; i++ {
		even[i] = complexArray[2*i]
		odd[i] = complexArray[2*i+1]
	}

	even = recursiveFFT(even)
	odd = recursiveFFT(odd)

	fftResult := make([]complex128, N)
	for k := 0; k < N/2; k++ {
		t := complex(math.Cos(-2*math.Pi*float64(k)/float64(N)), math.Sin(-2*math.Pi*float64(k)/float64(N)))
		fftResult[k] = even[k] + t*odd[k]
		fftResult[k+N/2] = even[k] - t*odd[k]
	}

	return fftResult
}
