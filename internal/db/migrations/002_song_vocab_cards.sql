-- Add source tracking to cards table for song vocabulary integration
-- This allows song vocab to be mixed into daily lessons

-- Add source column to track where the card came from
ALTER TABLE cards ADD COLUMN source TEXT DEFAULT 'curriculum';

-- Add reference to source song (for song vocab cards)
ALTER TABLE cards ADD COLUMN source_song_id INTEGER REFERENCES songs(id);

-- Index for efficient filtering by source
CREATE INDEX IF NOT EXISTS idx_cards_source ON cards(source);
CREATE INDEX IF NOT EXISTS idx_cards_source_song ON cards(source_song_id);
