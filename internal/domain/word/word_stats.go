package word

import "context"

// WordStats represents statistics needed for MPS calculation
type WordStats struct {
	WordID         string
	LastReviewedAt *string // Will be converted to time when needed
	AccuracyRate   float64
	TotalReviews   int
	RecentFailures int
	RecentReviews  int
	Confidence     int
	FrequencyScore float64
}

// WordStatsRepository defines the interface for word statistics
type WordStatsRepository interface {
	LoadStats(ctx context.Context, userID string) ([]WordStats, error)
}
