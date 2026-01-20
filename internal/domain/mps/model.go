package mps

type WordStats struct {
	DaysSinceLastReview int
	AccuracyRate        float64 // 0.0 - 1.0
	TotalReviews        int
	RecentFailures      int
	RecentReviews       int
	Confidence          int     // 1 - 5
	FrequencyScore      float64 // 0.0 - 1.0
}
