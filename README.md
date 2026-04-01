# ControlTasks — Backend

API REST em Go para o sistema de controle de horas e lançamentos por usuário.

## Stack

- **Go 1.22** com [Gin](https://github.com/gin-gonic/gin)
- **PostgreSQL 16** via `lib/pq`
- **JWT** para autenticação (`golang-jwt/jwt/v5`)
- **AES-GCM** para criptografia dos campos financeiros
- **Docker + Docker Compose** para ambiente completo

---

## Estrutura

```
.
├── cmd/api/            # Entrypoint da aplicação
├── internal/
│   ├── auth/           # Geração e validação de JWT
│   ├── crypto/         # Criptografia AES-GCM dos campos financeiros
│   ├── db/             # Conexão com o banco
│   ├── handler/        # Controllers HTTP (Gin)
│   ├── middleware/      # Auth JWT + CORS
│   ├── model/          # Structs de domínio e inputs
│   ├── repository/     # Queries SQL
│   └── service/        # Regras de negócio
├── migrations/         # Migrations SQL em ordem numérica
├── .env.example        # Variáveis de ambiente necessárias
├── docker-compose.yml  # Sobe API + PostgreSQL
└── Dockerfile          # Build multi-stage
```

---

## Variáveis de ambiente

Copie `.env.example` para `.env` e ajuste os valores:

```env
APP_PORT=8080
APP_ENV=development

DB_HOST=localhost
DB_PORT=5432
DB_USER=controltasks
DB_PASSWORD=controltasks
DB_NAME=controltasks
DB_SSLMODE=disable

ALLOWED_ORIGINS=http://localhost:5173,http://localhost:4173

JWT_SECRET=sua-chave-secreta-longa-e-aleatoria
FIELD_ENCRYPT_KEY=chave-de-32-bytes-para-aes-gcm
```

> `FIELD_ENCRYPT_KEY` precisa ter exatamente 32 caracteres (AES-256).

---

## Rodando com Docker (recomendado)

Sobe o banco e a API juntos:

```bash
docker compose up -d --build
```

A API ficará disponível em `http://localhost:8080`.

Para ver os logs:

```bash
docker compose logs -f api
```

Para parar:

```bash
docker compose down
```

---

## Rodando localmente (sem Docker)

Pré-requisitos: Go 1.22+ e PostgreSQL rodando.

```bash
# Instala dependências
go mod download

# Cria o .env
cp .env.example .env

# Aplica as migrations manualmente no banco
psql -U controltasks -d controltasks -f migrations/001_init.sql
psql -U controltasks -d controltasks -f migrations/002_auth.sql
# ... repita para todas as migrations em ordem

# Sobe a API
go run ./cmd/api
```

---

## Migrations

As migrations ficam em `migrations/` e são aplicadas em ordem numérica. Quando o banco é criado via Docker pela primeira vez, todas são aplicadas automaticamente via `docker-entrypoint-initdb.d`.

Para bancos já existentes, aplique manualmente:

```bash
docker exec -i controltasks_db psql -U controltasks -d controltasks < migrations/008_user_id_entries.sql
```

| Arquivo | Descrição |
|---|---|
| `001_init.sql` | Tabelas base: `task_entries`, `user_settings` |
| `002_auth.sql` | Tabelas de autenticação: `users`, `sessions` |
| `003_encrypt_financial.sql` | Migra campos financeiros para texto criptografado |
| `004_monthly_goal.sql` | Adiciona meta mensal em `user_settings` |
| `005_entry_time_range.sql` | Adiciona `start_time` e `end_time` em `task_entries` |
| `006_categories.sql` | Tabela de categorias customizáveis |
| `007_category_billable.sql` | Flag `billable` nas categorias |
| `008_user_id_entries.sql` | Adiciona `user_id` em `task_entries` |
| `009_assign_existing_entries.sql` | Atribui lançamentos existentes ao usuário padrão |
| `010_user_settings_per_user.sql` | Isola `user_settings` por usuário |

---

## Endpoints

Todas as rotas protegidas exigem o header:
```
Authorization: Bearer <token>
```

### Auth

| Método | Rota | Descrição |
|---|---|---|
| `POST` | `/api/v1/auth/register` | Cadastro de usuário |
| `POST` | `/api/v1/auth/login` | Login, retorna JWT |
| `POST` | `/api/v1/auth/logout` | Revoga a sessão atual |
| `GET` | `/api/v1/auth/me` | Dados do usuário autenticado |

### Lançamentos

| Método | Rota | Descrição |
|---|---|---|
| `GET` | `/api/v1/entries` | Lista lançamentos do usuário |
| `POST` | `/api/v1/entries` | Cria um lançamento |
| `GET` | `/api/v1/entries/:id` | Busca lançamento por ID |
| `PUT` | `/api/v1/entries/:id` | Atualiza lançamento |
| `DELETE` | `/api/v1/entries/:id` | Remove lançamento |
| `GET` | `/api/v1/entries/meta/projects` | Lista projetos distintos do usuário |
| `GET` | `/api/v1/entries/meta/categories` | Lista categorias distintas do usuário |

Filtros disponíveis via query string em `GET /entries`:

| Parâmetro | Exemplo | Descrição |
|---|---|---|
| `period` | `week` | Atalho: `today`, `week`, `month` |
| `start_date` | `2026-03-01` | Data inicial (formato `YYYY-MM-DD`) |
| `end_date` | `2026-03-31` | Data final |
| `status` | `done` | `pending`, `in_progress`, `done` |
| `category` | `Backend` | Nome da categoria |
| `project` | `MeuProjeto` | Nome do projeto |
| `search` | `login` | Busca em código e descrição |

### Dashboard

| Método | Rota | Descrição |
|---|---|---|
| `GET` | `/api/v1/dashboard` | Resumo de horas e valor do período |

Query string: `period=today|week|month`

### Configurações

| Método | Rota | Descrição |
|---|---|---|
| `GET` | `/api/v1/settings` | Retorna configurações do usuário |
| `PUT` | `/api/v1/settings` | Atualiza configurações |

### Categorias

| Método | Rota | Descrição |
|---|---|---|
| `GET` | `/api/v1/categories` | Lista categorias |
| `POST` | `/api/v1/categories` | Cria categoria |
| `PUT` | `/api/v1/categories/:id` | Atualiza categoria |
| `DELETE` | `/api/v1/categories/:id` | Remove categoria |

---

## Segurança

- Senhas armazenadas com **bcrypt**
- Tokens JWT com expiração de 24h, validados contra a tabela `sessions` (permite revogação via logout)
- Campos `hourly_rate` e `total_amount` criptografados com **AES-256-GCM** no banco
- Todos os dados de lançamentos e configurações são isolados por `user_id`
