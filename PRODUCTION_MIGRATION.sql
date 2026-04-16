-- Migração para produção: Adicionar suporte a categorias automáticas
-- Execute este SQL no banco de produção (Railway)

-- Verificar se as colunas já existem
SELECT column_name 
FROM information_schema.columns 
WHERE table_name = 'user_settings' 
AND column_name IN ('default_category_name', 'category_codes');

-- Adicionar as novas colunas se não existirem
ALTER TABLE user_settings 
ADD COLUMN IF NOT EXISTS default_category_name VARCHAR(255) DEFAULT 'Desenvolvimento',
ADD COLUMN IF NOT EXISTS category_codes TEXT DEFAULT '[]';

-- Verificar se foram criadas com sucesso
SELECT column_name, data_type, column_default 
FROM information_schema.columns 
WHERE table_name = 'user_settings' 
AND column_name IN ('default_category_name', 'category_codes');

-- Verificar estrutura completa da tabela
\d user_settings;