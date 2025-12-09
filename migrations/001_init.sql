-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Users table
CREATE TABLE IF NOT EXISTS users (
    user_id VARCHAR(128) PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    display_name VARCHAR(255),
    department VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Memos table
CREATE TABLE IF NOT EXISTS memos (
    memo_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(128) REFERENCES users(user_id) ON DELETE CASCADE,
    user_name VARCHAR(255) NOT NULL,
    title VARCHAR(255),
    audio_url TEXT NOT NULL,
    text TEXT NOT NULL,
    duration_seconds INTEGER,
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    location_accuracy FLOAT,
    address TEXT,
    park_name VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_memos_user_created ON memos(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_memos_location ON memos(latitude, longitude);
CREATE INDEX IF NOT EXISTS idx_memos_park ON memos(park_name);
CREATE INDEX IF NOT EXISTS idx_memos_created ON memos(created_at DESC);

-- Full-text search index on text
CREATE INDEX IF NOT EXISTS idx_memos_text_search ON memos USING gin(to_tsvector('english', text));

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger to auto-update updated_at
DROP TRIGGER IF EXISTS update_memos_updated_at ON memos;
CREATE TRIGGER update_memos_updated_at 
    BEFORE UPDATE ON memos 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

