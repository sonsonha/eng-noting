package db

import (
	"time"
)

type WordStatsRow struct {
	WordID         string
	LastReviewedAt *time.Time
	AccuracyRate   float64 // 0.0 - 1.0
	TotalReviews   int
	RecentFailures int
	RecentReviews  int
	Confidence     int     // 1 - 5
	FrequencyScore float64 // 0.0 - 1.0 temporary default 0.5
}
