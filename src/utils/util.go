package utils

func NormallizePercent(value float32) float32 {
	return max(min(value, 100), 0)
}
