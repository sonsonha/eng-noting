package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/sonsonha/eng-noting/internal/domain/ai"
	wordDomain "github.com/sonsonha/eng-noting/internal/domain/word"
)

// WordUseCase handles word-related business logic
type WordUseCase struct {
	wordRepo wordDomain.WordRepository
	aiSvc    ai.AIService
}

// NewWordUseCase creates a new WordUseCase
func NewWordUseCase(wordRepo wordDomain.WordRepository, aiSvc ai.AIService) *WordUseCase {
	return &WordUseCase{
		wordRepo: wordRepo,
		aiSvc:    aiSvc,
	}
}

// CreateWordInput represents input for creating a word
type CreateWordInput struct {
	UserID  string
	Text    string
	Context string
}

// CreateWordOutput represents output from creating a word
type CreateWordOutput struct {
	WordID string
}

// CreateWord creates a new word and triggers AI explanation asynchronously
func (uc *WordUseCase) CreateWord(ctx context.Context, input CreateWordInput) (*CreateWordOutput, error) {
	wordID := uuid.NewString()
	now := time.Now()
	confidence := 3

	word := &wordDomain.Word{
		ID:         wordID,
		UserID:     input.UserID,
		Text:       input.Text,
		Context:    &input.Context,
		Confidence: &confidence,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := uc.wordRepo.Create(ctx, word); err != nil {
		return nil, err
	}

	// Trigger AI explanation asynchronously (non-blocking)
	go uc.generateAIExplanation(wordID, input.Text, input.Context)

	return &CreateWordOutput{WordID: wordID}, nil
}

// generateAIExplanation generates and stores AI explanation for a word
func (uc *WordUseCase) generateAIExplanation(wordID, word, wordContext string) {
	exp, err := uc.aiSvc.ExplainWord(word, wordContext)
	if err != nil {
		// Log error but don't fail - word is already created
		return
	}

	aiData := &wordDomain.WordAIData{
		WordID:       wordID,
		Definition:   exp.Definition,
		ExampleGood:  exp.ExampleGood,
		ExampleBad:   &exp.ExampleBad,
		PartOfSpeech: &exp.PartOfSpeech,
		CEFRLevel:    &exp.CEFRLevel,
		GeneratedAt:  time.Now(),
	}

	// Use background context since this is async
	ctx := context.Background()
	_ = uc.wordRepo.StoreAIData(ctx, wordID, aiData)
}

// GetWordInput represents input for getting a word
type GetWordInput struct {
	WordID string
	UserID string
}

// GetWordOutput represents output from getting a word
type GetWordOutput struct {
	Word *wordDomain.Word
}

// GetWord retrieves a word by ID
func (uc *WordUseCase) GetWord(ctx context.Context, input GetWordInput) (*GetWordOutput, error) {
	word, err := uc.wordRepo.GetByID(ctx, input.WordID, input.UserID)
	if err != nil {
		return nil, err
	}

	return &GetWordOutput{Word: word}, nil
}

// ListWordsInput represents input for listing words
type ListWordsInput struct {
	UserID string
	Limit  int
	Offset int
}

// ListWordsOutput represents output from listing words
type ListWordsOutput struct {
	Words []*wordDomain.Word
	Total int
}

// ListWords retrieves a list of words for a user
func (uc *WordUseCase) ListWords(ctx context.Context, input ListWordsInput) (*ListWordsOutput, error) {
	words, err := uc.wordRepo.List(ctx, input.UserID, input.Limit, input.Offset)
	if err != nil {
		return nil, err
	}

	total, err := uc.wordRepo.Count(ctx, input.UserID)
	if err != nil {
		// Fallback to length if count fails
		total = len(words)
	}

	return &ListWordsOutput{
		Words: words,
		Total: total,
	}, nil
}
