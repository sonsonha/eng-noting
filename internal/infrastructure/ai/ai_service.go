package ai

import (
	"github.com/sonsonha/eng-noting/internal/domain/ai"
)

// AIService implements domain.AIService using the AI client
type AIService struct {
	client ai.Client
}

// NewAIService creates a new AIService
func NewAIService(client ai.Client) *AIService {
	return &AIService{client: client}
}

// ExplainWord generates an AI explanation for a word
func (s *AIService) ExplainWord(word, context string) (*ai.AIExplanation, error) {
	exp, err := ai.ExplainWordSafe(s.client, word, context)
	if err != nil {
		return nil, err
	}

	return &ai.AIExplanation{
		Definition:   exp.Definition,
		ExampleGood:  exp.ExampleGood,
		ExampleBad:   exp.ExampleBad,
		PartOfSpeech: exp.PartOfSpeech,
		CEFRLevel:    exp.CEFRLevel,
	}, nil
}
