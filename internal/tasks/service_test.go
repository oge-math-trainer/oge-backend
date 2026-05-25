package tasks

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/oge-math-trainer/oge-backend.git/internal/app"
)

func TestGenerateWithoutPreparedTaskGeneratesOnDemand(t *testing.T) {
	repo := &fakeRepo{}
	ai := &fakeAI{configured: true}
	service := NewService(repo, ai)
	oge := 9

	task, err := service.Generate(context.Background(), GenerateRequest{
		UserID:      1,
		Mode:        ModeCustom,
		OgeNumber:   &oge,
		SubtypeCode: "linear_equation",
	})
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	if task.Question != "generated" {
		t.Fatalf("expected on-demand generated task, got %q", task.Question)
	}
	if task.Source != "gpt-4o-mini" {
		t.Fatalf("expected gpt-4o-mini source, got %q", task.Source)
	}
	if ai.generateCalls != 1 {
		t.Fatalf("expected one foreground AI call, got %d", ai.generateCalls)
	}
	if repo.created.UserID != 1 || repo.created.OgeNumber != 9 || repo.created.SubtypeCode != "linear_equation" {
		t.Fatalf("unexpected created task: %+v", repo.created)
	}
	if repo.prepared == nil {
		t.Fatal("expected on-demand task to be cached as prepared task")
	}
	if repo.prepared.Question != "generated" || repo.prepared.Source != "gpt-4o-mini" {
		t.Fatalf("unexpected cached prepared task: %+v", repo.prepared)
	}
	if repo.createPreparedCalls != 1 {
		t.Fatalf("expected one prepared task cache write, got %d", repo.createPreparedCalls)
	}
	if ai.reviewCalls != 0 {
		t.Fatalf("expected on-demand generation to skip worker review, got %d review calls", ai.reviewCalls)
	}
}

func TestGenerateOnDemandIgnoresDuplicatePreparedCache(t *testing.T) {
	repo := &fakeRepo{createPreparedErrs: []error{app.Conflict("duplicate prepared task")}}
	ai := &fakeAI{configured: true}
	service := NewService(repo, ai)
	oge := 9

	task, err := service.Generate(context.Background(), GenerateRequest{
		UserID:      1,
		Mode:        ModeCustom,
		OgeNumber:   &oge,
		SubtypeCode: "linear_equation",
	})
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	if task.Question != "generated" {
		t.Fatalf("expected on-demand generated task, got %q", task.Question)
	}
	if repo.createPreparedCalls != 1 {
		t.Fatalf("expected one prepared task cache write, got %d", repo.createPreparedCalls)
	}
}

func TestPrepareTaskVisualFallbackWithoutAIKey(t *testing.T) {
	repo := &fakeRepo{}
	service := NewService(repo, &fakeAI{configured: false})

	task, err := service.PrepareTask(context.Background(), Target{OgeNumber: 11, SubtypeCode: "graphs_linear"})
	if err != nil {
		t.Fatalf("PrepareTask returned error: %v", err)
	}
	if task.Source != "fallback" {
		t.Fatalf("expected fallback source, got %q", task.Source)
	}
	if len(repo.prepared.VisualData) == 0 {
		t.Fatal("expected fallback visual_data")
	}
	if got := repo.prepared.VisualData["type"]; got != string(VisualKindGraph) {
		t.Fatalf("expected graph visual_data, got %v", got)
	}
}

func TestPrepareTaskRetriesInvalidVisualDataThenFallback(t *testing.T) {
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

	task, err := service.PrepareTask(context.Background(), Target{OgeNumber: oge, SubtypeCode: "graphs_linear"})
	if err != nil {
		t.Fatalf("PrepareTask returned error: %v", err)
	}
	if ai.generateCalls != 3 {
		t.Fatalf("expected 3 AI calls, got %d", ai.generateCalls)
	}
	if task.Source != "fallback" {
		t.Fatalf("expected fallback source, got %q", task.Source)
	}
	if err := ValidateVisualData(Target{OgeNumber: 11, SubtypeCode: "graphs_linear"}, repo.prepared.VisualData); err != nil {
		t.Fatalf("fallback visual_data is invalid: %v", err)
	}
}

func TestPrepareTaskRetriesAndUsesSecondValidVisualData(t *testing.T) {
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

	task, err := service.PrepareTask(context.Background(), Target{OgeNumber: oge, SubtypeCode: "numberline_compare"})
	if err != nil {
		t.Fatalf("PrepareTask returned error: %v", err)
	}
	if ai.generateCalls != 2 {
		t.Fatalf("expected 2 AI calls, got %d", ai.generateCalls)
	}
	if task.Source != "gpt-4o-mini" {
		t.Fatalf("expected gpt-4o-mini source, got %q", task.Source)
	}
	if repo.prepared.VisualData["type"] != string(VisualKindNumberLine) {
		t.Fatalf("expected number_line visual_data, got %v", repo.prepared.VisualData["type"])
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

func TestGenerateUsesPreparedTaskFromCache(t *testing.T) {
	prepared := &PreparedTask{
		ID:              123,
		Mode:            ModeCustom,
		OgeNumber:       9,
		SubtypeCode:     "linear_equation",
		Question:        "cached question",
		CorrectAnswer:   "3",
		SolutionSteps:   []string{"step 1", "step 2", "step 3"},
		SelfCheck:       "check",
		IsValid:         true,
		ValidationNotes: "ok",
		Source:          "prepared",
	}
	repo := &fakeRepo{prepared: prepared}
	service := NewService(repo, &fakeAI{configured: true})
	oge := 9

	task, err := service.Generate(context.Background(), GenerateRequest{
		UserID:      1,
		Mode:        ModeCustom,
		OgeNumber:   &oge,
		SubtypeCode: "linear_equation",
	})
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	if task.Question != "cached question" {
		t.Fatalf("expected cached question, got %q", task.Question)
	}
	if task.Source != "prepared" {
		t.Fatalf("expected prepared source, got %q", task.Source)
	}
	if task.Mode != ModeCustom {
		t.Fatalf("expected mode %q, got %q", ModeCustom, task.Mode)
	}
	if repo.created.UserID != 1 {
		t.Fatalf("expected created user id 1, got %d", repo.created.UserID)
	}
	if repo.preparedUserID != 1 {
		t.Fatalf("expected prepared lookup user id 1, got %d", repo.preparedUserID)
	}
}

func TestGenerateRejectsInvalidPreparedTaskAndFallsBackOnDemand(t *testing.T) {
	prepared := &PreparedTask{
		ID:              123,
		Mode:            ModeCustom,
		OgeNumber:       9,
		SubtypeCode:     "linear_equation",
		Question:        `\begin{tikzpicture}\draw (0,0) -- (1,0);\end{tikzpicture}`,
		CorrectAnswer:   "3",
		SolutionSteps:   []string{"step 1", "step 2", "step 3"},
		SelfCheck:       "check",
		IsValid:         true,
		ValidationNotes: "ok",
		Source:          "prepared",
	}
	repo := &fakeRepo{prepared: prepared}
	ai := &fakeAI{configured: true}
	service := NewService(repo, ai)
	oge := 9

	task, err := service.Generate(context.Background(), GenerateRequest{
		UserID:      1,
		Mode:        ModeCustom,
		OgeNumber:   &oge,
		SubtypeCode: "linear_equation",
	})
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	if task.Question != "generated" {
		t.Fatalf("expected on-demand generated task, got %q", task.Question)
	}
	if ai.generateCalls != 1 {
		t.Fatalf("expected one foreground AI call, got %d", ai.generateCalls)
	}
}

func TestGenerateSkipsDuplicatePreparedTaskForUser(t *testing.T) {
	repo := &fakeRepo{
		preparedQueue: []PreparedTask{
			validPreparedTask(1, "duplicate question"),
			validPreparedTask(2, "fresh question"),
		},
		createGeneratedErrs: []error{app.Conflict("duplicate generated task"), nil},
	}
	ai := &fakeAI{configured: true}
	service := NewService(repo, ai)
	oge := 9

	task, err := service.Generate(context.Background(), GenerateRequest{
		UserID:      1,
		Mode:        ModeCustom,
		OgeNumber:   &oge,
		SubtypeCode: "linear_equation",
	})
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	if task.Question != "fresh question" {
		t.Fatalf("expected second prepared task, got %q", task.Question)
	}
	if ai.generateCalls != 0 {
		t.Fatalf("expected no on-demand AI calls, got %d", ai.generateCalls)
	}
}

func TestGenerateFallsBackOnDemandAfterDuplicatePreparedTasks(t *testing.T) {
	repo := &fakeRepo{
		preparedQueue: []PreparedTask{
			validPreparedTask(1, "duplicate question 1"),
			validPreparedTask(2, "duplicate question 2"),
		},
		createGeneratedErrs: []error{
			app.Conflict("duplicate generated task"),
			app.Conflict("duplicate generated task"),
		},
	}
	ai := &fakeAI{configured: true}
	service := NewService(repo, ai)
	oge := 9

	task, err := service.Generate(context.Background(), GenerateRequest{
		UserID:      1,
		Mode:        ModeCustom,
		OgeNumber:   &oge,
		SubtypeCode: "linear_equation",
	})
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	if task.Question != "generated" {
		t.Fatalf("expected on-demand generated task, got %q", task.Question)
	}
	if ai.generateCalls != 1 {
		t.Fatalf("expected one foreground AI call, got %d", ai.generateCalls)
	}
}

func TestPrepareTaskStoresValidatedPreparedTask(t *testing.T) {
	repo := &fakeRepo{}
	service := NewService(repo, &fakeAI{configured: true})
	target := Target{TaskTypeID: ptrInt64(7), OgeNumber: 9, SubtypeCode: "linear_equation"}

	prepared, err := service.PrepareTask(context.Background(), target)
	if err != nil {
		t.Fatalf("PrepareTask returned error: %v", err)
	}
	if prepared.ID == 0 {
		t.Fatal("expected saved prepared task id")
	}
	if prepared.Source != "gpt-4o-mini" {
		t.Fatalf("expected source gpt-4o-mini, got %q", prepared.Source)
	}
	if repo.prepared == nil {
		t.Fatal("expected prepared task to be stored in repo")
	}
	if repo.prepared.OgeNumber != 9 || repo.prepared.SubtypeCode != "linear_equation" {
		t.Fatalf("unexpected prepared task target: %+v", repo.prepared)
	}
}

func TestPrepareTaskRetriesDuplicatePreparedContent(t *testing.T) {
	repo := &fakeRepo{createPreparedErrs: []error{app.Conflict("duplicate prepared task"), nil}}
	ai := &fakeAI{configured: true}
	service := NewService(repo, ai)

	_, err := service.PrepareTask(context.Background(), Target{TaskTypeID: ptrInt64(7), OgeNumber: 9, SubtypeCode: "linear_equation"})
	if err != nil {
		t.Fatalf("PrepareTask returned error: %v", err)
	}
	if ai.generateCalls != 2 {
		t.Fatalf("expected retry after duplicate prepared task, got %d calls", ai.generateCalls)
	}
}

func TestPrepareTaskRunsWorkerReviewBeforeStoring(t *testing.T) {
	repo := &fakeRepo{}
	ai := &fakeAI{configured: true}
	service := NewService(repo, ai)

	_, err := service.PrepareTask(context.Background(), Target{TaskTypeID: ptrInt64(7), OgeNumber: 9, SubtypeCode: "equations_linear"})
	if err != nil {
		t.Fatalf("PrepareTask returned error: %v", err)
	}
	if ai.reviewCalls != 1 {
		t.Fatalf("expected one worker review call, got %d", ai.reviewCalls)
	}
}

func TestPrepareTaskRetriesWhenWorkerReviewRejects(t *testing.T) {
	repo := &fakeRepo{}
	ai := &fakeAI{
		configured: true,
		reviews: []GenerationReview{
			{IsValid: false, Reason: "ответ не совпадает с решением", CorrectAnswer: "5"},
			{IsValid: true, CorrectAnswer: "1"},
		},
	}
	service := NewService(repo, ai)

	_, err := service.PrepareTask(context.Background(), Target{TaskTypeID: ptrInt64(7), OgeNumber: 9, SubtypeCode: "equations_linear"})
	if err != nil {
		t.Fatalf("PrepareTask returned error: %v", err)
	}
	if ai.generateCalls != 2 {
		t.Fatalf("expected generation retry after review rejection, got %d calls", ai.generateCalls)
	}
	if ai.reviewCalls != 2 {
		t.Fatalf("expected two worker review calls, got %d", ai.reviewCalls)
	}
	if len(ai.feedbacks) < 2 || !strings.Contains(ai.feedbacks[1], "reviewer rejected") {
		t.Fatalf("expected reviewer feedback on retry, got %#v", ai.feedbacks)
	}
}

func TestPrepareTaskPassesDuplicateFeedbackToNextGeneration(t *testing.T) {
	repo := &fakeRepo{createPreparedErrs: []error{app.Conflict("duplicate prepared task"), nil}}
	ai := &fakeAI{configured: true}
	service := NewService(repo, ai)

	_, err := service.PrepareTask(context.Background(), Target{TaskTypeID: ptrInt64(7), OgeNumber: 9, SubtypeCode: "equations_linear"})
	if err != nil {
		t.Fatalf("PrepareTask returned error: %v", err)
	}
	if len(ai.feedbacks) < 2 {
		t.Fatalf("expected feedback on retry, got %#v", ai.feedbacks)
	}
	if !strings.Contains(ai.feedbacks[1], "duplicate") {
		t.Fatalf("expected duplicate feedback on retry, got %q", ai.feedbacks[1])
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
	prepared            *PreparedTask
	preparedQueue       []PreparedTask
	preparedUserID      int64
	created             CreateTask
	task                Task
	attemptMode         string
	createGeneratedErrs []error
	createPreparedErrs  []error
	createPreparedCalls int
}

func (r *fakeRepo) TryAcquirePreparationLock(context.Context) (func(), bool, error) {
	return func() {}, true, nil
}

func (r *fakeRepo) GetWeakTarget(context.Context, int64) (Target, error) {
	return Target{OgeNumber: 9, SubtypeCode: "weak_topic"}, nil
}

func (r *fakeRepo) CountPreparedTasks(context.Context) (int, error) {
	return 0, nil
}

func (r *fakeRepo) CreatePreparedTask(_ context.Context, task CreateTask) (PreparedTask, error) {
	r.createPreparedCalls++
	if len(r.createPreparedErrs) > 0 {
		err := r.createPreparedErrs[0]
		r.createPreparedErrs = r.createPreparedErrs[1:]
		if err != nil {
			return PreparedTask{}, err
		}
	}
	prepared := PreparedTask{
		ID:              100,
		Mode:            task.Mode,
		TaskTypeID:      task.TaskTypeID,
		OgeNumber:       task.OgeNumber,
		SubtypeCode:     task.SubtypeCode,
		Question:        task.Question,
		CorrectAnswer:   task.CorrectAnswer,
		SolutionSteps:   task.SolutionSteps,
		SelfCheck:       task.SelfCheck,
		IsValid:         task.IsValid,
		ValidationNotes: task.ValidationNotes,
		Source:          task.Source,
		VisualData:      task.VisualData,
		Graphs:          task.Graphs,
	}
	r.prepared = &prepared
	return prepared, nil
}

func (r *fakeRepo) GetPreparedTask(_ context.Context, userID int64, _ Target) (PreparedTask, error) {
	r.preparedUserID = userID
	if len(r.preparedQueue) > 0 {
		prepared := r.preparedQueue[0]
		r.preparedQueue = r.preparedQueue[1:]
		return prepared, nil
	}
	if r.prepared == nil {
		return PreparedTask{}, app.NotFound("Готовая задача не найдена")
	}
	return *r.prepared, nil
}

func (r *fakeRepo) GetRandomTaskType(context.Context) (Target, error) {
	return Target{OgeNumber: 10, SubtypeCode: "random_topic"}, nil
}

func (r *fakeRepo) GetPreparationTarget(context.Context, int) (Target, int, error) {
	return Target{OgeNumber: 10, SubtypeCode: "random_topic"}, 0, nil
}

func (r *fakeRepo) ResolveTarget(_ context.Context, target Target) (Target, error) {
	id := int64(7)
	target.TaskTypeID = &id
	return target, nil
}

func (r *fakeRepo) CreateGeneratedTask(_ context.Context, task CreateTask) (Task, error) {
	if len(r.createGeneratedErrs) > 0 {
		err := r.createGeneratedErrs[0]
		r.createGeneratedErrs = r.createGeneratedErrs[1:]
		if err != nil {
			return Task{}, err
		}
	}
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
	reviews       []GenerationReview
	generateErr   error
	reviewErr     error
	generateCalls int
	reviewCalls   int
	feedbacks     []string
}

func (f fakeAI) IsConfigured() bool {
	return f.configured
}

func (f *fakeAI) GenerateTask(_ context.Context, _ Target, feedback string) (GeneratedContent, error) {
	f.generateCalls++
	f.feedbacks = append(f.feedbacks, feedback)
	if f.generateErr != nil {
		return GeneratedContent{}, f.generateErr
	}
	if len(f.generated) >= f.generateCalls {
		return f.generated[f.generateCalls-1], nil
	}
	return validGeneratedContent(nil), nil
}

func (f *fakeAI) ReviewGeneratedTask(context.Context, Target, GeneratedContent) (GenerationReview, error) {
	f.reviewCalls++
	if f.reviewErr != nil {
		return GenerationReview{}, f.reviewErr
	}
	if len(f.reviews) >= f.reviewCalls {
		return f.reviews[f.reviewCalls-1], nil
	}
	return GenerationReview{IsValid: true, CorrectAnswer: "1"}, nil
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

func ptrInt64(value int64) *int64 {
	return &value
}

func validPreparedTask(id int64, question string) PreparedTask {
	return PreparedTask{
		ID:              id,
		Mode:            ModeCustom,
		OgeNumber:       9,
		SubtypeCode:     "linear_equation",
		Question:        question,
		CorrectAnswer:   "3",
		SolutionSteps:   []string{"step 1", "step 2", "step 3"},
		SelfCheck:       "check",
		IsValid:         true,
		ValidationNotes: "ok",
		Source:          "gpt-4o-mini",
	}
}
