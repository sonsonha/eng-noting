package ai

import (
	"context"
	"database/sql"
)

// Service is a placeholder for future AI service functionality
// Currently, AI operations are handled directly in HTTP handlers
type Service struct {
	client Client
	db     *sql.DB
	logger interface{} // Logger interface - placeholder
}

func (s *Service) ExplainAndStore(
	ctx context.Context,
	wordID string,
	word string,
	context string,
) {
	// calls ExplainWordSafe
	// validates
	// inserts into word_ai_data
}
