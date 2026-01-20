package usecase

import (
	"time"

	"github.com/sonsonha/eng-noting/internal/domain"
	"github.com/sonsonha/eng-noting/internal/mps"
)

// MPSService handles Memory Priority Score calculations
type MPSService struct{}

// NewMPSService creates a new MPSService
func NewMPSService() *MPSService {
	return &MPSService{}
}

// CalculateMPSInput represents input for MPS calculation
type CalculateMPSInput struct {
	WordStats domain.WordStats
}

// CalculateMPSOutput represents output from MPS calculation
type CalculateMPSOutput struct {
	Score float64
}

// CalculateMPS calculates the Memory Priority Score for a word
func (s *MPSService) CalculateMPS(input CalculateMPSInput) (CalculateMPSOutput, string) {
	// Convert domain.WordStats to mps.WordStats
	stats := mps.WordStats{
		DaysSinceLastReview: s.daysSinceLastReview(input.WordStats.LastReviewedAt),
		AccuracyRate:         input.WordStats.AccuracyRate,
		TotalReviews:         input.WordStats.TotalReviews,
		RecentFailures:       input.WordStats.RecentFailures,
		RecentReviews:        input.WordStats.RecentReviews,
		Confidence:           input.WordStats.Confidence,
		FrequencyScore:       input.WordStats.FrequencyScore,
	}

	score, reason := mps.CalculateMPS(stats)
	return CalculateMPSOutput{Score: score}, reason
}

// daysSinceLastReview calculates days since last review
func (s *MPSService) daysSinceLastReview(lastReviewedAt *string) int {
	if lastReviewedAt == nil || *lastReviewedAt == "" {
		return 999 // Very high number for never reviewed
	}

	// Parse the time string (assuming RFC3339 format)
	t, err := time.Parse(time.RFC3339, *lastReviewedAt)
	if err != nil {
		// If parsing fails, try other common formats
		t, err = time.Parse("2006-01-02T15:04:05Z07:00", *lastReviewedAt)
		if err != nil {
			return 999 // Default to high number if parsing fails
		}
	}

	days := int(time.Since(t).Hours() / 24)
	if days < 0 {
		return 0
	}
	return days
}
