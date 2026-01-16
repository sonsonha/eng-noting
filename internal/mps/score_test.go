package mps

import "testing"

func TestHighAccuracyRecentReviewLowScore(t *testing.T) {
	s := WordStats{
		DaysSinceLastReview: 1,
		AccuracyRate:        0.95,
		TotalReviews:        10,
		RecentFailures:      0,
		RecentReviews:       5,
		Confidence:          5,
		FrequencyScore:      0.2,
	}

	score, _ := CalculateMPS(s)
	if score > 30 {
		t.Fatalf("expected low score, got %.2f", score)
	}
}

func TestLowAccuracyLongGapHighScore(t *testing.T) {
	s := WordStats{
		DaysSinceLastReview: 10,
		AccuracyRate:        0.4,
		TotalReviews:        5,
		RecentFailures:      3,
		RecentReviews:       5,
		Confidence:          2,
		FrequencyScore:      0.3,
	}

	score, reason := CalculateMPS(s)
	if score < 60 {
		t.Fatalf("expected high score, got %.2f", score)
	}

	if reason == "" {
		t.Fatal("expected non-empty reason")
	}
}

func TestFailureFactorDominates(t *testing.T) {
	s := WordStats{
		DaysSinceLastReview: 2,
		AccuracyRate:        0.8,
		TotalReviews:        20,
		RecentFailures:      4,
		RecentReviews:       5,
		Confidence:          4,
		FrequencyScore:      0.1,
	}

	_, reason := CalculateMPS(s)
	if reason != "You recently made mistakes with this word" {
		t.Fatalf("unexpected reason: %s", reason)
	}
}

func TestNoReviewsHandledGracefully(t *testing.T) {
	s := WordStats{
		DaysSinceLastReview: 14,
		AccuracyRate:        1.0,
		TotalReviews:        0,
		RecentFailures:      0,
		RecentReviews:       0,
		Confidence:          3,
		FrequencyScore:      0.5,
	}

	score, _ := CalculateMPS(s)
	if score <= 0 {
		t.Fatal("expected non-zero score")
	}
}
