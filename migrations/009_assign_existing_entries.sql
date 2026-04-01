-- Atribui todos os lançamentos sem user_id ao usuário padrão
UPDATE task_entries
SET user_id = '5bc87aba-b618-4973-98a8-daac19c857ca'
WHERE user_id IS NULL;
