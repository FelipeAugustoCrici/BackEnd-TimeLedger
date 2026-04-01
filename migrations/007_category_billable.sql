ALTER TABLE categories ADD COLUMN IF NOT EXISTS billable BOOLEAN NOT NULL DEFAULT true;
UPDATE categories SET billable = false WHERE name IN ('Pausa');
