package job

import (
	"context"
	"database/sql"
)

func RebuildReviewStats(ctx context.Context, db *sql.DB) error {

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `TRUNCATE review_stats`)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO review_stats (
			word_id,
			total_reviews,
			correct_reviews,
			last_reviewed_at,
			accuracy_rate
		)
		SELECT
			r.word_id,
			COUNT(*) AS total_reviews,
			COUNT(*) FILTER (WHERE r.result = true),
			MAX(r.reviewed_at),
			COUNT(*) FILTER (WHERE r.result = true)::float / COUNT(*)
		FROM reviews r
		GROUP BY r.word_id
	`)
	if err != nil {
		return err
	}

	return tx.Commit()
}
