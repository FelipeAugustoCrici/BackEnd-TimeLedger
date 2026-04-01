-- Adiciona meta de ganho mensal nas configurações do usuário
ALTER TABLE user_settings
  ADD COLUMN IF NOT EXISTS monthly_goal TEXT NOT NULL DEFAULT '';
