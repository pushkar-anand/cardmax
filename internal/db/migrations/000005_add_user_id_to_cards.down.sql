-- Revert adding user_id column and foreign key from cards table (SQLite compatible)

PRAGMA foreign_keys=off;

-- Create a temporary table without the user_id column and FK constraint
CREATE TABLE cards_old (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    question TEXT NOT NULL,
    answer TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Copy data from the current cards table to the temporary table
INSERT INTO cards_old (id, question, answer, created_at)
SELECT id, question, answer, created_at
FROM cards;

-- Drop the current cards table (which includes user_id and the FK)
DROP TABLE cards;

-- Rename the temporary table to the original table name
ALTER TABLE cards_old RENAME TO cards;

-- Add back any indexes that might have existed on the original table (if needed)
-- e.g., CREATE INDEX ...

PRAGMA foreign_keys=on;
