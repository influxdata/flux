package arrowutil_test

import (
	"math/rand"
	"strings"
)

func generateInt64() int64 {
	return int64(rand.Intn(201) - 100)
}

func generateUint64() uint64 {
	return uint64(rand.Intn(201) - 100)
}

func generateFloat64() float64 {
	return rand.NormFloat64() * 50
}

func generateBoolean() bool {
	return rand.Intn(2) != 0
}

func generateString() string {
	var buf strings.Builder
	for i := 0; i < 3; i++ {
		chars := 62
		if i == 0 {
			chars = 52
		}
		switch n := rand.Intn(chars); {
		case n >= 0 && n < 26:
			buf.WriteByte('A' + byte(n))
		case n >= 26 && n < 52:
			buf.WriteByte('a' + byte(n-26))
		case n >= 52:
			buf.WriteByte('0' + byte(n-52))
		}
	}
	return buf.String()
}
