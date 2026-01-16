package mps

import "time"

func FromDBRow(
	confidence int,
	accuracy float64,
	totalReviews int,
	lastReviewed *time.Time,
	recentFailures int,
	recentReviews int,
	frequency float64,
) WordStats {
	days := 999
	if lastReviewed != nil {
		days = int(time.Since(*lastReviewed).Hours() / 24)
	}

	return WordStats{
		DaysSinceLastReview: days,
		AccuracyRate:        accuracy,
		TotalReviews:        totalReviews,
		RecentFailures:      recentFailures,
		RecentReviews:       recentReviews,
		Confidence:          confidence,
		FrequencyScore:      frequency,
	}
}
