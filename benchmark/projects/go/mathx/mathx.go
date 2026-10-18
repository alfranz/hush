package mathx

// Clamp limits v to the range [lo, hi].
func Clamp(v, lo, hi int) int {
	return max(lo, min(v, hi))
}

// Mean returns the average of xs, or 0 for no values.
func Mean(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	sum := 0.0
	for _, x := range xs {
		sum += x
	}
	return sum / float64(len(xs))
}
