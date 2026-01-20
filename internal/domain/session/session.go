package session

import "context"

// SessionItem represents an item in a review session
type SessionItem struct {
	WordID        string
	ReviewType    string
	PriorityScore float64
	Reason        string
}

// Session represents a review session
type Session struct {
	UserID string
	Items  []SessionItem
	Index  int
}

// Current returns the current item in the session
func (s *Session) Current() *SessionItem {
	if s.Index >= len(s.Items) {
		return nil
	}
	return &s.Items[s.Index]
}

// Advance moves to the next item in the session
func (s *Session) Advance() {
	if s.Index < len(s.Items) {
		s.Index++
	}
}

// Done returns true if the session is complete
func (s *Session) Done() bool {
	return s.Index >= len(s.Items)
}

// ReviewQueueItem represents an item in the review queue
type ReviewQueueItem struct {
	UserID        string
	WordID        string
	PriorityScore float64
	Reason        string
}

// ReviewQueueRepository defines the interface for review queue persistence
type ReviewQueueRepository interface {
	Rebuild(ctx context.Context, userID string, items []ReviewQueueItem) error
	GetQueueItems(ctx context.Context, userID string) ([]ReviewQueueItem, error)
}
