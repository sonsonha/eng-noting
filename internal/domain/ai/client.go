package ai

type Client interface {
	ExplainWord(systemPrompt, userPrompt string) (string, error)
}
