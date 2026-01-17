package http

import (
	"database/sql"
	"log"

	"github.com/sonsonha/eng-noting/internal/ai"
)

type Handler struct {
	db       *sql.DB
	aiClient ai.Client
	logger   Logger
}

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

func NewHandler(db *sql.DB, aiClient ai.Client) *Handler {
	return &Handler{
		db:       db,
		aiClient: aiClient,
		logger:   &stdLogger{},
	}
}
