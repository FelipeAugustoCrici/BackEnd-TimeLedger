-- Converte colunas financeiras de NUMERIC para TEXT (armazenará ciphertext base64)
-- Executar apenas uma vez. Dados existentes serão perdidos (re-inserir após migração).

ALTER TABLE task_entries
  ALTER COLUMN hourly_rate TYPE TEXT USING hourly_rate::TEXT,
  ALTER COLUMN total_amount TYPE TEXT USING total_amount::TEXT;

ALTER TABLE user_settings
  ALTER COLUMN hourly_rate      TYPE TEXT USING hourly_rate::TEXT,
  ALTER COLUMN daily_hours_goal TYPE TEXT USING daily_hours_goal::TEXT;
