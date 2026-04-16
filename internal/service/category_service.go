package service

import (
	"encoding/json"
	"fmt"

	"controltasks/internal/model"
	"controltasks/internal/repository"
)

// CategoryCode representa um código mapeado para uma categoria
type CategoryCode struct {
	ID           string `json:"id"`
	Code         string `json:"code"`
	CategoryName string `json:"categoryName"`
	Description  string `json:"description,omitempty"`
}

type CategoryService struct {
	settingsRepo  *repository.SettingsRepository
	categoryRepo  *repository.CategoryRepository
}

func NewCategoryService(settingsRepo *repository.SettingsRepository, categoryRepo *repository.CategoryRepository) *CategoryService {
	return &CategoryService{
		settingsRepo:  settingsRepo,
		categoryRepo:  categoryRepo,
	}
}

// GetCategoryByCode retorna a categoria correspondente a um código específico
func (s *CategoryService) GetCategoryByCode(userID, code string) (string, error) {
	// Buscar configurações do usuário
	settings, err := s.settingsRepo.Get(userID)
	if err != nil {
		return "", fmt.Errorf("failed to get user settings: %w", err)
	}

	// Categoria padrão
	defaultCategory := "Desenvolvimento"
	if settings.DefaultCategoryName != nil {
		defaultCategory = *settings.DefaultCategoryName
	}

	// Se não há códigos configurados, retorna a padrão
	if settings.CategoryCodes == nil || *settings.CategoryCodes == "" {
		return defaultCategory, nil
	}

	// Parse dos códigos configurados
	var categoryCodes []CategoryCode
	if err := json.Unmarshal([]byte(*settings.CategoryCodes), &categoryCodes); err != nil {
		// Se erro no parse, retorna categoria padrão
		return defaultCategory, nil
	}

	// Buscar código específico (case insensitive)
	for _, categoryCode := range categoryCodes {
		if categoryCode.Code == code {
			// Verificar se a categoria ainda existe
			categories, err := s.categoryRepo.List()
			if err != nil {
				return defaultCategory, nil
			}

			// Validar se categoria existe
			for _, cat := range categories {
				if cat.Name == categoryCode.CategoryName {
					return categoryCode.CategoryName, nil
				}
			}
		}
	}

	// Se código não encontrado ou categoria não existe mais, retorna padrão
	return defaultCategory, nil
}

// ValidateCategoryExists verifica se uma categoria existe para o usuário
func (s *CategoryService) ValidateCategoryExists(userID, categoryName string) (bool, error) {
	categories, err := s.categoryRepo.List()
	if err != nil {
		return false, err
	}

	for _, cat := range categories {
		if cat.Name == categoryName {
			return true, nil
		}
	}

	return false, nil
}

// GetAvailableCategories retorna todas as categorias disponíveis para o usuário
func (s *CategoryService) GetAvailableCategories(userID string) ([]model.Category, error) {
	return s.categoryRepo.List()
}

// SuggestCategoryForCode sugere uma categoria baseada no código
func (s *CategoryService) SuggestCategoryForCode(userID, code string) (*CategorySuggestion, error) {
	categoryName, err := s.GetCategoryByCode(userID, code)
	if err != nil {
		return nil, err
	}

	// Buscar configurações para verificar se é customizado
	settings, err := s.settingsRepo.Get(userID)
	if err != nil {
		return nil, err
	}

	defaultCategory := "Desenvolvimento"
	if settings.DefaultCategoryName != nil {
		defaultCategory = *settings.DefaultCategoryName
	}

	isCustom := categoryName != defaultCategory

	// Buscar informações da categoria
	categories, err := s.categoryRepo.List()
	if err != nil {
		return nil, err
	}

	var categoryInfo *model.Category
	for _, cat := range categories {
		if cat.Name == categoryName {
			categoryInfo = &cat
			break
		}
	}

	return &CategorySuggestion{
		CategoryName: categoryName,
		IsCustom:     isCustom,
		CategoryInfo: categoryInfo,
	}, nil
}

// CategorySuggestion representa uma sugestão de categoria
type CategorySuggestion struct {
	CategoryName string           `json:"categoryName"`
	IsCustom     bool             `json:"isCustom"`
	CategoryInfo *model.Category  `json:"categoryInfo,omitempty"`
}