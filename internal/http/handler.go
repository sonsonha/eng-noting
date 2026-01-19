package http

import (
	"log"

	"github.com/sonsonha/eng-noting/internal/usecase"
)

// Handler holds all HTTP handlers and their dependencies
type Handler struct {
	wordUseCase    *usecase.WordUseCase
	reviewUseCase  *usecase.ReviewUseCase
	sessionUseCase *usecase.SessionUseCase
	logger         Logger
}

// Logger interface for logging
type Logger interface {
	Warn(msg string, fields ...any)
	Error(msg string, fields ...any)
	Info(msg string, fields ...any)
}

type stdLogger struct{}

func (l *stdLogger) Warn(msg string, fields ...any) {
	log.Printf("[WARN] %s %v", msg, fields)
}

func (l *stdLogger) Error(msg string, fields ...any) {
	log.Printf("[ERROR] %s %v", msg, fields)
}

func (l *stdLogger) Info(msg string, fields ...any) {
	log.Printf("[INFO] %s %v", msg, fields)
}

// NewHandler creates a new HTTP handler with use cases
func NewHandler(
	wordUseCase *usecase.WordUseCase,
	reviewUseCase *usecase.ReviewUseCase,
	sessionUseCase *usecase.SessionUseCase,
) *Handler {
	return &Handler{
		wordUseCase:    wordUseCase,
		reviewUseCase:  reviewUseCase,
		sessionUseCase: sessionUseCase,
		logger:         &stdLogger{},
	}
}
