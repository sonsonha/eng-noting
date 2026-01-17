package job

import (
	"context"
	"database/sql"

	"github.com/sonsonha/eng-noting/internal/db"
	"github.com/sonsonha/eng-noting/internal/mps"
	"github.com/sonsonha/eng-noting/internal/review"
)

func RebuildReviewQueue(
	ctx context.Context,
	conn *sql.DB,
	userID string,
) error {

	// 1. Load DB stats
	rows, err := db.LoadWordStats(ctx, conn, userID)
	if err != nil {
		return err
	}

	// 2. Transaction (atomic rebuild)
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 3. Clear old queue (idempotent)
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM review_queue WHERE user_id = $1`,
		userID,
	); err != nil {
		return err
	}

	// 4. Recompute everything
	for _, r := range rows {

		stats := mps.FromDBRow(
			r.Confidence,
			r.AccuracyRate,
			r.TotalReviews,
			r.LastReviewedAt,
			r.RecentFailures,
			r.RecentReviews,
			r.FrequencyScore,
		)

		score, mpsReason := mps.CalculateMPS(stats)

		// Skip low priority words
		if score < 30 {
			continue
		}

		ctxReview := review.Context{
			MPS:            score,
			AccuracyRate:   r.AccuracyRate,
			TotalReviews:   r.TotalReviews,
			LastReviewType: "", // safe default for now
		}

		reviewType := review.SelectType(ctxReview)
		reviewReason := review.Reason(ctxReview, reviewType)

		finalReason := mpsReason + ". " + reviewReason

		_, err := tx.ExecContext(ctx, `
			INSERT INTO review_queue (
				user_id,
				word_id,
				priority_score,
				reason
			)
			VALUES ($1, $2, $3, $4)
		`,
			userID,
			r.WordID,
			score,
			finalReason,
		)
		if err != nil {
			return err
		}
	}

	// 5. Commit atomically
	return tx.Commit()
}
