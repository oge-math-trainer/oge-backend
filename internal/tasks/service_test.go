package tasks

import (
	"context"
	"errors"
	"testing"

	"oge-ai-trainer/backend/internal/app"
)

func TestGenerateWithoutAIKeyReturnsAIUnavailable(t *testing.T) {
	repo := &fakeRepo{}
	service := NewService(repo, fakeAI{configured: false})
	oge := 9

	_, err := service.Generate(context.Background(), GenerateRequest{
		UserID:      1,
		Mode:        ModeCustom,
		OgeNumber:   &oge,
		SubtypeCode: "linear_equation",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	var appErr *app.Error
	if !errors.As(err, &appErr) || appErr.Code != app.CodeAIUnavailable {
		t.Fatalf("expected ai_unavailable, got %v", err)
	}
}

func TestGenerateCustomRequiresTargetFields(t *testing.T) {
	service := NewService(&fakeRepo{}, fakeAI{configured: false})

	_, err := service.Generate(context.Background(), GenerateRequest{UserID: 1, Mode: ModeCustom})
	if err == nil {
		t.Fatal("expected validation error")
	}
	var appErr *app.Error
	if !errors.As(err, &appErr) || appErr.Code != app.CodeValidation {
		t.Fatalf("expected validation_error, got %v", err)
	}
}

func TestCheckWithoutAIKeyReturnsAIUnavailable(t *testing.T) {
	repo := &fakeRepo{
		task: Task{ID: 1, UserID: 1, Mode: ModeWeak, OgeNumber: 9, SubtypeCode: "linear_equation", Question: "x + 2 = 5", CorrectAnswer: "3"},
	}
	service := NewService(repo, fakeAI{configured: false})

	_, err := service.Check(context.Background(), CheckRequest{
		UserID:        1,
		TaskID:        1,
		StudentAnswer: "3",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	var appErr *app.Error
	if !errors.As(err, &appErr) || appErr.Code != app.CodeAIUnavailable {
		t.Fatalf("expected ai_unavailable, got %v", err)
	}
}

func TestCheckStoresOriginalTaskMode(t *testing.T) {
	repo := &fakeRepo{
		task: Task{ID: 1, UserID: 1, Mode: ModeWeak, OgeNumber: 9, SubtypeCode: "linear_equation", Question: "x + 2 = 5", CorrectAnswer: "3"},
	}
	service := NewService(repo, fakeAI{configured: true})

	_, err := service.Check(context.Background(), CheckRequest{
		UserID:        1,
		TaskID:        1,
		StudentAnswer: "3",
	})
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if repo.attemptMode != ModeWeak {
		t.Fatalf("expected attempt mode %q, got %q", ModeWeak, repo.attemptMode)
	}
}

type fakeRepo struct {
	created     CreateTask
	task        Task
	attemptMode string
}

func (r *fakeRepo) GetWeakTarget(context.Context, int64) (Target, error) {
	return Target{OgeNumber: 9, SubtypeCode: "weak_topic"}, nil
}

func (r *fakeRepo) GetRandomTaskType(context.Context) (Target, error) {
	return Target{OgeNumber: 10, SubtypeCode: "random_topic"}, nil
}

func (r *fakeRepo) ResolveTarget(_ context.Context, target Target) (Target, error) {
	id := int64(7)
	target.TaskTypeID = &id
	return target, nil
}

func (r *fakeRepo) CreateGeneratedTask(_ context.Context, task CreateTask) (Task, error) {
	r.created = task
	return Task{
		ID:            100,
		UserID:        task.UserID,
		Mode:          task.Mode,
		OgeNumber:     task.OgeNumber,
		SubtypeCode:   task.SubtypeCode,
		Question:      task.Question,
		CorrectAnswer: task.CorrectAnswer,
		Source:        task.Source,
	}, nil
}

func (r *fakeRepo) GetGeneratedTaskForUser(_ context.Context, userID, _ int64) (Task, error) {
	if r.task.ID == 0 {
		return Task{}, app.NotFound("Задача не найдена")
	}
	if r.task.UserID != userID {
		return Task{}, app.NotFound("Задача не найдена")
	}
	return r.task, nil
}

func (r *fakeRepo) SaveAttempt(_ context.Context, _, _ int64, mode, _ string, _ bool, _ any) error {
	r.attemptMode = mode
	return nil
}

func (r *fakeRepo) UpdateProgress(context.Context, int64, Task, bool) error {
	return nil
}

type fakeAI struct {
	configured bool
}

func (f fakeAI) IsConfigured() bool {
	return f.configured
}

func (f fakeAI) GenerateTask(context.Context, Target) (GeneratedContent, error) {
	return GeneratedContent{Question: "generated", CorrectAnswer: "1", IsValid: true}, nil
}

func (f fakeAI) CheckAnswer(context.Context, Task, string) (CheckResult, error) {
	return CheckResult{IsCorrect: true}, nil
}

func (f fakeAI) Hint(context.Context, Task) (HintResult, error) {
	return HintResult{Hint: "hint"}, nil
}

func (f fakeAI) Explain(context.Context, Task) (ExplainResult, error) {
	return ExplainResult{Explanation: "explain"}, nil
}
