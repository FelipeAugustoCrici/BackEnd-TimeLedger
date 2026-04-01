package service

import (
	"errors"

	"golang.org/x/crypto/bcrypt"

	"controltasks/internal/auth"
	"controltasks/internal/model"
	"controltasks/internal/repository"
)

var (
	ErrEmailTaken     = errors.New("e-mail já cadastrado")
	ErrInvalidCreds   = errors.New("e-mail ou senha incorretos")
	ErrUserNotFound   = errors.New("usuário não encontrado")
)

type AuthService struct {
	repo *repository.AuthRepository
}

func NewAuthService(repo *repository.AuthRepository) *AuthService {
	return &AuthService{repo: repo}
}

// Register cria um novo usuário com senha hasheada.
func (s *AuthService) Register(in model.RegisterInput) (*model.AuthResponse, error) {
	exists, err := s.repo.EmailExists(in.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrEmailTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user, err := s.repo.CreateUser(in.Name, in.Email, string(hash))
	if err != nil {
		return nil, err
	}

	return s.issueToken(*user)
}

// Login valida credenciais e emite um token.
func (s *AuthService) Login(in model.LoginInput) (*model.AuthResponse, error) {
	user, err := s.repo.GetUserByEmail(in.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCreds
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(in.Password)); err != nil {
		return nil, ErrInvalidCreds
	}

	return s.issueToken(*user)
}

// Logout revoga a sessão pelo token hash.
func (s *AuthService) Logout(tokenHash string) error {
	return s.repo.DeleteSession(tokenHash)
}

// ValidateSession verifica se o token está ativo no banco.
func (s *AuthService) ValidateSession(tokenHash string) (bool, error) {
	return s.repo.SessionExists(tokenHash)
}

// Me retorna o usuário pelo ID.
func (s *AuthService) Me(userID string) (*model.User, error) {
	u, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, ErrUserNotFound
	}
	return u, nil
}

// issueToken gera JWT e persiste a sessão.
func (s *AuthService) issueToken(user model.User) (*model.AuthResponse, error) {
	token, expiresAt, err := auth.Generate(user)
	if err != nil {
		return nil, err
	}

	if err := s.repo.CreateSession(user.ID, auth.Hash(token), expiresAt); err != nil {
		return nil, err
	}

	return &model.AuthResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      user,
	}, nil
}
