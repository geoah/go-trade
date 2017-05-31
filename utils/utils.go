package utils

import "math"

// TrimFloat64 -
func TrimFloat64(f float64, places int) float64 {
	shift := math.Pow(10, float64(places))
	return math.Floor(f*shift) / shift
}
