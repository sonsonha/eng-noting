package db

import (
	"context"
	"database/sql"
)

func UpdateReviewStats(
	ctx context.Context,
	tx *sql.Tx,
	wordID string,
	result bool,
) error {

	_, err := tx.ExecContext(ctx, `
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
				/ (review_stats.total_reviews + 1);
	`, wordID, result)

	return err
}
