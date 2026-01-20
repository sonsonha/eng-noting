package ai

// AIExplanation represents AI-generated explanation for a word
type AIExplanation struct {
	Definition   string
	ExampleGood  string
	ExampleBad   string
	PartOfSpeech string
	CEFRLevel    string
}

// AIService defines the interface for AI operations
type AIService interface {
	ExplainWord(word, context string) (*AIExplanation, error)
}
