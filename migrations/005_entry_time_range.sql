-- Adiciona horário de início e fim nos lançamentos
ALTER TABLE task_entries
  ADD COLUMN IF NOT EXISTS start_time VARCHAR(5) DEFAULT NULL,
  ADD COLUMN IF NOT EXISTS end_time   VARCHAR(5) DEFAULT NULL;
-- Formato: "HH:MM" ex: "09:00", "17:30"
