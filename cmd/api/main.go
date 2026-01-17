package main

import (
	"database/sql"
	"fmt"
	"log"
	stdhttp "net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/lib/pq"
	"github.com/sonsonha/eng-noting/internal/ai/openai"
	"github.com/sonsonha/eng-noting/internal/config"
	httphandler "github.com/sonsonha/eng-noting/internal/http"
)

func main() {
	cfg := config.LoadConfig()

	// Database setup
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("Database connection established successfully.")

	// AI client setup
	aiClient := openai.NewClient(cfg.AIAPIKey)
	if aiClient == nil {
		log.Println("Warning: AI client not initialized (AI_API_KEY not set)")
	}

	// HTTP handler setup
	handler := httphandler.NewHandler(db, aiClient)

	// Router setup
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	// API routes with authentication
	r.Route("/api", func(r chi.Router) {
		r.Use(httphandler.AuthMiddleware)

		// Word endpoints
		r.Post("/words", handler.CreateWord)
		r.Get("/words", handler.ListWords)
		r.Get("/words/{id}", handler.GetWord)

		// Review endpoints
		r.Post("/reviews/session", handler.StartSession)
		r.Get("/reviews/session/current", handler.GetCurrentItem)
		r.Post("/reviews/session/advance", handler.AdvanceSession)
		r.Post("/reviews/submit", handler.SubmitReview)
	})

	// Health check endpoint (no auth)
	r.Get("/health", func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		w.WriteHeader(stdhttp.StatusOK)
		w.Write([]byte("OK"))
	})

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Server starting on %s", addr)
	if err := stdhttp.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
