-- ─── 001: Init ───────────────────────────────────────────────────────────────
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

DO $$ BEGIN
  CREATE TYPE task_status AS ENUM ('pending', 'in_progress', 'done');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

CREATE TABLE IF NOT EXISTS user_settings (
  id               UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
  hourly_rate      NUMERIC(10,2) NOT NULL DEFAULT 0,
  daily_hours_goal NUMERIC(4,2)  NOT NULL DEFAULT 8,
  updated_at       TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS task_entries (
  id                 UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
  date               DATE        NOT NULL,
  task_code          VARCHAR(50) NOT NULL,
  description        TEXT        NOT NULL,
  time_spent_minutes INT         NOT NULL CHECK (time_spent_minutes > 0),
  hourly_rate        NUMERIC(10,2) NOT NULL,
  total_amount       NUMERIC(12,2) NOT NULL,
  status             task_status NOT NULL DEFAULT 'done',
  category           VARCHAR(100),
  project            VARCHAR(100),
  notes              TEXT,
  created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_task_entries_date     ON task_entries (date DESC);
CREATE INDEX IF NOT EXISTS idx_task_entries_status   ON task_entries (status);
CREATE INDEX IF NOT EXISTS idx_task_entries_project  ON task_entries (project);
CREATE INDEX IF NOT EXISTS idx_task_entries_category ON task_entries (category);

-- ─── 002: Auth ────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS users (
  id            UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
  name          VARCHAR(100) NOT NULL,
  email         VARCHAR(150) NOT NULL UNIQUE,
  password_hash TEXT         NOT NULL,
  created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS sessions (
  id         UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash TEXT        NOT NULL UNIQUE,
  expires_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sessions_token_hash ON sessions (token_hash);
CREATE INDEX IF NOT EXISTS idx_sessions_user_id    ON sessions (user_id);

-- ─── 003: Encrypt financial fields ───────────────────────────────────────────
ALTER TABLE task_entries
  ALTER COLUMN hourly_rate  TYPE TEXT USING hourly_rate::TEXT,
  ALTER COLUMN total_amount TYPE TEXT USING total_amount::TEXT;

ALTER TABLE user_settings
  ALTER COLUMN hourly_rate      TYPE TEXT USING hourly_rate::TEXT,
  ALTER COLUMN daily_hours_goal TYPE TEXT USING daily_hours_goal::TEXT;

-- ─── 004: Monthly goal ────────────────────────────────────────────────────────
ALTER TABLE user_settings
  ADD COLUMN IF NOT EXISTS monthly_goal TEXT NOT NULL DEFAULT '';

-- ─── 005: Entry time range ────────────────────────────────────────────────────
ALTER TABLE task_entries
  ADD COLUMN IF NOT EXISTS start_time VARCHAR(5) DEFAULT NULL,
  ADD COLUMN IF NOT EXISTS end_time   VARCHAR(5) DEFAULT NULL;

-- ─── 006: Categories ─────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS categories (
  id         UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
  name       VARCHAR(100) NOT NULL UNIQUE,
  color      VARCHAR(7)   NOT NULL DEFAULT '#6366f1',
  created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

INSERT INTO categories (name, color) VALUES
  ('Trabalho',  '#3b82f6'),
  ('Reunião',   '#8b5cf6'),
  ('Pausa',     '#64748b'),
  ('Estudo',    '#f59e0b'),
  ('Pessoal',   '#10b981'),
  ('Urgente',   '#ef4444'),
  ('Frontend',  '#06b6d4'),
  ('Backend',   '#6366f1'),
  ('DevOps',    '#f97316'),
  ('QA',        '#84cc16')
ON CONFLICT (name) DO NOTHING;

-- ─── 007: Category billable ──────────────────────────────────────────────────
ALTER TABLE categories ADD COLUMN IF NOT EXISTS billable BOOLEAN NOT NULL DEFAULT true;
UPDATE categories SET billable = false WHERE name IN ('Pausa');

-- ─── 008: User ID on entries ─────────────────────────────────────────────────
ALTER TABLE task_entries ADD COLUMN IF NOT EXISTS user_id UUID REFERENCES users(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_task_entries_user_id ON task_entries (user_id);

-- ─── 009: (skip - no existing data in new DB) ────────────────────────────────

-- ─── 010: User settings per user ─────────────────────────────────────────────
ALTER TABLE user_settings ADD COLUMN IF NOT EXISTS user_id UUID REFERENCES users(id) ON DELETE CASCADE;
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_settings_user_id ON user_settings (user_id);

-- ─── 011: Active timers ──────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS active_timers (
  id              UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id         UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  status          VARCHAR(20) NOT NULL DEFAULT 'running' CHECK (status IN ('running', 'paused')),
  started_at      TIMESTAMPTZ,
  elapsed_seconds INT         NOT NULL DEFAULT 0,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_active_timers_user_id ON active_timers (user_id);

-- ─── 012: Category settings ──────────────────────────────────────────────────
ALTER TABLE user_settings 
ADD COLUMN IF NOT EXISTS default_category_name VARCHAR(255) DEFAULT 'Desenvolvimento',
ADD COLUMN IF NOT EXISTS category_codes TEXT DEFAULT '[]';
