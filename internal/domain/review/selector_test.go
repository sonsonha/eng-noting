package review

import "testing"

func TestNewWordUsesMCQ(t *testing.T) {
	ctx := Context{
		TotalReviews:   0,
		AccuracyRate:   0,
		LastReviewType: "",
	}

	if got := SelectType(ctx); got != "mcq" {
		t.Fatalf("expected mcq, got %s", got)
	}
}

func TestLowAccuracyUsesMCQ(t *testing.T) {
	ctx := Context{
		TotalReviews:   3,
		AccuracyRate:   0.3,
		LastReviewType: "match",
	}

	if got := SelectType(ctx); got != "mcq" {
		t.Fatalf("expected mcq, got %s", got)
	}
}

func TestMediumAccuracyUsesMatch(t *testing.T) {
	ctx := Context{
		TotalReviews:   3,
		AccuracyRate:   0.6,
		LastReviewType: "mcq",
	}

	if got := SelectType(ctx); got != "match" {
		t.Fatalf("expected match, got %s", got)
	}
}

func TestHighAccuracyEscalates(t *testing.T) {
	ctx := Context{
		TotalReviews:   6,
		AccuracyRate:   0.85,
		LastReviewType: "typing",
	}

	if got := SelectType(ctx); got != "fill_blank" {
		t.Fatalf("expected fill_blank, got %s", got)
	}
}

func TestNoRepeatFormat(t *testing.T) {
	ctx := Context{
		TotalReviews:   6,
		AccuracyRate:   0.85,
		LastReviewType: "fill_blank",
	}

	if got := SelectType(ctx); got == "fill_blank" {
		t.Fatal("should not repeat same format")
	}
}
