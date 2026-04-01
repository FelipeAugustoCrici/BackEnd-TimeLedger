-- Adiciona user_id à tabela task_entries para isolar lançamentos por usuário
ALTER TABLE task_entries ADD COLUMN IF NOT EXISTS user_id UUID REFERENCES users(id) ON DELETE CASCADE;

-- Cria índice para performance nas queries filtradas por usuário
CREATE INDEX IF NOT EXISTS idx_task_entries_user_id ON task_entries (user_id);
