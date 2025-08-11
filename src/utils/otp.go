package utils

import (
	"math/rand/v2"
	// Use jwt/v5
)

func GenerateOTPCode() int {
	// From 10000 to 99999
	code := rand.Int32N(90000) + 10000
	return int(code)
}
