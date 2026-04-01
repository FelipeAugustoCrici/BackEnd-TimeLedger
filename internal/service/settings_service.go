package service

import (
	"controltasks/internal/model"
	"controltasks/internal/repository"
)

type SettingsService struct {
	repo *repository.SettingsRepository
}

func NewSettingsService(repo *repository.SettingsRepository) *SettingsService {
	return &SettingsService{repo: repo}
}

func (s *SettingsService) Get(userID string) (*model.UserSettings, error) {
	return s.repo.Get(userID)
}

func (s *SettingsService) Update(userID string, in model.UpdateSettingsInput) (*model.UserSettings, error) {
	return s.repo.Update(userID, in)
}
