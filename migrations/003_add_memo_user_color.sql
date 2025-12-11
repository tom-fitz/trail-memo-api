-- Add user_color column to memos table
ALTER TABLE memos ADD COLUMN user_color VARCHAR(7);

-- Set default colors for existing memos by joining with users table
UPDATE memos m
SET user_color = u.color
FROM users u
WHERE m.user_id = u.user_id;

-- Make user_color NOT NULL after setting defaults
ALTER TABLE memos ALTER COLUMN user_color SET NOT NULL;

