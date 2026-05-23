package tasks

import (
	"context"
	"log"
	"time"
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
	count, err := w.repo.CountPreparedTasks(ctx)
	if err != nil {
		return err
	}
	if count >= w.minReady {
		log.Printf("prepared task worker: enough tasks available (%d >= %d)", count, w.minReady)
		return nil
	}
	missing := w.minReady - count
	log.Printf("prepared task worker: under threshold, need %d more tasks", missing)

	for i := 0; i < missing; i++ {
		target, err := w.repo.GetRandomTaskType(ctx)
		if err != nil {
			return err
		}
		prepared, err := w.service.PrepareTask(ctx, target)
		if err != nil {
			log.Printf("failed to prepare task for target %d/%s: %v", target.OgeNumber, target.SubtypeCode, err)
			continue
		}
		log.Printf("prepared task created: id=%d target=%d/%s source=%s", prepared.ID, prepared.OgeNumber, prepared.SubtypeCode, prepared.Source)
	}
	return nil
}
