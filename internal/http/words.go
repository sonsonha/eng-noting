package http

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/sonsonha/eng-noting/internal/ai"
)

type CreateWordRequest struct {
	Text    string `json:"text"`
	Context string `json:"context"`
}

type CreateWordResponse struct {
	WordID string `json:"word_id"`
}

func (h *Handler) CreateWord(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req CreateWordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	userID := mustUserIDFromContext(ctx) // already authenticated

	wordID := uuid.NewString()

	// 1️⃣ Insert word FIRST (never block on AI)
	_, err := h.db.ExecContext(ctx, `
		INSERT INTO words (id, user_id, text, context, confidence)
		VALUES ($1, $2, $3, $4, 3)
	`, wordID, userID, req.Text, req.Context)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save word")
		return
	}

	// 2️⃣ Trigger AI explanation asynchronously (safe)
	go h.generateAIExplanation(wordID, req.Text, req.Context)

	// 3️⃣ Respond immediately
	resp := CreateWordResponse{WordID: wordID}
	writeJSON(w, http.StatusCreated, resp)
}

func (h *Handler) generateAIExplanation(
	wordID string,
	word string,
	context string,
) {
	exp, err := ai.ExplainWordSafe(h.aiClient, word, context)
	if err != nil {
		h.logger.Warn("AI explain failed", "word", word, "err", err)
		return
	}

	_, err = h.db.Exec(`
		INSERT INTO word_ai_data (
			word_id,
			definition,
			example_good,
			example_bad,
			pos,
			cefr_level
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (word_id) DO NOTHING
	`,
		wordID,
		exp.Definition,
		exp.ExampleGood,
		exp.ExampleBad,
		exp.PartOfSpeech,
		exp.CEFRLevel,
	)

	if err != nil {
		h.logger.Warn("failed to store AI data", "word_id", wordID, "err", err)
	}
}

type WordResponse struct {
	ID          string  `json:"id"`
	Text        string  `json:"text"`
	Context     *string `json:"context,omitempty"`
	Source      *string `json:"source,omitempty"`
	Confidence  *int    `json:"confidence,omitempty"`
	CreatedAt   string  `json:"created_at"`
	Definition  *string `json:"definition,omitempty"`
	ExampleGood *string `json:"example_good,omitempty"`
	ExampleBad  *string `json:"example_bad,omitempty"`
	PartOfSpeech *string `json:"part_of_speech,omitempty"`
	CEFRLevel   *string `json:"cefr_level,omitempty"`
}

func (h *Handler) GetWord(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := mustUserIDFromContext(ctx)

	wordID := r.PathValue("id")
	if wordID == "" {
		writeError(w, http.StatusBadRequest, "missing word ID")
		return
	}

	var word WordResponse
	var createdAt sql.NullTime

	err := h.db.QueryRowContext(ctx, `
		SELECT
			w.id,
			w.text,
			w.context,
			w.source,
			w.confidence,
			w.created_at,
			ai.definition,
			ai.example_good,
			ai.example_bad,
			ai.pos,
			ai.cefr_level
		FROM words w
		LEFT JOIN word_ai_data ai ON ai.word_id = w.id
		WHERE w.id = $1 AND w.user_id = $2
	`, wordID, userID).Scan(
		&word.ID,
		&word.Text,
		&word.Context,
		&word.Source,
		&word.Confidence,
		&createdAt,
		&word.Definition,
		&word.ExampleGood,
		&word.ExampleBad,
		&word.PartOfSpeech,
		&word.CEFRLevel,
	)

	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "word not found")
		return
	}
	if err != nil {
		h.logger.Error("failed to get word", "word_id", wordID, "err", err)
		writeError(w, http.StatusInternalServerError, "failed to get word")
		return
	}

	if createdAt.Valid {
		word.CreatedAt = createdAt.Time.Format("2006-01-02T15:04:05Z07:00")
	}

	writeJSON(w, http.StatusOK, word)
}

type ListWordsResponse struct {
	Words []WordResponse `json:"words"`
	Total int            `json:"total"`
}

func (h *Handler) ListWords(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := mustUserIDFromContext(ctx)

	// Simple pagination (can be enhanced later)
	limit := 50
	offset := 0

	rows, err := h.db.QueryContext(ctx, `
		SELECT
			w.id,
			w.text,
			w.context,
			w.source,
			w.confidence,
			w.created_at,
			ai.definition,
			ai.example_good,
			ai.example_bad,
			ai.pos,
			ai.cefr_level
		FROM words w
		LEFT JOIN word_ai_data ai ON ai.word_id = w.id
		WHERE w.user_id = $1
		ORDER BY w.created_at DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)
	if err != nil {
		h.logger.Error("failed to list words", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to list words")
		return
	}
	defer rows.Close()

	var words []WordResponse
	for rows.Next() {
		var word WordResponse
		var createdAt sql.NullTime

		if err := rows.Scan(
			&word.ID,
			&word.Text,
			&word.Context,
			&word.Source,
			&word.Confidence,
			&createdAt,
			&word.Definition,
			&word.ExampleGood,
			&word.ExampleBad,
			&word.PartOfSpeech,
			&word.CEFRLevel,
		); err != nil {
			h.logger.Error("failed to scan word", "err", err)
			continue
		}

		if createdAt.Valid {
			word.CreatedAt = createdAt.Time.Format("2006-01-02T15:04:05Z07:00")
		}

		words = append(words, word)
	}

	// Get total count
	var total int
	err = h.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM words WHERE user_id = $1
	`, userID).Scan(&total)
	if err != nil {
		total = len(words)
	}

	writeJSON(w, http.StatusOK, ListWordsResponse{
		Words: words,
		Total: total,
	})
}
