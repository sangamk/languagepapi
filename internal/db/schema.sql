-- Spanish OS Database Schema
-- FSRS-based spaced repetition with gamification

-- Users table (single user for now, but extensible)
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    total_xp INTEGER DEFAULT 0,
    current_streak INTEGER DEFAULT 0,
    longest_streak INTEGER DEFAULT 0,
    last_active_date DATE
);

-- Islands (vocabulary categories/scripts)
CREATE TABLE IF NOT EXISTS islands (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT,
    icon TEXT,
    unlock_xp INTEGER DEFAULT 0,
    sort_order INTEGER DEFAULT 0
);

-- Cards (vocabulary items)
CREATE TABLE IF NOT EXISTS cards (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    island_id INTEGER REFERENCES islands(id),
    term TEXT NOT NULL,
    translation TEXT NOT NULL,
    example_sentence TEXT,
    notes TEXT,
    audio_url TEXT,
    frequency_rank INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Bridges (polyglot connections: Hindi/Dutch/English)
CREATE TABLE IF NOT EXISTS bridges (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    card_id INTEGER NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
    bridge_type TEXT NOT NULL CHECK(bridge_type IN ('hindi_phonetic', 'dutch_syntax', 'english_cognate')),
    bridge_content TEXT NOT NULL,
    explanation TEXT
);

-- Card Progress (FSRS scheduling data per user per card)
CREATE TABLE IF NOT EXISTS card_progress (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id),
    card_id INTEGER NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
    stability REAL DEFAULT 0,
    difficulty REAL DEFAULT 0,
    elapsed_days INTEGER DEFAULT 0,
    scheduled_days INTEGER DEFAULT 0,
    reps INTEGER DEFAULT 0,
    lapses INTEGER DEFAULT 0,
    state TEXT DEFAULT 'new' CHECK(state IN ('new', 'learning', 'review', 'relearning')),
    due DATETIME,
    last_review DATETIME,
    UNIQUE(user_id, card_id)
);

-- Review Logs (history for FSRS optimization and analytics)
CREATE TABLE IF NOT EXISTS review_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id),
    card_id INTEGER NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
    rating INTEGER NOT NULL CHECK(rating IN (1, 2, 3, 4)),
    elapsed_days INTEGER,
    scheduled_days INTEGER,
    reviewed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    review_duration_ms INTEGER
);

-- Daily Logs (for heat map and daily stats)
CREATE TABLE IF NOT EXISTS daily_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id),
    date DATE NOT NULL,
    xp_earned INTEGER DEFAULT 0,
    cards_reviewed INTEGER DEFAULT 0,
    cards_correct INTEGER DEFAULT 0,
    minutes_active INTEGER DEFAULT 0,
    new_cards_added INTEGER DEFAULT 0,
    UNIQUE(user_id, date)
);

-- Achievements (gamification badges)
CREATE TABLE IF NOT EXISTS achievements (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT,
    icon TEXT,
    xp_reward INTEGER DEFAULT 0,
    condition_type TEXT,
    condition_value INTEGER
);

-- User Achievements (earned badges)
CREATE TABLE IF NOT EXISTS user_achievements (
    user_id INTEGER NOT NULL REFERENCES users(id),
    achievement_id INTEGER NOT NULL REFERENCES achievements(id),
    earned_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY(user_id, achievement_id)
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_card_progress_due ON card_progress(due);
CREATE INDEX IF NOT EXISTS idx_card_progress_user_state ON card_progress(user_id, state);
CREATE INDEX IF NOT EXISTS idx_review_logs_user_card ON review_logs(user_id, card_id);
CREATE INDEX IF NOT EXISTS idx_daily_logs_user_date ON daily_logs(user_id, date);
CREATE INDEX IF NOT EXISTS idx_cards_island ON cards(island_id);
CREATE INDEX IF NOT EXISTS idx_cards_frequency ON cards(frequency_rank);
CREATE INDEX IF NOT EXISTS idx_bridges_card ON bridges(card_id);

-- Seed Data: Default user
INSERT OR IGNORE INTO users (id, username) VALUES (1, 'sangam');

-- Seed Data: Islands (vocabulary categories)
INSERT OR IGNORE INTO islands (id, name, description, icon, unlock_xp, sort_order) VALUES
    (1, 'Core Essentials', 'Top 100 most frequent words', '1', 0, 1),
    (2, 'Common Words', 'Words 101-250 by frequency', '2', 500, 2),
    (3, 'Expanding Vocabulary', 'Words 251-500 by frequency', '3', 1500, 3),
    (4, 'Advanced Vocabulary', 'Words 501-1000 by frequency', '4', 3000, 4),
    (5, 'Core Verbs', 'Essential action words', '5', 500, 5),
    (6, 'Career & Work', 'Professional vocabulary', '6', 1000, 6),
    (7, 'Social & Daily Life', 'Everyday conversations', '7', 1000, 7),
    (8, 'Past Tense', 'Narrating the past', '8', 2000, 8),
    (9, 'Subjunctive', 'Wishes, hopes, and uncertainty', '9', 5000, 9);

-- Seed Data: Sample achievements
INSERT OR IGNORE INTO achievements (id, name, description, icon, xp_reward, condition_type, condition_value) VALUES
    (1, 'First Steps', 'Review your first card', 'sparkles', 10, 'cards_reviewed', 1),
    (2, 'Getting Started', 'Review 10 cards', 'seedling', 25, 'cards_reviewed', 10),
    (3, 'Dedicated Learner', 'Review 100 cards', 'books', 100, 'cards_reviewed', 100),
    (4, 'Century Club', 'Review 500 cards', 'hundred', 250, 'cards_reviewed', 500),
    (5, 'Vocabulary Master', 'Review 1000 cards', 'crown', 500, 'cards_reviewed', 1000),
    (6, 'Week Warrior', '7-day streak', 'fire', 50, 'streak', 7),
    (7, 'Fortnight Fighter', '14-day streak', 'zap', 100, 'streak', 14),
    (8, 'Month Master', '30-day streak', 'trophy', 200, 'streak', 30),
    (9, 'Streak Legend', '100-day streak', 'star', 1000, 'streak', 100),
    (10, 'Word Collector', 'Learn 50 words', 'gem', 75, 'words_learned', 50),
    (11, 'Polyglot Path', 'Learn 200 words', 'rocket', 200, 'words_learned', 200),
    (12, 'Lexicon Lord', 'Learn 500 words', 'brain', 500, 'words_learned', 500);
