CREATE TABLE users (
    id UUID PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE words (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    text TEXT NOT NULL,
    context TEXT,
    source TEXT,
    confidence SMALLINT CHECK (confidence BETWEEN 1 AND 5),
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX idx_words_user ON words(user_id); -- Note index 1
CREATE INDEX idx_words_text ON words(text); -- Note index 2

CREATE TABLE word_ai_data ( -- Avoids polluting core table
    word_id UUID PRIMARY KEY REFERENCES words(id),
    definition TEXT NOT NULL,
    example_good TEXT NOT NULL,
    example_bad TEXT,
    pos TEXT,
    translation TEXT,
    cefr_level TEXT,
    generated_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE reviews (
    id UUID PRIMARY KEY,
    word_id UUID NOT NULL REFERENCES words(id),
    user_id UUID NOT NULL REFERENCES users(id),
    result BOOLEAN NOT NULL,
    review_type TEXT NOT NULL, -- "mcq", "typing", "fill_blank"
    reviewed_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX idx_reviews_word ON reviews(word_id); -- Note index 3
CREATE INDEX idx_reviews_user ON reviews(user_id); -- Note index 4

CREATE TABLE review_stats ( -- Avoids heavy aggregation queries and allows instant review queue generation
    word_id UUID PRIMARY KEY REFERENCES words(id),
    total_reviews INTEGER NOT NULL DEFAULT 0,
    correct_reviews INTEGER NOT NULL DEFAULT 0,
    last_reviewed_at TIMESTAMP,
    accuracy_rate FLOAT NOT NULL DEFAULT 0,
    memory_score FLOAT NOT NULL DEFAULT 0
);

CREATE TABLE review_queue ( -- Rebuilt: once per day or after review session
    user_id UUID NOT NULL,
    word_id UUID NOT NULL,
    priority_score FLOAT NOT NULL,
    reason TEXT,
    PRIMARY KEY (user_id, word_id)
);

-- Do not score formula logic in DB
-- This allows for flexibility in the scoring formula and makes it easier to change the formula
-- The formula should be stored in the code and applied to the data in the code

