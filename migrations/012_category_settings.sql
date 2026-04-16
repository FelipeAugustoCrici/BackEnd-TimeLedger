-- Adiciona suporte a categoria padrão e códigos personalizados nas configurações do usuário

-- Adicionar colunas para categoria padrão e códigos
ALTER TABLE user_settings 
ADD COLUMN IF NOT EXISTS default_category_name VARCHAR(255) DEFAULT 'Desenvolvimento',
ADD COLUMN IF NOT EXISTS category_codes TEXT DEFAULT '[]';

-- Comentários sobre as colunas
COMMENT ON COLUMN user_settings.default_category_name IS 'Nome da categoria padrão do usuário (deve corresponder a uma categoria existente)';
COMMENT ON COLUMN user_settings.category_codes IS 'JSON string contendo array de objetos CategoryCode para mapeamento código -> categoria';

-- Exemplo de dados que serão armazenados em category_codes:
-- [
--   {
--     "id": "1640995200000",
--     "code": "#48263", 
--     "categoryName": "Reunião",
--     "description": "Reuniões de projeto"
--   }
-- ]