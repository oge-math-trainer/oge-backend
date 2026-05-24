package tasks

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/oge-math-trainer/oge-backend.git/internal/app"
)

func TestWorkerFillPausesAfterConsecutivePreparationFailures(t *testing.T) {
	repo := &fakeRepo{}
	ai := &fakeAI{configured: true, generateErr: errors.New("temporary ai failure")}
	service := NewService(repo, ai)
	worker := NewWorker(service, repo, 50, time.Minute)

	if err := worker.fillPreparedTasks(context.Background()); err != nil {
		t.Fatalf("expected fill to pause without stopping worker: %v", err)
	}
	if ai.generateCalls != 15 {
		t.Fatalf("expected 5 preparation attempts with 3 AI calls each, got %d AI calls", ai.generateCalls)
	}
}

func TestWorkerFillCompletesSelectedTargetBeforeNextTarget(t *testing.T) {
	repo := &sequencedPreparationRepo{
		responses: []preparationResponse{
			{target: Target{OgeNumber: 9, SubtypeCode: "equations_linear"}, ready: 48},
			{err: app.NotFound("done")},
		},
	}
	ai := &fakeAI{configured: true}
	service := NewService(repo, ai)
	worker := NewWorker(service, repo, 50, time.Minute)

	if err := worker.fillPreparedTasks(context.Background()); err != nil {
		t.Fatalf("fillPreparedTasks returned error: %v", err)
	}
	if repo.getCalls != 2 {
		t.Fatalf("expected worker to request a new target only after filling selected target, got %d calls", repo.getCalls)
	}
	if len(repo.createdTargets) != 2 {
		t.Fatalf("expected 2 prepared tasks to bring target from 48 to 50, got %d", len(repo.createdTargets))
	}
	for _, target := range repo.createdTargets {
		if target.OgeNumber != 9 || target.SubtypeCode != "equations_linear" {
			t.Fatalf("expected only selected target to be filled, got %+v", target)
		}
	}
}

type preparationResponse struct {
	target Target
	ready  int
	err    error
}

type sequencedPreparationRepo struct {
	fakeRepo
	responses      []preparationResponse
	getCalls       int
	createdTargets []Target
}

func (r *sequencedPreparationRepo) GetPreparationTarget(context.Context, int) (Target, int, error) {
	r.getCalls++
	if len(r.responses) == 0 {
		return Target{}, 0, app.NotFound("done")
	}
	response := r.responses[0]
	r.responses = r.responses[1:]
	return response.target, response.ready, response.err
}

func (r *sequencedPreparationRepo) CreatePreparedTask(ctx context.Context, task CreateTask) (PreparedTask, error) {
	r.createdTargets = append(r.createdTargets, Target{
		TaskTypeID:  task.TaskTypeID,
		OgeNumber:   task.OgeNumber,
		SubtypeCode: task.SubtypeCode,
	})
	return r.fakeRepo.CreatePreparedTask(ctx, task)
}
