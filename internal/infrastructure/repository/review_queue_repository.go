package repository

import (
	"context"
	"database/sql"

	"github.com/sonsonha/eng-noting/internal/domain"
)

// ReviewQueueRepository implements domain.ReviewQueueRepository using PostgreSQL
type ReviewQueueRepository struct {
	db *sql.DB
}

// NewReviewQueueRepository creates a new ReviewQueueRepository
func NewReviewQueueRepository(db *sql.DB) *ReviewQueueRepository {
	return &ReviewQueueRepository{db: db}
}

// Rebuild rebuilds the review queue for a user
func (r *ReviewQueueRepository) Rebuild(ctx context.Context, userID string, items []domain.ReviewQueueItem) error {
	// Start transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete existing queue for user
	_, err = tx.ExecContext(ctx, `DELETE FROM review_queue WHERE user_id = $1`, userID)
	if err != nil {
		return err
	}

	// Insert new queue items
	for _, item := range items {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO review_queue (user_id, word_id, priority_score, reason)
			VALUES ($1, $2, $3, $4)
		`, item.UserID, item.WordID, item.PriorityScore, item.Reason)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetQueueItems retrieves queue items for a user
func (r *ReviewQueueRepository) GetQueueItems(ctx context.Context, userID string) ([]domain.ReviewQueueItem, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT user_id, word_id, priority_score, reason
		FROM review_queue
		WHERE user_id = $1
		ORDER BY priority_score DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.ReviewQueueItem
	for rows.Next() {
		var item domain.ReviewQueueItem
		if err := rows.Scan(
			&item.UserID,
			&item.WordID,
			&item.PriorityScore,
			&item.Reason,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}
