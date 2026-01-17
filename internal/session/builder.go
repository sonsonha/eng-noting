package session

import (
	"context"
	"database/sql"

	"github.com/sonsonha/eng-noting/internal/review"
)

const (
	MaxCritical = 5
	MaxNormal   = 5
)

func BuildSession(
	ctx context.Context,
	db *sql.DB,
	userID string,
) (*Session, error) {

	rows, err := db.QueryContext(ctx, `
		SELECT
			rq.word_id,
			rq.priority_score,
			rq.reason,
			COALESCE(rs.accuracy_rate, 0) as accuracy_rate,
			COALESCE(rs.total_reviews, 0) as total_reviews,
			(
				SELECT review_type 
				FROM reviews 
				WHERE word_id = rq.word_id 
				ORDER BY reviewed_at DESC 
				LIMIT 1
			) as last_review_type
		FROM review_queue rq
		LEFT JOIN review_stats rs ON rs.word_id = rq.word_id
		WHERE rq.user_id = $1
		ORDER BY rq.priority_score DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var critical []Item
	var normal []Item

	for rows.Next() {
		var it Item
		var accuracyRate float64
		var totalReviews int
		var lastReviewType sql.NullString

		if err := rows.Scan(
			&it.WordID,
			&it.PriorityScore,
			&it.Reason,
			&accuracyRate,
			&totalReviews,
			&lastReviewType,
		); err != nil {
			return nil, err
		}

		// Calculate review type
		reviewCtx := review.Context{
			MPS:            it.PriorityScore,
			AccuracyRate:   accuracyRate,
			TotalReviews:   totalReviews,
			LastReviewType: "",
		}
		if lastReviewType.Valid {
			reviewCtx.LastReviewType = lastReviewType.String
		}
		it.ReviewType = review.SelectType(reviewCtx)

		switch {
		case it.PriorityScore >= 60 && len(critical) < MaxCritical:
			critical = append(critical, it)
		case it.PriorityScore >= 40 && len(normal) < MaxNormal:
			normal = append(normal, it)
		}

		if len(critical) == MaxCritical && len(normal) == MaxNormal {
			break
		}
	}

	items := append(critical, normal...)

	return &Session{
		UserID: userID,
		Items:  items,
		Index:  0,
	}, nil
}
