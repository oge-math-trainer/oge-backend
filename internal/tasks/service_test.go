package tasks

import (
	"context"
	"errors"
	"testing"

	"github.com/oge-math-trainer/oge-backend.git/internal/app"
)

func TestGenerateWithoutAIKeyReturnsAIUnavailable(t *testing.T) {
	repo := &fakeRepo{}
	service := NewService(repo, &fakeAI{configured: false})
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

func TestGenerateVisualFallbackWithoutAIKey(t *testing.T) {
	repo := &fakeRepo{}
	service := NewService(repo, &fakeAI{configured: false})
	oge := 11

	task, err := service.Generate(context.Background(), GenerateRequest{
		UserID:      1,
		Mode:        ModeCustom,
		OgeNumber:   &oge,
		SubtypeCode: "graphs_linear",
	})
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	if task.Source != "fallback" {
		t.Fatalf("expected fallback source, got %q", task.Source)
	}
	if len(repo.created.VisualData) == 0 {
		t.Fatal("expected fallback visual_data")
	}
	if got := repo.created.VisualData["type"]; got != string(VisualKindGraph) {
		t.Fatalf("expected graph visual_data, got %v", got)
	}
}

func TestGenerateRetriesInvalidVisualDataThenFallback(t *testing.T) {
	repo := &fakeRepo{}
	oge := 11
	ai := &fakeAI{
		configured: true,
		generated: []GeneratedContent{
			validGeneratedContent(VisualData{
				"type":   string(VisualKindGraph),
				"x_axis": map[string]any{"min": -5, "max": 5},
				"y_axis": map[string]any{"min": -5, "max": 5},
				"graphs": []any{
					map[string]any{"points": []any{point(0, 0), point(1, 1)}},
				},
			}),
			validGeneratedContent(VisualData{
				"type":   string(VisualKindGraph),
				"x_axis": map[string]any{"min": -5, "max": 5},
				"y_axis": map[string]any{"min": -5, "max": 5},
				"graphs": []any{
					map[string]any{"points": []any{point(0, 0)}},
				},
			}),
		},
	}
	service := NewService(repo, ai)

	task, err := service.Generate(context.Background(), GenerateRequest{
		UserID:      1,
		Mode:        ModeCustom,
		OgeNumber:   &oge,
		SubtypeCode: "graphs_linear",
	})
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	if ai.generateCalls != 2 {
		t.Fatalf("expected 2 AI calls, got %d", ai.generateCalls)
	}
	if task.Source != "fallback" {
		t.Fatalf("expected fallback source, got %q", task.Source)
	}
	if err := ValidateVisualData(Target{OgeNumber: 11, SubtypeCode: "graphs_linear"}, repo.created.VisualData); err != nil {
		t.Fatalf("fallback visual_data is invalid: %v", err)
	}
}

func TestGenerateRetriesAndUsesSecondValidVisualData(t *testing.T) {
	repo := &fakeRepo{}
	oge := 7
	validVisual := VisualData{
		"type":     string(VisualKindNumberLine),
		"axis":     map[string]any{"min": -5, "max": 5},
		"interval": map[string]any{"start": -1, "end": 4, "start_closed": true, "end_closed": false},
	}
	ai := &fakeAI{
		configured: true,
		generated: []GeneratedContent{
			validGeneratedContent(VisualData{
				"type":     string(VisualKindNumberLine),
				"axis":     map[string]any{"min": -5, "max": 5},
				"interval": map[string]any{"start": 4, "end": -1, "start_closed": true, "end_closed": false},
			}),
			validGeneratedContent(validVisual),
		},
	}
	service := NewService(repo, ai)

	task, err := service.Generate(context.Background(), GenerateRequest{
		UserID:      1,
		Mode:        ModeCustom,
		OgeNumber:   &oge,
		SubtypeCode: "numberline_compare",
	})
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	if ai.generateCalls != 2 {
		t.Fatalf("expected 2 AI calls, got %d", ai.generateCalls)
	}
	if task.Source != "qwen" {
		t.Fatalf("expected qwen source, got %q", task.Source)
	}
	if repo.created.VisualData["type"] != string(VisualKindNumberLine) {
		t.Fatalf("expected number_line visual_data, got %v", repo.created.VisualData["type"])
	}
}

func TestGenerateCustomRequiresTargetFields(t *testing.T) {
	service := NewService(&fakeRepo{}, &fakeAI{configured: false})

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
	service := NewService(repo, &fakeAI{configured: false})

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
	service := NewService(repo, &fakeAI{configured: true})

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
		VisualData:    task.VisualData,
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
	configured    bool
	generated     []GeneratedContent
	generateCalls int
}

func (f fakeAI) IsConfigured() bool {
	return f.configured
}

func (f *fakeAI) GenerateTask(context.Context, Target) (GeneratedContent, error) {
	f.generateCalls++
	if len(f.generated) >= f.generateCalls {
		return f.generated[f.generateCalls-1], nil
	}
	return validGeneratedContent(nil), nil
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

func validGeneratedContent(visualData VisualData) GeneratedContent {
	return GeneratedContent{
		Question:        "generated",
		CorrectAnswer:   "1",
		SolutionSteps:   []string{"step 1", "step 2", "step 3"},
		SelfCheck:       "check",
		IsValid:         true,
		ValidationNotes: "ok",
		VisualData:      visualData,
	}
}
