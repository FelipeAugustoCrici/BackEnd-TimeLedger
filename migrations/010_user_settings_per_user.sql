-- Adiciona user_id à tabela user_settings para isolar configurações por usuário
ALTER TABLE user_settings ADD COLUMN IF NOT EXISTS user_id UUID REFERENCES users(id) ON DELETE CASCADE;

-- Atribui a linha existente ao usuário padrão
UPDATE user_settings SET user_id = '5bc87aba-b618-4973-98a8-daac19c857ca' WHERE user_id IS NULL;

-- Garante unicidade: um registro de settings por usuário
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_settings_user_id ON user_settings (user_id);
