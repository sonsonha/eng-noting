package mps

func CalculateMPS(s WordStats) (float64, string) {
	timeFactor := min(float64(s.DaysSinceLastReview)/7.0, 1.0)
	accuracyFactor := 1.0 - s.AccuracyRate
	confidenceFactor := float64(5-s.Confidence) / 4.0

	failureFactor := 0.0
	if s.RecentReviews > 0 {
		failureFactor = float64(s.RecentFailures) / float64(s.RecentReviews)
	}

	score :=
		(timeFactor * 30) +
			(accuracyFactor * 30) +
			(confidenceFactor * 15) +
			(failureFactor * 15) +
			(s.FrequencyScore * 10)

	reason := generateReason(timeFactor, accuracyFactor, confidenceFactor, failureFactor)

	return clamp(score, 0, 100), reason
}
