-- Executado automaticamente no startup da API.
-- Todas as operações usam IF NOT EXISTS — seguro rodar múltiplas vezes.

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

DO $$ BEGIN
  CREATE TYPE task_status AS ENUM ('pending', 'in_progress', 'done');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

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

CREATE TABLE IF NOT EXISTS user_settings (
  id                    UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id               UUID        UNIQUE REFERENCES users(id) ON DELETE CASCADE,
  hourly_rate           TEXT        NOT NULL DEFAULT '0',
  daily_hours_goal      TEXT        NOT NULL DEFAULT '8',
  monthly_goal          TEXT        NOT NULL DEFAULT '',
  default_category_name VARCHAR(255) DEFAULT 'Desenvolvimento',
  category_codes        TEXT        DEFAULT '[]',
  updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_user_settings_user_id ON user_settings (user_id);

-- Adicionar colunas de categoria se não existirem (para bancos existentes)
ALTER TABLE user_settings 
ADD COLUMN IF NOT EXISTS default_category_name VARCHAR(255) DEFAULT 'Desenvolvimento',
ADD COLUMN IF NOT EXISTS category_codes TEXT DEFAULT '[]';

CREATE TABLE IF NOT EXISTS task_entries (
  id                 UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id            UUID        REFERENCES users(id) ON DELETE CASCADE,
  date               DATE        NOT NULL,
  task_code          VARCHAR(50) NOT NULL,
  description        TEXT        NOT NULL,
  time_spent_minutes INT         NOT NULL CHECK (time_spent_minutes > 0),
  hourly_rate        TEXT        NOT NULL,
  total_amount       TEXT        NOT NULL,
  status             task_status NOT NULL DEFAULT 'done',
  category           VARCHAR(100),
  project            VARCHAR(100),
  notes              TEXT,
  start_time         VARCHAR(5)  DEFAULT NULL,
  end_time           VARCHAR(5)  DEFAULT NULL,
  created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_task_entries_date    ON task_entries (date DESC);
CREATE INDEX IF NOT EXISTS idx_task_entries_user_id ON task_entries (user_id);
CREATE INDEX IF NOT EXISTS idx_task_entries_status  ON task_entries (status);
CREATE INDEX IF NOT EXISTS idx_task_entries_project ON task_entries (project);

CREATE TABLE IF NOT EXISTS categories (
  id         UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
  name       VARCHAR(100) NOT NULL UNIQUE,
  color      VARCHAR(7)   NOT NULL DEFAULT '#6366f1',
  billable   BOOLEAN      NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

INSERT INTO categories (name, color, billable) VALUES
  ('Trabalho',  '#3b82f6', true),
  ('Reunião',   '#8b5cf6', true),
  ('Pausa',     '#64748b', false),
  ('Estudo',    '#f59e0b', true),
  ('Pessoal',   '#10b981', true),
  ('Urgente',   '#ef4444', true),
  ('Frontend',  '#06b6d4', true),
  ('Backend',   '#6366f1', true),
  ('DevOps',    '#f97316', true),
  ('QA',        '#84cc16', true)
ON CONFLICT (name) DO NOTHING;

CREATE TABLE IF NOT EXISTS active_timers (
  id               UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id          UUID        NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
  status           VARCHAR(10) NOT NULL CHECK (status IN ('running', 'paused')),
  started_at       TIMESTAMPTZ,
  elapsed_seconds  INT         NOT NULL DEFAULT 0,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_active_timers_user_id ON active_timers (user_id);
