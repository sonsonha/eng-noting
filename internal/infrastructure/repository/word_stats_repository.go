package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/sonsonha/eng-noting/internal/domain"
)

// WordStatsRepository implements domain.WordStatsRepository using PostgreSQL
type WordStatsRepository struct {
	db *sql.DB
}

// NewWordStatsRepository creates a new WordStatsRepository
func NewWordStatsRepository(db *sql.DB) *WordStatsRepository {
	return &WordStatsRepository{db: db}
}

// LoadStats loads word statistics for a user
func (r *WordStatsRepository) LoadStats(ctx context.Context, userID string) ([]domain.WordStats, error) {
	const q = `
WITH recent_reviews AS (
    SELECT
        word_id,
        COUNT(*) AS recent_reviews,
        COUNT(*) FILTER (WHERE result = false) AS recent_failures
    FROM reviews
    WHERE user_id = $1
      AND reviewed_at >= now() - interval '7 days'
    GROUP BY word_id
)
SELECT
    w.id AS word_id,
    w.confidence,
    COALESCE(rs.accuracy_rate, 0) AS accuracy_rate,
    COALESCE(rs.total_reviews, 0) AS total_reviews,
    rs.last_reviewed_at,
    COALESCE(rr.recent_failures, 0) AS recent_failures,
    COALESCE(rr.recent_reviews, 0) AS recent_reviews
FROM words w
LEFT JOIN review_stats rs ON rs.word_id = w.id
LEFT JOIN recent_reviews rr ON rr.word_id = w.id
WHERE w.user_id = $1;
`

	rows, err := r.db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.WordStats

	for rows.Next() {
		var r domain.WordStats
		var lastReviewedAt sql.NullTime

		if err := rows.Scan(
			&r.WordID,
			&r.Confidence,
			&r.AccuracyRate,
			&r.TotalReviews,
			&lastReviewedAt,
			&r.RecentFailures,
			&r.RecentReviews,
		); err != nil {
			return nil, err
		}

		// Convert time to string for domain model
		if lastReviewedAt.Valid {
			timeStr := lastReviewedAt.Time.Format(time.RFC3339)
			r.LastReviewedAt = &timeStr
		}

		// Temporary constant until frequency source exists
		r.FrequencyScore = 0.5

		result = append(result, r)
	}

	return result, rows.Err()
}
