package utils

import "math"

func roundFloat64(f float64) float64 {
	return math.Floor(f + .5)
}

// RoundFloat64 -
func RoundFloat64(f float64, places int) float64 {
	shift := math.Pow(10, float64(places))
	return roundFloat64(f*shift) / shift
}
