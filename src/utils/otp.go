package utils

import (
	"math/rand/v2"
	"strconv"
	// Use jwt/v5
)

func GenerateOTPCode() string {
	// From 10000 to 99999
	code := rand.Int32N(900000) + 100000
	return strconv.Itoa(int(code))
}
