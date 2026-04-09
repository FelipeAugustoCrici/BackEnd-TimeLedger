package service

import (
	"controltasks/internal/model"
	"controltasks/internal/repository"
)

type EntryService struct {
	repo *repository.EntryRepository
}

func NewEntryService(repo *repository.EntryRepository) *EntryService {
	return &EntryService{repo: repo}
}

func (s *EntryService) List(f model.EntryFilters) ([]model.TaskEntry, error) {
	entries, err := s.repo.List(f)
	if entries == nil {
		entries = []model.TaskEntry{}
	}
	return entries, err
}

func (s *EntryService) GetByID(id string) (*model.TaskEntry, error) {
	return s.repo.GetByID(id)
}

func (s *EntryService) Create(in model.CreateTaskEntryInput) (*model.TaskEntry, error) {
	return s.repo.Create(in)
}

func (s *EntryService) Update(id string, in model.UpdateTaskEntryInput) (*model.TaskEntry, error) {
	return s.repo.Update(id, in)
}

func (s *EntryService) Delete(id string) error {
	return s.repo.Delete(id)
}

func (s *EntryService) ListProjects(userID string) ([]string, error) {
	p, err := s.repo.ListProjects(userID)
	if p == nil {
		p = []string{}
	}
	return p, err
}

func (s *EntryService) ListCategories(userID string) ([]string, error) {
	c, err := s.repo.ListCategories(userID)
	if c == nil {
		c = []string{}
	}
	return c, err
}

func (s *EntryService) Summary(userID, startDate, endDate string) (*model.DashboardSummary, error) {
	return s.repo.Summary(userID, startDate, endDate)
}

func (s *EntryService) ApplyRateToEntries(userID string, hourlyRate float64) (int, error) {
	return s.repo.ApplyRateToEntries(userID, hourlyRate)
}
