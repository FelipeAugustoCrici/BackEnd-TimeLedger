package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"controltasks/internal/service"
)

type CategoryCodeHandler struct {
	categorySvc *service.CategoryService
}

func NewCategoryCodeHandler(categorySvc *service.CategoryService) *CategoryCodeHandler {
	return &CategoryCodeHandler{categorySvc: categorySvc}
}

// GET /categories/by-code/:code
// Retorna a categoria correspondente a um código específico
func (h *CategoryCodeHandler) GetCategoryByCode(c *gin.Context) {
	userID, ok := userIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuário não autenticado"})
		return
	}

	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "código é obrigatório"})
		return
	}

	categoryName, err := h.categorySvc.GetCategoryByCode(userID, code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"code":         code,
			"categoryName": categoryName,
		},
	})
}

// POST /categories/suggest
// Sugere categoria baseada no código com informações completas
func (h *CategoryCodeHandler) SuggestCategory(c *gin.Context) {
	userID, ok := userIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuário não autenticado"})
		return
	}

	var req struct {
		Code string `json:"code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	suggestion, err := h.categorySvc.SuggestCategoryForCode(userID, req.Code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": suggestion})
}

// GET /categories/available
// Lista todas as categorias disponíveis para o usuário
func (h *CategoryCodeHandler) GetAvailableCategories(c *gin.Context) {
	userID, ok := userIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuário não autenticado"})
		return
	}

	categories, err := h.categorySvc.GetAvailableCategories(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": categories})
}