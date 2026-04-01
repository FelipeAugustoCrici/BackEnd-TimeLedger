package repository

import (
	"database/sql"
	"time"

	"controltasks/internal/model"
)

type AuthRepository struct {
	db *sql.DB
}

func NewAuthRepository(db *sql.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

// CreateUser insere um novo usuário.
func (r *AuthRepository) CreateUser(name, email, passwordHash string) (*model.User, error) {
	var u model.User
	err := r.db.QueryRow(`
		INSERT INTO users (name, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, name, email, created_at, updated_at`,
		name, email, passwordHash,
	).Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt, &u.UpdatedAt)
	return &u, err
}

// GetUserByEmail busca usuário pelo e-mail.
func (r *AuthRepository) GetUserByEmail(email string) (*model.User, error) {
	var u model.User
	err := r.db.QueryRow(`
		SELECT id, name, email, password_hash, created_at, updated_at
		FROM users WHERE email = $1`, email,
	).Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &u, err
}

// GetUserByID busca usuário pelo ID.
func (r *AuthRepository) GetUserByID(id string) (*model.User, error) {
	var u model.User
	err := r.db.QueryRow(`
		SELECT id, name, email, created_at, updated_at
		FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &u, err
}

// CreateSession armazena uma sessão ativa.
func (r *AuthRepository) CreateSession(userID, tokenHash string, expiresAt time.Time) error {
	_, err := r.db.Exec(`
		INSERT INTO sessions (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)`,
		userID, tokenHash, expiresAt,
	)
	return err
}

// SessionExists verifica se o token hash existe e não expirou.
func (r *AuthRepository) SessionExists(tokenHash string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM sessions
			WHERE token_hash = $1 AND expires_at > NOW()
		)`, tokenHash,
	).Scan(&exists)
	return exists, err
}

// DeleteSession remove a sessão (logout).
func (r *AuthRepository) DeleteSession(tokenHash string) error {
	_, err := r.db.Exec(`DELETE FROM sessions WHERE token_hash = $1`, tokenHash)
	return err
}

// DeleteExpiredSessions limpa sessões expiradas.
func (r *AuthRepository) DeleteExpiredSessions() error {
	_, err := r.db.Exec(`DELETE FROM sessions WHERE expires_at <= NOW()`)
	return err
}

// EmailExists verifica se o e-mail já está cadastrado.
func (r *AuthRepository) EmailExists(email string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`, email).Scan(&exists)
	return exists, err
}
