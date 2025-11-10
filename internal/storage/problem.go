package storage

import (
	"github.com/joshjms/castletown/internal/models"
	"github.com/joshjms/castletown/internal/repository"
)

type ProblemStorage struct {
	Problem map[string]*models.Problem
	repo    repository.Repository
}

func NewProblemStorage(repo repository.Repository) *ProblemStorage {
	return &ProblemStorage{
		Problem: make(map[string]*models.Problem),
		repo:    repo,
	}
}
