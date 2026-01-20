package http

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/sonsonha/eng-noting/internal/domain"
	"github.com/sonsonha/eng-noting/internal/usecase"
)

type SubmitReviewRequest struct {
	WordID     string `json:"word_id"`
	Result     bool   `json:"result"`
	ReviewType string `json:"review_type"`
}

type SubmitReviewResponse struct {
	Success bool `json:"success"`
}

func (h *Handler) SubmitReview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := mustUserIDFromContext(ctx)

	var req SubmitReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	if req.WordID == "" || req.ReviewType == "" {
		writeError(w, http.StatusBadRequest, "missing required fields")
		return
	}

	input := usecase.SubmitReviewInput{
		UserID:     userID,
		WordID:     req.WordID,
		Result:     req.Result,
		ReviewType: req.ReviewType,
	}

	output, err := h.reviewUseCase.SubmitReview(ctx, input)
	if err != nil {
		if err == usecase.ErrNotFound {
			writeError(w, http.StatusNotFound, "word not found")
			return
		}
		if err == usecase.ErrForbidden {
			writeError(w, http.StatusForbidden, "word does not belong to user")
			return
		}
		h.logger.Error("failed to submit review", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to process review")
		return
	}

	writeJSON(w, http.StatusOK, SubmitReviewResponse{Success: output.Success})
}

type StartSessionResponse struct {
	SessionID string        `json:"session_id"`
	Items     []SessionItem `json:"items"`
	Total     int           `json:"total"`
}

type SessionItem struct {
	WordID        string  `json:"word_id"`
	ReviewType    string  `json:"review_type"`
	PriorityScore float64 `json:"priority_score"`
	Reason        string  `json:"reason"`
}

// Session storage for MVP (in-memory)
// TODO: Replace with Redis for production
var (
	sessionStore = make(map[string]*domain.Session)
	sessionMutex sync.RWMutex
)

func (h *Handler) StartSession(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := mustUserIDFromContext(ctx)

	input := usecase.StartSessionInput{
		UserID: userID,
	}

	output, err := h.sessionUseCase.StartSession(ctx, input)
	if err != nil {
		h.logger.Error("failed to start session", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to create session")
		return
	}

	// Convert session items to response items
	items := make([]SessionItem, len(output.Items))
	for i, item := range output.Items {
		items[i] = SessionItem{
			WordID:        item.WordID,
			ReviewType:    item.ReviewType,
			PriorityScore: item.PriorityScore,
			Reason:        item.Reason,
		}
	}

	// Store session in memory
	sessionID := uuid.NewString()
	session := &domain.Session{
		UserID: userID,
		Items:  output.Items,
		Index:  0,
	}

	sessionMutex.Lock()
	sessionStore[sessionID] = session
	sessionMutex.Unlock()

	writeJSON(w, http.StatusOK, StartSessionResponse{
		SessionID: sessionID,
		Items:     items,
		Total:     output.Total,
	})
}

func (h *Handler) GetCurrentItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := mustUserIDFromContext(ctx)

	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		writeError(w, http.StatusBadRequest, "missing session_id")
		return
	}

	sessionMutex.RLock()
	sess, exists := sessionStore[sessionID]
	sessionMutex.RUnlock()

	if !exists {
		writeError(w, http.StatusNotFound, "session not found")
		return
	}

	// Verify session belongs to user
	if sess.UserID != userID {
		writeError(w, http.StatusForbidden, "session does not belong to user")
		return
	}

	item := sess.Current()
	if item == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"done": true,
		})
		return
	}

	writeJSON(w, http.StatusOK, SessionItem{
		WordID:        item.WordID,
		ReviewType:    item.ReviewType,
		PriorityScore: item.PriorityScore,
		Reason:        item.Reason,
	})
}

func (h *Handler) AdvanceSession(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := mustUserIDFromContext(ctx)

	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		writeError(w, http.StatusBadRequest, "missing session_id")
		return
	}

	sessionMutex.Lock()
	sess, exists := sessionStore[sessionID]
	sessionMutex.Unlock()

	if !exists {
		writeError(w, http.StatusNotFound, "session not found")
		return
	}

	// Verify session belongs to user
	if sess.UserID != userID {
		writeError(w, http.StatusForbidden, "session does not belong to user")
		return
	}

	sess.Advance()

	writeJSON(w, http.StatusOK, map[string]bool{
		"success": true,
		"done":    sess.Done(),
	})
}
