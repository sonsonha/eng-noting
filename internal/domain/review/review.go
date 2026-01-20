package review

import (
	"context"
	"time"
)

// Review represents a review entity in the domain
type Review struct {
	ID         string
	WordID     string
	UserID     string
	Result     bool
	ReviewType string
	ReviewedAt time.Time
}

// ReviewStats represents aggregated statistics for a word's reviews
type ReviewStats struct {
	WordID         string
	TotalReviews   int
	CorrectReviews int
	LastReviewedAt *time.Time
	AccuracyRate   float64
	MemoryScore    float64
}

// ReviewRepository defines the interface for review persistence
type ReviewRepository interface {
	Create(ctx context.Context, review *Review) error
	GetStats(ctx context.Context, wordID string) (*ReviewStats, error)
	UpdateStats(ctx context.Context, wordID string, result bool) error
	GetLastReviewType(ctx context.Context, wordID string) (string, error)
}
