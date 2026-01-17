package ai

import "strings"

func ValidateExplanation(word string, e Explanation) bool {
	if e.Definition == "" || e.ExampleGood == "" {
		return false
	}

	// Prevent circular definitions
	if strings.Contains(
		strings.ToLower(e.Definition),
		strings.ToLower(word),
	) {
		return false
	}

	switch e.CEFRLevel {
	case "A2", "B1", "B2":
	default:
		return false
	}

	return true
}
