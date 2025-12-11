-- Add color column to users table
ALTER TABLE users ADD COLUMN color VARCHAR(7);

-- Set default colors for existing users (generate based on user_id hash)
-- This will be handled by the application for new users
UPDATE users SET color = '#' || SUBSTRING(MD5(user_id) FROM 1 FOR 6) WHERE color IS NULL;

-- Make color NOT NULL after setting defaults
ALTER TABLE users ALTER COLUMN color SET NOT NULL;

