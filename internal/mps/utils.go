package mps

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func clamp(v, minV, maxV float64) float64 {
	if v < minV {
		return minV
	}
	if v > maxV {
		return maxV
	}
	return v
}
