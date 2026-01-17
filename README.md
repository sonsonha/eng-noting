# English Noting

> **Never forget a word you've learned — and only review what actually matters.**

An AI-powered English learning web application centered on vocabulary capture and high-efficiency review. This backend API provides intelligent word management, AI-generated explanations, and a smart selective review system that prioritizes what you actually need to practice.

## Features

### Core Functionality

- **Zero-friction Word Capture**: Add words instantly from reading, videos, or conversations
- **AI-Generated Memory Cards**: Automatic definitions, examples, and CEFR-level classification
- **Smart Selective Review**: Memory Priority Score (MPS) system that determines what to review based on:
  - Time since last review
  - Past accuracy rates
  - User confidence levels
  - Recent failure patterns
  - Word frequency
- **Adaptive Review Formats**: Automatically selects review type (multiple choice, matching, typing, fill-in-blank) based on word mastery level
- **Session Management**: Structured review sessions with prioritized word queues

## Architecture

### Memory Priority Score (MPS)

The MPS is a deterministic scoring algorithm that calculates review priority (0-100) using:

```
MPS = (time_factor × 30) + (accuracy_factor × 30) + 
      (confidence_factor × 15) + (failure_factor × 15) + 
      (frequency_factor × 10)
```

This ensures:
- **Explainable**: Users can understand why each word is prioritized
- **Deterministic**: Same inputs always produce same output
- **Forgiving**: Avoids review burnout by capping daily load

### Review Format Selection

Automatically chooses review format based on mastery:

- **Multiple Choice** (`mcq`): New words or low accuracy (<40%)
- **Matching** (`match`): Medium accuracy (40-70%)
- **Typing** (`typing`): High accuracy (>70%) with low urgency
- **Fill Blank** (`fill_blank`): Mastered words (>80% accuracy, ≥5 reviews)

## Tech Stack

- **Language**: Go 1.25.5+
- **Database**: PostgreSQL
- **HTTP Router**: [chi](https://github.com/go-chi/chi)
- **AI**: OpenAI GPT-4o-mini
- **ORM**: Standard `database/sql` with PostgreSQL driver

## Prerequisites

- Go 1.25.5 or higher
- PostgreSQL 12+
- OpenAI API key (for AI word explanations)

## Setup

### 1. Clone the Repository

```bash
git clone <repository-url>
cd engnoting
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Database Setup

Make sure PostgreSQL is running, then run migrations:

```bash
# Using psql directly
psql -U your_user -d your_database -f migrations/000001_init_schema.up.sql

# Or using a migration tool like migrate
migrate -path migrations -database "postgres://user:pass@localhost/dbname?sslmode=disable" up
```

### 4. Environment Variables

Create a `.env` file or set environment variables:

```bash
DATABASE_URL=postgres://user:password@localhost:5432/engnoting?sslmode=disable
AI_API_KEY=sk-your-openai-api-key
PORT=8080  # Optional, defaults to 8080
```

### 5. Run the Server

```bash
go run cmd/api/main.go
```

The server will start on `http://localhost:8080` (or your configured PORT).

### Docker Compose (Optional)

```bash
docker-compose up -d
```

This will start PostgreSQL. Update `DATABASE_URL` to connect to the containerized database.

## API Documentation

All API endpoints require authentication via Bearer token in the Authorization header:

```
Authorization: Bearer <user_id>
```

> **Note**: The current implementation uses a simple Bearer token with user ID. Replace with proper JWT authentication for production.

### Word Management

#### Create Word

```http
POST /api/words
Content-Type: application/json

{
  "text": "resilient",
  "context": "She stayed resilient after the failure."
}
```

**Response:**
```json
{
  "word_id": "uuid-here"
}
```

The AI explanation is generated asynchronously and stored automatically.

#### List Words

```http
GET /api/words
```

**Response:**
```json
{
  "words": [
    {
      "id": "uuid",
      "text": "resilient",
      "context": "She stayed resilient...",
      "confidence": 3,
      "created_at": "2024-01-15T10:30:00Z",
      "definition": "able to recover quickly",
      "example_good": "She stayed calm after the failure.",
      "example_bad": "She resilient the problem easily.",
      "part_of_speech": "adjective",
      "cefr_level": "B1"
    }
  ],
  "total": 42
}
```

#### Get Word

```http
GET /api/words/{id}
```

Returns detailed information about a single word including AI-generated data.

### Review System

#### Start Review Session

```http
POST /api/reviews/session
```

Rebuilds the review queue and creates a new session. Returns up to 10 words (5 critical + 5 normal priority).

**Response:**
```json
{
  "session_id": "uuid",
  "items": [
    {
      "word_id": "uuid",
      "review_type": "mcq",
      "priority_score": 75.5,
      "reason": "You haven't reviewed this word recently. This word is new — choose the correct meaning"
    }
  ],
  "total": 10
}
```

#### Get Current Item

```http
GET /api/reviews/session/current?session_id={session_id}
```

Returns the current word in the review session.

**Response:**
```json
{
  "word_id": "uuid",
  "review_type": "mcq",
  "priority_score": 75.5,
  "reason": "You haven't reviewed this word recently"
}
```

Or if session is complete:
```json
{
  "done": true
}
```

#### Submit Review

```http
POST /api/reviews/submit
Content-Type: application/json

{
  "word_id": "uuid",
  "result": true,
  "review_type": "mcq"
}
```

Records the review result and updates statistics.

**Response:**
```json
{
  "success": true
}
```

#### Advance Session

```http
POST /api/reviews/session/advance?session_id={session_id}
```

Moves to the next item in the session.

**Response:**
```json
{
  "success": true,
  "done": false
}
```

### Health Check

```http
GET /health
```

Returns `200 OK` if the server is running.

## Project Structure

```
engnoting/
├── cmd/
│   └── api/
│       └── main.go              # Application entry point
├── internal/
│   ├── ai/                      # AI integration
│   │   ├── client.go            # AI client interface
│   │   ├── explain.go           # Word explanation logic
│   │   ├── openai/              # OpenAI implementation
│   │   ├── prompts.go           # AI prompt templates
│   │   └── validate.go          # Response validation
│   ├── config/                  # Configuration management
│   ├── db/                      # Database operations
│   │   ├── load_word_stats.go   # Load review statistics
│   │   └── update_review_stats.go # Update statistics
│   ├── http/                    # HTTP handlers
│   │   ├── handler.go           # Handler struct
│   │   ├── middleware.go        # Auth middleware
│   │   ├── reviews.go           # Review endpoints
│   │   └── words.go             # Word endpoints
│   ├── job/                     # Background jobs
│   │   └── rebuild_review_queue.go # Queue rebuild logic
│   ├── mps/                     # Memory Priority Score
│   │   ├── score.go             # MPS calculation
│   │   ├── model.go             # Data structures
│   │   └── reason.go            # Reason generation
│   ├── review/                  # Review format selection
│   │   ├── selector.go          # Format selection logic
│   │   └── context.go           # Review context
│   └── session/                 # Review sessions
│       ├── builder.go           # Session construction
│       └── model.go             # Session data structures
├── migrations/                  # Database migrations
└── docker-compose.yml          # Local development setup
```

## Development

### Running Tests

```bash
go test ./...
```

### Code Style

Follow standard Go conventions. The project uses:
- Standard library `database/sql` for database access
- No ORM to maintain explicitness and performance
- Clear separation of concerns (MPS, review selection, AI, HTTP)

## Design Decisions

### Why Deterministic MPS?

- **Trust**: Users can understand why words are prioritized
- **Debugging**: Same inputs produce same outputs
- **No Surprises**: Predictable behavior reduces confusion

### Why Rule-Based Review Selection?

- **Low Latency**: No AI call needed during review
- **Cost Effective**: AI only used for word explanations
- **Extensible**: AI can augment later without breaking existing logic

### Why Asynchronous AI Calls?

Word creation responds immediately (< 100ms) while AI explanation happens in the background. This ensures:
- Fast user experience
- No blocking on external API calls
- Graceful degradation if AI fails

## Future Enhancements

- [ ] Proper JWT authentication
- [ ] Redis-based session storage
- [ ] Word frequency API integration
- [ ] Browser extension for word capture
- [ ] Spaced repetition analytics
- [ ] Confusion pair detection (AI-powered)
- [ ] Multiple language support

## License

[Add your license here]

## Contributing

[Add contribution guidelines here]
