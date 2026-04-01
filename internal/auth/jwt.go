package auth

import (
	"crypto/sha256"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"controltasks/internal/model"
)

const tokenDuration = 24 * time.Hour

type jwtClaims struct {
	model.Claims
	jwt.RegisteredClaims
}

func secret() []byte {
	s := os.Getenv("JWT_SECRET")
	if s == "" {
		s = "change-me-in-production"
	}
	return []byte(s)
}

// Generate cria um JWT assinado e retorna o token + expiração.
func Generate(user model.User) (string, time.Time, error) {
	exp := time.Now().Add(tokenDuration)

	claims := jwtClaims{
		Claims: model.Claims{
			UserID: user.ID,
			Email:  user.Email,
			Name:   user.Name,
		},
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(secret())
	return signed, exp, err
}

// Validate valida o JWT e retorna os claims.
func Validate(tokenStr string) (*model.Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &jwtClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("método de assinatura inesperado: %v", t.Header["alg"])
		}
		return secret(), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*jwtClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("token inválido")
	}

	return &claims.Claims, nil
}

// Hash retorna o SHA-256 do token para armazenar na tabela sessions.
func Hash(token string) string {
	h := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", h)
}
