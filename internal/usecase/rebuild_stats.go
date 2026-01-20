package usecase

import (
	"context"
	"database/sql"
)

// RebuildStatsUseCase handles rebuilding review statistics
// This is a background job use case for maintaining data consistency
type RebuildStatsUseCase struct {
	db *sql.DB
}

// NewRebuildStatsUseCase creates a new RebuildStatsUseCase
func NewRebuildStatsUseCase(db *sql.DB) *RebuildStatsUseCase {
	return &RebuildStatsUseCase{db: db}
}

// RebuildStats rebuilds all review statistics from scratch
// This is useful for data migration or fixing inconsistencies
func (uc *RebuildStatsUseCase) RebuildStats(ctx context.Context) error {
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Clear existing stats
	_, err = tx.ExecContext(ctx, `TRUNCATE review_stats`)
	if err != nil {
		return err
	}

	// Rebuild from reviews table
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
