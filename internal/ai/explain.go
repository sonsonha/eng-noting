package ai

import (
	"encoding/json"
	"fmt"
	"time"
)

type Explanation struct {
	Definition   string `json:"definition"`
	ExampleGood  string `json:"example_good"`
	ExampleBad   string `json:"example_bad"`
	PartOfSpeech string `json:"part_of_speech"`
	CEFRLevel    string `json:"cefr_level"`
}

// ExplainWordSafe calls the AI client, parses and validates the response,
// and retries once on failure
func ExplainWordSafe(client Client, word, context string) (*Explanation, error) {
	if client == nil {
		return nil, fmt.Errorf("AI client is nil")
	}

	maxRetries := 2
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// Brief delay before retry
			time.Sleep(time.Second * time.Duration(attempt))
		}

		response, err := client.ExplainWord(systemPrompt, explanationPrompt(word, context))
		if err != nil {
			lastErr = fmt.Errorf("AI call failed: %w", err)
			continue
		}

		var exp Explanation
		if err := json.Unmarshal([]byte(response), &exp); err != nil {
			lastErr = fmt.Errorf("failed to parse JSON response: %w", err)
			continue
		}

		if !ValidateExplanation(word, exp) {
			lastErr = fmt.Errorf("validation failed for explanation")
			continue
		}

		return &exp, nil
	}

	return nil, fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}
