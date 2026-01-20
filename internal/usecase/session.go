package usecase

import (
	"context"

	"github.com/sonsonha/eng-noting/internal/domain/review"
	"github.com/sonsonha/eng-noting/internal/domain/session"
	"github.com/sonsonha/eng-noting/internal/domain/word"
)

// SessionUseCase handles session-related business logic
type SessionUseCase struct {
	queueRepo     session.ReviewQueueRepository
	wordStatsRepo word.WordStatsRepository
	reviewRepo    review.ReviewRepository
	mpsService    *MPSService
}

// NewSessionUseCase creates a new SessionUseCase
func NewSessionUseCase(
	queueRepo session.ReviewQueueRepository,
	wordStatsRepo word.WordStatsRepository,
	reviewRepo review.ReviewRepository,
	mpsService *MPSService,
) *SessionUseCase {
	return &SessionUseCase{
		queueRepo:     queueRepo,
		wordStatsRepo: wordStatsRepo,
		reviewRepo:    reviewRepo,
		mpsService:    mpsService,
	}
}

// StartSessionInput represents input for starting a session
type StartSessionInput struct {
	UserID string
}

// StartSessionOutput represents output from starting a session
type StartSessionOutput struct {
	SessionID string
	Items     []session.SessionItem
	Total     int
}

// StartSession starts a new review session
func (uc *SessionUseCase) StartSession(ctx context.Context, input StartSessionInput) (*StartSessionOutput, error) {
	// Rebuild review queue
	if err := uc.rebuildReviewQueue(ctx, input.UserID); err != nil {
		return nil, err
	}

	// Build session from queue
	session, err := uc.buildSession(ctx, input.UserID)
	if err != nil {
		return nil, err
	}

	return &StartSessionOutput{
		SessionID: "", // Will be set by handler if needed
		Items:     session.Items,
		Total:     len(session.Items),
	}, nil
}

// rebuildReviewQueue rebuilds the review queue for a user
func (uc *SessionUseCase) rebuildReviewQueue(ctx context.Context, userID string) error {
	// Load word stats
	stats, err := uc.wordStatsRepo.LoadStats(ctx, userID)
	if err != nil {
		return err
	}

	// Calculate MPS for each word and build queue items
	var queueItems []session.ReviewQueueItem
	for _, stat := range stats {
		mpsInput := CalculateMPSInput{
			WordStats: stat,
		}
		mpsOutput, mpsReason := uc.mpsService.CalculateMPS(mpsInput)

		// Skip low priority words
		if mpsOutput.Score < 30 {
			continue
		}

		queueItems = append(queueItems, session.ReviewQueueItem{
			UserID:        userID,
			WordID:        stat.WordID,
			PriorityScore: mpsOutput.Score,
			Reason:        mpsReason,
		})
	}

	return uc.queueRepo.Rebuild(ctx, userID, queueItems)
}

// buildSession builds a session from the review queue
func (uc *SessionUseCase) buildSession(ctx context.Context, userID string) (*session.Session, error) {
	queueItems, err := uc.queueRepo.GetQueueItems(ctx, userID)
	if err != nil {
		return nil, err
	}

	const (
		MaxCritical = 5
		MaxNormal   = 5
	)

	var critical []session.SessionItem
	var normal []session.SessionItem

	for _, item := range queueItems {
		// Get review stats for this word
		stats, err := uc.reviewRepo.GetStats(ctx, item.WordID)
		if err != nil {
			continue
		}

		// Get last review type
		lastReviewType, _ := uc.reviewRepo.GetLastReviewType(ctx, item.WordID)

		// Calculate review type
		reviewCtx := review.Context{
			MPS:            item.PriorityScore,
			AccuracyRate:   stats.AccuracyRate,
			TotalReviews:   stats.TotalReviews,
			LastReviewType: lastReviewType,
		}
		reviewType := review.SelectType(reviewCtx)

		// Enhance reason with review-specific reason
		reviewReason := review.Reason(reviewCtx, reviewType)
		enhancedReason := item.Reason + ". " + reviewReason

		sessionItem := session.SessionItem{
			WordID:        item.WordID,
			ReviewType:    reviewType,
			PriorityScore: item.PriorityScore,
			Reason:        enhancedReason,
		}

		switch {
		case item.PriorityScore >= 60 && len(critical) < MaxCritical:
			critical = append(critical, sessionItem)
		case item.PriorityScore >= 40 && len(normal) < MaxNormal:
			normal = append(normal, sessionItem)
		}

		if len(critical) == MaxCritical && len(normal) == MaxNormal {
			break
		}
	}

	items := append(critical, normal...)

	return &session.Session{
		UserID: userID,
		Items:  items,
		Index:  0,
	}, nil
}
