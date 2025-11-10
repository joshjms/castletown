package grader

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/joshjms/castletown/internal/cache"
	"github.com/joshjms/castletown/internal/config"
	"github.com/joshjms/castletown/internal/models"
	"github.com/joshjms/castletown/internal/repository"
	"github.com/joshjms/castletown/internal/sandbox"
	"github.com/rs/zerolog"
)

type Grader struct {
	log zerolog.Logger
	cfg config.Config

	sm    *sandbox.Manager
	repo  repository.Repository
	store *cache.ProblemStore
}

func NewGrader(log zerolog.Logger, cfg config.Config, sm *sandbox.Manager, repo repository.Repository, store *cache.ProblemStore) *Grader {
	return &Grader{
		log:   log,
		cfg:   cfg,
		sm:    sm,
		repo:  repo,
		store: store,
	}
}

func (g *Grader) Handle(ctx context.Context, sub *models.Submission) error {
	var (
		problem   *models.Problem
		testcases []models.Testcase
		err       error
	)

	if g.store != nil {
		problem, testcases, err = g.store.GetProblem(ctx, sub.ProblemID)
	} else {
		problem, err = g.repo.GetProblemDetails(ctx, sub.ProblemID)
		if err != nil {
			return err
		}

		testcases, err = g.repo.GetTestcases(ctx, sub.ProblemID)
	}
	if err != nil {
		return err
	}

	submissionDir := filepath.Join(g.cfg.WorkRoot, fmt.Sprintf("submission_%d", sub.ID))
	if err := os.Mkdir(submissionDir, 0700); err != nil {
		return err
	}
	defer os.RemoveAll(submissionDir)

	switch sub.Language {
	case "cpp":
		return g.handleCpp(ctx, sub, problem, testcases, submissionDir)
	default:
		return fmt.Errorf("unsupported language: %s", sub.Language)
	}
}
