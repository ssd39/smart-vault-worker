package utils

// #include "random.h"
import "C"

func GenerateRandom64Bytes() []byte {
	var out []byte
	var x C.uint16_t
	var retry C.int = 1
	for i := 0; i < 64; i++ {
		C.rdrand_16(&x, retry)
		out = append(out, byte(x))
	}
	return out
}
