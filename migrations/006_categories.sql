CREATE TABLE IF NOT EXISTS categories (
  id         UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
  name       VARCHAR(100) NOT NULL UNIQUE,
  color      VARCHAR(7)   NOT NULL DEFAULT '#6366f1',  -- hex color
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
