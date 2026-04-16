# Sistema de Categorias e Códigos - Backend

Implementação completa do sistema de categorias automáticas no backend Go.

## 📊 Mudanças no Banco de Dados

### Nova Migração: `012_category_settings.sql`
```sql
ALTER TABLE user_settings 
ADD COLUMN IF NOT EXISTS default_category_name VARCHAR(255) DEFAULT 'Desenvolvimento',
ADD COLUMN IF NOT EXISTS category_codes TEXT DEFAULT '[]';
```

### Estrutura dos Dados
- **default_category_name**: Nome da categoria padrão do usuário
- **category_codes**: JSON string com array de objetos CategoryCode

## 🏗️ Mudanças no Backend

### 1. Modelo Atualizado (`model/model.go`)
```go
type UserSettings struct {
    ID                  string    `json:"id"`
    HourlyRate          float64   `json:"hourly_rate"`
    DailyHoursGoal      float64   `json:"daily_hours_goal"`
    MonthlyGoal         float64   `json:"monthly_goal"`
    DefaultCategoryName *string   `json:"default_category_name,omitempty"`
    CategoryCodes       *string   `json:"category_codes,omitempty"` // JSON string
    UpdatedAt           time.Time `json:"updated_at"`
}

type UpdateSettingsInput struct {
    HourlyRate          float64 `json:"hourly_rate"      binding:"required,gt=0"`
    DailyHoursGoal      float64 `json:"daily_hours_goal" binding:"required,gt=0"`
    MonthlyGoal         float64 `json:"monthly_goal"`
    DefaultCategoryName *string `json:"default_category_name"`
    CategoryCodes       *string `json:"category_codes"` // JSON string
}
```

### 2. Repository Atualizado (`repository/settings_repository.go`)
- **Get()**: Inclui novas colunas na query SELECT
- **Update()**: Inclui novas colunas na query INSERT/UPDATE
- **Valores padrão**: "Desenvolvimento" e "[]" para novos usuários

### 3. Novo Service (`service/category_service.go`)
```go
type CategoryService struct {
    settingsRepo  *repository.SettingsRepository
    categoryRepo  *repository.CategoryRepository
}

// Principais métodos:
func (s *CategoryService) GetCategoryByCode(userID, code string) (string, error)
func (s *CategoryService) ValidateCategoryExists(userID, categoryName string) (bool, error)
func (s *CategoryService) SuggestCategoryForCode(userID, code string) (*CategorySuggestion, error)
```

### 4. Novo Handler (`handler/category_code_handler.go`)
```go
// Endpoints disponíveis:
GET  /categories/by-code/:code  // Busca categoria por código
POST /categories/suggest        // Sugere categoria com informações completas
GET  /categories/available      // Lista categorias disponíveis
```

## 🔧 Estrutura JSON dos CategoryCodes

```json
[
  {
    "id": "1640995200000",
    "code": "#48263",
    "categoryName": "Reunião", 
    "description": "Reuniões de projeto"
  },
  {
    "id": "1640995300000",
    "code": "#ADMIN",
    "categoryName": "Administrativo",
    "description": "Tarefas administrativas"
  }
]
```

## 🚀 Como Usar

### 1. Executar Migração
```bash
# Execute a migração no seu banco de dados
psql -d seu_banco < migrations/012_category_settings.sql
```

### 2. Configurar Rotas (exemplo)
```go
// No seu arquivo de rotas
categoryService := service.NewCategoryService(settingsRepo, categoryRepo)
categoryCodeHandler := handler.NewCategoryCodeHandler(categoryService)

api.GET("/categories/by-code/:code", categoryCodeHandler.GetCategoryByCode)
api.POST("/categories/suggest", categoryCodeHandler.SuggestCategory)
api.GET("/categories/available", categoryCodeHandler.GetAvailableCategories)
```

### 3. Usar no Frontend
```typescript
// O frontend já está configurado para usar essas APIs
const categoryName = await getCategoryByCode('#48263');
// → "Reunião" (se configurado) ou "Desenvolvimento" (padrão)
```

## 🔄 Fluxo de Funcionamento

1. **Usuário configura** categoria padrão e códigos no Settings
2. **Frontend salva** via PUT /settings (JSON é armazenado no banco)
3. **Ao criar lançamento**, frontend consulta categoria por código
4. **Backend busca** nos category_codes ou retorna categoria padrão
5. **Validação** se categoria ainda existe no sistema

## ✅ Funcionalidades Implementadas

- ✅ **Armazenamento** de categoria padrão e códigos no banco
- ✅ **API** para buscar categoria por código
- ✅ **Validação** se categoria existe
- ✅ **Sugestões** com informações completas da categoria
- ✅ **Compatibilidade** com sistema de categorias existente
- ✅ **Migração** segura com valores padrão

## 🔧 Próximos Passos

1. **Execute a migração** `012_category_settings.sql`
2. **Configure as rotas** no seu router
3. **Teste** as APIs com Postman/Insomnia
4. **Frontend** já está pronto para usar

O sistema está completamente integrado e pronto para uso!