package tasks

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/oge-math-trainer/oge-backend.git/internal/app"
)

type Worker struct {
	service  *Service
	repo     Repository
	minReady int
	interval time.Duration
}

func NewWorker(service *Service, repo Repository, minReady int, interval time.Duration) *Worker {
	return &Worker{service: service, repo: repo, minReady: minReady, interval: interval}
}

func (w *Worker) Run(ctx context.Context) error {
	if err := w.fillPreparedTasks(ctx); err != nil {
		log.Printf("prepared task worker initial fill failed: %v", err)
	}

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := w.fillPreparedTasks(ctx); err != nil {
				log.Printf("prepared task worker fill failed: %v", err)
			}
		}
	}
}

func (w *Worker) fillPreparedTasks(ctx context.Context) error {
	failures := 0
	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		target, ready, err := w.repo.GetPreparationTarget(ctx, w.minReady)
		if err != nil {
			var appErr *app.Error
			if errors.As(err, &appErr) && appErr.Code == app.CodeNotFound {
				log.Printf("prepared task worker: every target has at least %d ready tasks", w.minReady)
				return nil
			}
			return err
		}

		log.Printf("prepared task worker: target %d/%s has %d/%d ready tasks",
			target.OgeNumber, target.SubtypeCode, ready, w.minReady)
		prepared, err := w.service.PrepareTask(ctx, target)
		if err != nil {
			log.Printf("failed to prepare task for target %d/%s: %v", target.OgeNumber, target.SubtypeCode, err)
			failures++
			if failures >= 5 {
				return fmt.Errorf("too many preparation failures")
			}
			continue
		}
		failures = 0
		log.Printf("prepared task created: id=%d target=%d/%s source=%s", prepared.ID, prepared.OgeNumber, prepared.SubtypeCode, prepared.Source)
	}
}
