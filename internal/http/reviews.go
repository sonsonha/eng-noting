package http

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/sonsonha/eng-noting/internal/db"
	"github.com/sonsonha/eng-noting/internal/job"
	"github.com/sonsonha/eng-noting/internal/session"
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

	// Start transaction
	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		h.logger.Error("failed to begin transaction", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to process review")
		return
	}
	defer tx.Rollback()

	// Verify word belongs to user
	var wordUserID string
	err = tx.QueryRowContext(ctx, `SELECT user_id FROM words WHERE id = $1`, req.WordID).Scan(&wordUserID)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "word not found")
		return
	}
	if err != nil {
		h.logger.Error("failed to verify word ownership", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to process review")
		return
	}
	if wordUserID != userID {
		writeError(w, http.StatusForbidden, "word does not belong to user")
		return
	}

	// Insert review record
	reviewID := uuid.NewString()
	_, err = tx.ExecContext(ctx, `
		INSERT INTO reviews (id, word_id, user_id, result, review_type)
		VALUES ($1, $2, $3, $4, $5)
	`, reviewID, req.WordID, userID, req.Result, req.ReviewType)
	if err != nil {
		h.logger.Error("failed to insert review", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to save review")
		return
	}

	// Update review stats
	if err := db.UpdateReviewStats(ctx, tx, req.WordID, req.Result); err != nil {
		h.logger.Error("failed to update review stats", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to update stats")
		return
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		h.logger.Error("failed to commit transaction", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to save review")
		return
	}

	writeJSON(w, http.StatusOK, SubmitReviewResponse{Success: true})
}

type StartSessionResponse struct {
	SessionID string         `json:"session_id"`
	Items     []SessionItem  `json:"items"`
	Total     int            `json:"total"`
}

type SessionItem struct {
	WordID        string  `json:"word_id"`
	ReviewType    string  `json:"review_type"`
	PriorityScore float64 `json:"priority_score"`
	Reason        string  `json:"reason"`
}

// Session storage for MVP (in-memory)
// TODO: Replace with Redis for production
var sessionStore = make(map[string]*session.Session)

func (h *Handler) StartSession(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := mustUserIDFromContext(ctx)

	// Rebuild review queue
	if err := job.RebuildReviewQueue(ctx, h.db, userID); err != nil {
		h.logger.Error("failed to rebuild review queue", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to prepare review session")
		return
	}

	// Build session
	sess, err := session.BuildSession(ctx, h.db, userID)
	if err != nil {
		h.logger.Error("failed to build session", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to create session")
		return
	}

	// Convert session items to response items
	items := make([]SessionItem, len(sess.Items))
	for i, item := range sess.Items {
		items[i] = SessionItem{
			WordID:        item.WordID,
			ReviewType:    item.ReviewType,
			PriorityScore: item.PriorityScore,
			Reason:        item.Reason,
		}
	}

	// Store session
	sessionID := uuid.NewString()
	sessionStore[sessionID] = sess

	writeJSON(w, http.StatusOK, StartSessionResponse{
		SessionID: sessionID,
		Items:     items,
		Total:     len(items),
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

	sess, exists := sessionStore[sessionID]
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

	sess, exists := sessionStore[sessionID]
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
