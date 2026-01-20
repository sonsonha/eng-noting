package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/sonsonha/eng-noting/internal/domain"
)

// ReviewUseCase handles review-related business logic
type ReviewUseCase struct {
	reviewRepo domain.ReviewRepository
	wordRepo   domain.WordRepository
}

// NewReviewUseCase creates a new ReviewUseCase
func NewReviewUseCase(reviewRepo domain.ReviewRepository, wordRepo domain.WordRepository) *ReviewUseCase {
	return &ReviewUseCase{
		reviewRepo: reviewRepo,
		wordRepo:   wordRepo,
	}
}

// SubmitReviewInput represents input for submitting a review
type SubmitReviewInput struct {
	UserID     string
	WordID     string
	Result     bool
	ReviewType string
}

// SubmitReviewOutput represents output from submitting a review
type SubmitReviewOutput struct {
	Success bool
}

// SubmitReview submits a review for a word
func (uc *ReviewUseCase) SubmitReview(ctx context.Context, input SubmitReviewInput) (*SubmitReviewOutput, error) {
	// Verify word belongs to user
	word, err := uc.wordRepo.GetByID(ctx, input.WordID, input.UserID)
	if err != nil {
		return nil, err
	}

	if word.UserID != input.UserID {
		return nil, ErrForbidden
	}

	review := &domain.Review{
		ID:         uuid.NewString(),
		WordID:     input.WordID,
		UserID:     input.UserID,
		Result:     input.Result,
		ReviewType: input.ReviewType,
	}

	if err := uc.reviewRepo.Create(ctx, review); err != nil {
		return nil, err
	}

	// Update review statistics
	if err := uc.reviewRepo.UpdateStats(ctx, input.WordID, input.Result); err != nil {
		return nil, err
	}

	return &SubmitReviewOutput{Success: true}, nil
}
