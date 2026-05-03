package diagnostic

import (
	"context"
	"errors"
	"testing"

	"github.com/oge-math-trainer/oge-backend.git/internal/app"
	"github.com/oge-math-trainer/oge-backend.git/internal/tasks"
)

func TestSubmitChecksSessionOwnerBeforeAI(t *testing.T) {
	repo := &fakeRepo{ownerErr: app.NotFound("Диагностическая сессия не найдена")}
	service := NewService(repo, fakeTaskGenerator{}, fakeAI{configured: false})

	_, err := service.Submit(context.Background(), 1, 99, []AnswerInput{
		{TaskID: 1, StudentAnswer: "3"},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	var appErr *app.Error
	if !errors.As(err, &appErr) || appErr.Code != app.CodeNotFound {
		t.Fatalf("expected not_found before ai_unavailable, got %v", err)
	}
	if !repo.ownerChecked {
		t.Fatal("expected owner check")
	}
}

type fakeRepo struct {
	ownerErr     error
	ownerChecked bool
}

func (r *fakeRepo) CreateDiagnosticSession(context.Context, int64) (int64, error) {
	return 1, nil
}

func (r *fakeRepo) EnsureDiagnosticSessionOwner(context.Context, int64, int64) error {
	r.ownerChecked = true
	return r.ownerErr
}

func (r *fakeRepo) GetDiagnosticTargets(context.Context) ([]tasks.Target, error) {
	return nil, nil
}

func (r *fakeRepo) GetGeneratedTaskForUser(context.Context, int64, int64) (tasks.Task, error) {
	return tasks.Task{ID: 1, UserID: 1, Mode: tasks.ModeDiagnostic, OgeNumber: 9, SubtypeCode: "linear_equation"}, nil
}

func (r *fakeRepo) SaveDiagnosticAnswer(context.Context, int64, tasks.Task, string, bool, any) error {
	return nil
}

func (r *fakeRepo) FinishDiagnosticSession(context.Context, int64, Analysis, []string) error {
	return nil
}

func (r *fakeRepo) UpdateProgress(context.Context, int64, tasks.Task, bool) error {
	return nil
}

type fakeTaskGenerator struct{}

func (fakeTaskGenerator) Generate(context.Context, tasks.GenerateRequest) (tasks.Task, error) {
	return tasks.Task{}, nil
}

type fakeAI struct {
	configured bool
}

func (f fakeAI) IsConfigured() bool {
	return f.configured
}

func (f fakeAI) CheckAnswer(context.Context, tasks.Task, string) (tasks.CheckResult, error) {
	return tasks.CheckResult{IsCorrect: true}, nil
}

func (f fakeAI) AnalyzeDiagnostic(context.Context, []AnswerResult) (Analysis, error) {
	return Analysis{Summary: "ok"}, nil
}
