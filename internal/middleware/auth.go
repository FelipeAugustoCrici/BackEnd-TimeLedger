package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"controltasks/internal/auth"
	"controltasks/internal/service"
)

// Auth valida o JWT e verifica se a sessão está ativa no banco.
func Auth(authSvc *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		bearer := c.GetHeader("Authorization")
		if !strings.HasPrefix(bearer, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token não informado"})
			return
		}

		token := strings.TrimPrefix(bearer, "Bearer ")

		// 1. Valida assinatura e expiração
		claims, err := auth.Validate(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token inválido ou expirado"})
			return
		}

		// 2. Verifica se a sessão ainda está ativa (não foi revogada via logout)
		active, err := authSvc.ValidateSession(auth.Hash(token))
		if err != nil || !active {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "sessão expirada ou revogada"})
			return
		}

		c.Set("claims", claims)
		c.Next()
	}
}
