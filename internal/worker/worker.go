package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/joshjms/castletown/internal/cache"
	"github.com/joshjms/castletown/internal/config"
	"github.com/joshjms/castletown/internal/grader"
	"github.com/joshjms/castletown/internal/models"
	"github.com/joshjms/castletown/internal/mq"
	"github.com/joshjms/castletown/internal/repository"
	"github.com/joshjms/castletown/internal/sandbox"
	"github.com/joshjms/castletown/internal/telemetry"
	"github.com/rs/zerolog"
)

type SubmissionGrader interface {
	Handle(ctx context.Context, sub *models.Submission) error
}

type Worker struct {
	log     zerolog.Logger
	metrics *telemetry.Metrics

	sm *sandbox.Manager
	g  SubmissionGrader

	queueConsumer mq.Consumer
}

func NewWorker(cfg config.Config) *Worker {
	sm := sandbox.NewManager(&cfg)
	repo, _ := repository.NewPostgresRepository(cfg.Database.DSN)
	_ = os.MkdirAll(cfg.ProblemCacheDir, 0755)
	store := cache.NewProblemStore(repo, 256, cfg.ProblemCacheDir)

	log := zerolog.New(os.Stdout).With().Timestamp().Logger()

	w := &Worker{
		log:     log,
		metrics: telemetry.NewMetricsRegistry(),
		sm:      sm,
		g:       grader.NewGrader(log, cfg, sm, repo, store),
	}

	w.queueConsumer = mq.NewConsumer(cfg.RabbitMQ, log, cfg.MaxParallelSandboxes)

	return w
}

func (w *Worker) Run(ctx context.Context) error {
	if w.queueConsumer == nil {
		<-ctx.Done()
		return ctx.Err()
	}

	w.queueConsumer.Run(ctx, w.handleQueueMessage)
	return ctx.Err()
}

func (w *Worker) handle(ctx context.Context, sub *models.Submission) error {
	if w.g == nil {
		return nil
	}
	return w.g.Handle(ctx, sub)
}

func (w *Worker) handleQueueMessage(ctx context.Context, body []byte) error {
	var sub models.Submission
	if err := json.Unmarshal(body, &sub); err != nil {
		return fmt.Errorf("invalid submission payload: %w", err)
	}
	w.log.Info().Int64("submission_id", sub.ID).Msg("Processing submission")
	return w.handle(ctx, &sub)
}
