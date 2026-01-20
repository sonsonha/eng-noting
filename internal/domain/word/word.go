package word

import (
	"context"
	"time"
)

// Word represents a word entity in the domain
type Word struct {
	ID         string
	UserID     string
	Text       string
	Context    *string
	Source     *string
	Confidence *int
	CreatedAt  time.Time
	UpdatedAt  time.Time
	AIData     *WordAIData
}

// WordAIData represents AI-generated data for a word
type WordAIData struct {
	WordID       string
	Definition   string
	ExampleGood  string
	ExampleBad   *string
	PartOfSpeech *string
	CEFRLevel    *string
	GeneratedAt  time.Time
}

// WordRepository defines the interface for word persistence
type WordRepository interface {
	Create(ctx context.Context, word *Word) error
	GetByID(ctx context.Context, wordID, userID string) (*Word, error)
	List(ctx context.Context, userID string, limit, offset int) ([]*Word, error)
	Count(ctx context.Context, userID string) (int, error)
	StoreAIData(ctx context.Context, wordID string, aiData *WordAIData) error
}
