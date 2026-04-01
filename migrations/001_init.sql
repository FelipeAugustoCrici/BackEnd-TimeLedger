-- ─── Extensions ──────────────────────────────────────────────────────────────
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ─── Enum: task_status ────────────────────────────────────────────────────────
DO $$ BEGIN
  CREATE TYPE task_status AS ENUM ('pending', 'in_progress', 'done');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

-- ─── Table: user_settings ─────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS user_settings (
  id               UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
  hourly_rate      NUMERIC(10,2) NOT NULL DEFAULT 0,
  daily_hours_goal NUMERIC(4,2)  NOT NULL DEFAULT 8,
  updated_at       TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

-- Garante que sempre existe exatamente uma linha de configuração
INSERT INTO user_settings (hourly_rate, daily_hours_goal)
SELECT 120, 8
WHERE NOT EXISTS (SELECT 1 FROM user_settings);

-- ─── Table: task_entries ──────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS task_entries (
  id                 UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
  date               DATE          NOT NULL,
  task_code          VARCHAR(50)   NOT NULL,
  description        TEXT          NOT NULL,
  time_spent_minutes INT           NOT NULL CHECK (time_spent_minutes > 0),
  hourly_rate        NUMERIC(10,2) NOT NULL,
  total_amount       NUMERIC(12,2) NOT NULL,
  status             task_status   NOT NULL DEFAULT 'done',
  category           VARCHAR(100),
  project            VARCHAR(100),
  notes              TEXT,
  created_at         TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
  updated_at         TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

-- ─── Indexes ──────────────────────────────────────────────────────────────────
CREATE INDEX IF NOT EXISTS idx_task_entries_date     ON task_entries (date DESC);
CREATE INDEX IF NOT EXISTS idx_task_entries_status   ON task_entries (status);
CREATE INDEX IF NOT EXISTS idx_task_entries_project  ON task_entries (project);
CREATE INDEX IF NOT EXISTS idx_task_entries_category ON task_entries (category);
