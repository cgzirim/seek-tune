package shazam

import (
	"math"
	"math/cmplx"
)

// fft performs the Fast Fourier Transform on the input signal.
func Fft(complexArray []complex128) []complex128 {
	fftResult := make([]complex128, len(complexArray))
	copy(fftResult, complexArray) // Copy input to result buffer
	return recursiveFFT(fftResult)
}

// recursiveFFT performs the recursive FFT algorithm.
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
		t := cmplx.Exp(-2i * math.Pi * complex(float64(k), 0) / complex(float64(N), 0))
		fftResult[k] = even[k] + t*odd[k]
		fftResult[k+N/2] = even[k] - t*odd[k]
	}

	return fftResult
}
