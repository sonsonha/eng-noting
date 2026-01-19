package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/sonsonha/eng-noting/internal/usecase"
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
	userID := mustUserIDFromContext(ctx)

	var req CreateWordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	input := usecase.CreateWordInput{
		UserID:  userID,
		Text:    req.Text,
		Context: req.Context,
	}

	output, err := h.wordUseCase.CreateWord(ctx, input)
	if err != nil {
		h.logger.Error("failed to create word", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to save word")
		return
	}

	resp := CreateWordResponse{WordID: output.WordID}
	writeJSON(w, http.StatusCreated, resp)
}

type WordResponse struct {
	ID           string  `json:"id"`
	Text         string  `json:"text"`
	Context      *string `json:"context,omitempty"`
	Source       *string `json:"source,omitempty"`
	Confidence   *int    `json:"confidence,omitempty"`
	CreatedAt    string  `json:"created_at"`
	Definition   *string `json:"definition,omitempty"`
	ExampleGood  *string `json:"example_good,omitempty"`
	ExampleBad   *string `json:"example_bad,omitempty"`
	PartOfSpeech *string `json:"part_of_speech,omitempty"`
	CEFRLevel    *string `json:"cefr_level,omitempty"`
}

func (h *Handler) GetWord(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := mustUserIDFromContext(ctx)

	wordID := r.PathValue("id")
	if wordID == "" {
		writeError(w, http.StatusBadRequest, "missing word ID")
		return
	}

	input := usecase.GetWordInput{
		WordID: wordID,
		UserID: userID,
	}

	output, err := h.wordUseCase.GetWord(ctx, input)
	if err != nil {
		if err == usecase.ErrNotFound {
			writeError(w, http.StatusNotFound, "word not found")
			return
		}
		h.logger.Error("failed to get word", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to get word")
		return
	}

	word := output.Word
	resp := WordResponse{
		ID:         word.ID,
		Text:       word.Text,
		Context:    word.Context,
		Source:     word.Source,
		Confidence: word.Confidence,
		CreatedAt:  word.CreatedAt.Format(time.RFC3339),
	}

	if word.AIData != nil {
		resp.Definition = &word.AIData.Definition
		resp.ExampleGood = &word.AIData.ExampleGood
		resp.ExampleBad = word.AIData.ExampleBad
		resp.PartOfSpeech = word.AIData.PartOfSpeech
		resp.CEFRLevel = word.AIData.CEFRLevel
	}

	writeJSON(w, http.StatusOK, resp)
}

type ListWordsResponse struct {
	Words []WordResponse `json:"words"`
	Total int            `json:"total"`
}

func (h *Handler) ListWords(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := mustUserIDFromContext(ctx)

	// Parse pagination parameters
	limit := 50
	offset := 0
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	input := usecase.ListWordsInput{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	}

	output, err := h.wordUseCase.ListWords(ctx, input)
	if err != nil {
		h.logger.Error("failed to list words", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to list words")
		return
	}

	words := make([]WordResponse, len(output.Words))
	for i, word := range output.Words {
		words[i] = WordResponse{
			ID:         word.ID,
			Text:       word.Text,
			Context:    word.Context,
			Source:     word.Source,
			Confidence: word.Confidence,
			CreatedAt:  word.CreatedAt.Format(time.RFC3339),
		}

		if word.AIData != nil {
			words[i].Definition = &word.AIData.Definition
			words[i].ExampleGood = &word.AIData.ExampleGood
			words[i].ExampleBad = word.AIData.ExampleBad
			words[i].PartOfSpeech = word.AIData.PartOfSpeech
			words[i].CEFRLevel = word.AIData.CEFRLevel
		}
	}

	writeJSON(w, http.StatusOK, ListWordsResponse{
		Words: words,
		Total: output.Total,
	})
}
