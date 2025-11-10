package repository

import (
	"context"

	"github.com/joshjms/castletown/internal/models"
)

type Repository interface {
	GetProblemDetails(ctx context.Context, problemID int64) (*models.Problem, error)
	GetTestcases(ctx context.Context, problemID int64) ([]models.Testcase, error)
}
