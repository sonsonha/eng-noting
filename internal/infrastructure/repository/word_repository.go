package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/sonsonha/eng-noting/internal/domain"
	wordDomain "github.com/sonsonha/eng-noting/internal/domain/word"
)

// WordRepository implements domain.WordRepository using PostgreSQL
type WordRepository struct {
	db *sql.DB
}

// NewWordRepository creates a new WordRepository
func NewWordRepository(db *sql.DB) *WordRepository {
	return &WordRepository{db: db}
}

// Create creates a new word
func (r *WordRepository) Create(ctx context.Context, word *wordDomain.Word) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO words (id, user_id, text, context, confidence, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, word.ID, word.UserID, word.Text, word.Context, word.Confidence, word.CreatedAt, word.UpdatedAt)
	return err
}

// GetByID retrieves a word by ID
func (r *WordRepository) GetByID(ctx context.Context, wordID, userID string) (*wordDomain.Word, error) {
	var word wordDomain.Word
	var createdAt, updatedAt sql.NullTime

	var aiDefinition, aiExampleGood sql.NullString
	var aiExampleBad, aiPOS, aiCEFR sql.NullString

	err := r.db.QueryRowContext(ctx, `
		SELECT
			w.id,
			w.user_id,
			w.text,
			w.context,
			w.source,
			w.confidence,
			w.created_at,
			w.updated_at,
			ai.definition,
			ai.example_good,
			ai.example_bad,
			ai.pos,
			ai.cefr_level
		FROM words w
		LEFT JOIN word_ai_data ai ON ai.word_id = w.id
		WHERE w.id = $1 AND w.user_id = $2
	`, wordID, userID).Scan(
		&word.ID,
		&word.UserID,
		&word.Text,
		&word.Context,
		&word.Source,
		&word.Confidence,
		&createdAt,
		&updatedAt,
		&aiDefinition,
		&aiExampleGood,
		&aiExampleBad,
		&aiPOS,
		&aiCEFR,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrWordNotFound
	}
	if err != nil {
		return nil, err
	}

	if createdAt.Valid {
		word.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		word.UpdatedAt = updatedAt.Time
	}

	// Set AI data if available
	if aiDefinition.Valid {
		word.AIData = &wordDomain.WordAIData{
			WordID:      wordID,
			Definition:  aiDefinition.String,
			ExampleGood: aiExampleGood.String,
			GeneratedAt: time.Now(), // Could fetch from DB if needed
		}
		if aiExampleBad.Valid {
			word.AIData.ExampleBad = &aiExampleBad.String
		}
		if aiPOS.Valid {
			word.AIData.PartOfSpeech = &aiPOS.String
		}
		if aiCEFR.Valid {
			word.AIData.CEFRLevel = &aiCEFR.String
		}
	}

	return &word, nil
}

// List retrieves a list of words for a user
func (r *WordRepository) List(ctx context.Context, userID string, limit, offset int) ([]*wordDomain.Word, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			w.id,
			w.user_id,
			w.text,
			w.context,
			w.source,
			w.confidence,
			w.created_at,
			w.updated_at,
			ai.definition,
			ai.example_good,
			ai.example_bad,
			ai.pos,
			ai.cefr_level
		FROM words w
		LEFT JOIN word_ai_data ai ON ai.word_id = w.id
		WHERE w.user_id = $1
		ORDER BY w.created_at DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var words []*wordDomain.Word
	for rows.Next() {
		var word wordDomain.Word
		var createdAt, updatedAt sql.NullTime
		var aiDefinition, aiExampleGood sql.NullString
		var aiExampleBad, aiPOS, aiCEFR sql.NullString

		err := rows.Scan(
			&word.ID,
			&word.UserID,
			&word.Text,
			&word.Context,
			&word.Source,
			&word.Confidence,
			&createdAt,
			&updatedAt,
			&aiDefinition,
			&aiExampleGood,
			&aiExampleBad,
			&aiPOS,
			&aiCEFR,
		)
		if err != nil {
			continue
		}

		if createdAt.Valid {
			word.CreatedAt = createdAt.Time
		}
		if updatedAt.Valid {
			word.UpdatedAt = updatedAt.Time
		}

		if aiDefinition.Valid {
			aiData := &wordDomain.WordAIData{
				WordID:      word.ID,
				Definition:  aiDefinition.String,
				ExampleGood: aiExampleGood.String,
				GeneratedAt: time.Now(),
			}
			if aiExampleBad.Valid {
				aiData.ExampleBad = &aiExampleBad.String
			}
			if aiPOS.Valid {
				aiData.PartOfSpeech = &aiPOS.String
			}
			if aiCEFR.Valid {
				aiData.CEFRLevel = &aiCEFR.String
			}
			word.AIData = aiData
		}

		words = append(words, &word)
	}

	return words, nil
}

// Count returns the total count of words for a user
func (r *WordRepository) Count(ctx context.Context, userID string) (int, error) {
	var total int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM words WHERE user_id = $1
	`, userID).Scan(&total)
	return total, err
}

// StoreAIData stores AI-generated data for a word
func (r *WordRepository) StoreAIData(ctx context.Context, wordID string, aiData *wordDomain.WordAIData) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO word_ai_data (
			word_id,
			definition,
			example_good,
			example_bad,
			pos,
			cefr_level,
			generated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (word_id) DO NOTHING
	`,
		wordID,
		aiData.Definition,
		aiData.ExampleGood,
		aiData.ExampleBad,
		aiData.PartOfSpeech,
		aiData.CEFRLevel,
		time.Now(),
	)
	return err
}
