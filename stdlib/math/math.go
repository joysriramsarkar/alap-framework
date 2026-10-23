// Package math provides standard mathematical operations and constants for NilLang.
package math

import (
	"math"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

const (
	PI = math.Pi
	E  = math.E
)

func Abs(x float64) float64    { return math.Abs(x) }
func Sqrt(x float64) float64   { return math.Sqrt(x) }
func Floor(x float64) float64  { return math.Floor(x) }
func Ceil(x float64) float64   { return math.Ceil(x) }
func Round(x float64) float64  { return math.Round(x) }
func Sin(x float64) float64    { return math.Sin(x) }
func Cos(x float64) float64    { return math.Cos(x) }
func Tan(x float64) float64    { return math.Tan(x) }
func Pow(x, y float64) float64 { return math.Pow(x, y) }
func Log(x float64) float64    { return math.Log(x) }
func Log10(x float64) float64  { return math.Log10(x) }

func Min(a, b float64) float64 {
	return math.Min(a, b)
}

func Max(a, b float64) float64 {
	return math.Max(a, b)
}

func Clamp(val, min, max float64) float64 {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

// Random returns a pseudo-random number in [0.0, 1.0).
func Random() float64 {
	return rand.Float64()
}

// RandomInt returns a random integer in [min, max].
func RandomInt(min, max int64) int64 {
	if min >= max {
		return min
	}
	return min + rand.Int63n(max-min+1)
}
