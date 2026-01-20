package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/sonsonha/eng-noting/internal/domain"
)

// ReviewRepository implements domain.ReviewRepository using PostgreSQL
type ReviewRepository struct {
	db *sql.DB
}

// NewReviewRepository creates a new ReviewRepository
func NewReviewRepository(db *sql.DB) *ReviewRepository {
	return &ReviewRepository{db: db}
}

// Create creates a new review within a transaction
func (r *ReviewRepository) Create(ctx context.Context, review *domain.Review) error {
	tx, ok := ctx.Value("tx").(*sql.Tx)
	if !ok {
		// If no transaction, start one
		var err error
		tx, err = r.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()

		_, err = tx.ExecContext(ctx, `
			INSERT INTO reviews (id, word_id, user_id, result, review_type, reviewed_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, review.ID, review.WordID, review.UserID, review.Result, review.ReviewType, time.Now())

		if err != nil {
			return err
		}

		// Update stats
		if err := r.UpdateStats(ctx, review.WordID, review.Result); err != nil {
			return err
		}

		return tx.Commit()
	}

	// Use existing transaction
	_, err := tx.ExecContext(ctx, `
		INSERT INTO reviews (id, word_id, user_id, result, review_type, reviewed_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, review.ID, review.WordID, review.UserID, review.Result, review.ReviewType, time.Now())
	return err
}

// GetStats retrieves review statistics for a word
func (r *ReviewRepository) GetStats(ctx context.Context, wordID string) (*domain.ReviewStats, error) {
	var stats domain.ReviewStats
	var lastReviewedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, `
		SELECT
			word_id,
			total_reviews,
			correct_reviews,
			last_reviewed_at,
			accuracy_rate,
			memory_score
		FROM review_stats
		WHERE word_id = $1
	`, wordID).Scan(
		&stats.WordID,
		&stats.TotalReviews,
		&stats.CorrectReviews,
		&lastReviewedAt,
		&stats.AccuracyRate,
		&stats.MemoryScore,
	)

	if err == sql.ErrNoRows {
		// Return default stats if not found
		return &domain.ReviewStats{
			WordID:         wordID,
			TotalReviews:   0,
			CorrectReviews: 0,
			AccuracyRate:   0.0,
			MemoryScore:    0.0,
		}, nil
	}
	if err != nil {
		return nil, err
	}

	if lastReviewedAt.Valid {
		stats.LastReviewedAt = &lastReviewedAt.Time
	}

	return &stats, nil
}

// UpdateStats updates review statistics for a word
func (r *ReviewRepository) UpdateStats(ctx context.Context, wordID string, result bool) error {
	tx, ok := ctx.Value("tx").(*sql.Tx)
	
	var err error
	if ok {
		_, err = tx.ExecContext(ctx, `
		INSERT INTO review_stats (
			word_id,
			total_reviews,
			correct_reviews,
			last_reviewed_at,
			accuracy_rate
		)
		VALUES (
			$1,
			1,
			CASE WHEN $2 = true THEN 1 ELSE 0 END,
			now(),
			CASE WHEN $2 = true THEN 1.0 ELSE 0.0 END
		)
		ON CONFLICT (word_id)
		DO UPDATE SET
			total_reviews = review_stats.total_reviews + 1,
			correct_reviews =
				review_stats.correct_reviews
				+ CASE WHEN $2 = true THEN 1 ELSE 0 END,
			last_reviewed_at = now(),
			accuracy_rate =
				(review_stats.correct_reviews
				 + CASE WHEN $2 = true THEN 1 ELSE 0 END)::float
				/ (review_stats.total_reviews + 1)
	`, wordID, result)
	} else {
		// If no transaction, use db directly
		_, err = r.db.ExecContext(ctx, `
		INSERT INTO review_stats (
			word_id,
			total_reviews,
			correct_reviews,
			last_reviewed_at,
			accuracy_rate
		)
		VALUES (
			$1,
			1,
			CASE WHEN $2 = true THEN 1 ELSE 0 END,
			now(),
			CASE WHEN $2 = true THEN 1.0 ELSE 0.0 END
		)
		ON CONFLICT (word_id)
		DO UPDATE SET
			total_reviews = review_stats.total_reviews + 1,
			correct_reviews =
				review_stats.correct_reviews
				+ CASE WHEN $2 = true THEN 1 ELSE 0 END,
			last_reviewed_at = now(),
			accuracy_rate =
				(review_stats.correct_reviews
				 + CASE WHEN $2 = true THEN 1 ELSE 0 END)::float
				/ (review_stats.total_reviews + 1)
	`, wordID, result)
	}

	return err
}

// GetLastReviewType retrieves the last review type for a word
func (r *ReviewRepository) GetLastReviewType(ctx context.Context, wordID string) (string, error) {
	var reviewType sql.NullString

	err := r.db.QueryRowContext(ctx, `
		SELECT review_type 
		FROM reviews 
		WHERE word_id = $1 
		ORDER BY reviewed_at DESC 
		LIMIT 1
	`, wordID).Scan(&reviewType)

	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}

	if reviewType.Valid {
		return reviewType.String, nil
	}

	return "", nil
}
