package diagnostic

import (
	"context"
	"errors"
	"testing"

	"github.com/oge-math-trainer/oge-backend.git/internal/app"
	"github.com/oge-math-trainer/oge-backend.git/internal/tasks"
)

func TestSubmitChecksSessionOwnerBeforeAI(t *testing.T) {
	repo := &fakeRepo{ownerErr: app.NotFound("session not found")}
	ai := &fakeAI{configured: true}
	service := NewService(repo, &fakeTaskGenerator{}, ai)

	_, err := service.Submit(context.Background(), 1, 99, []AnswerInput{
		{TaskID: 1, StudentAnswer: "3"},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	var appErr *app.Error
	if !errors.As(err, &appErr) || appErr.Code != app.CodeNotFound {
		t.Fatalf("expected not_found before ai call, got %v", err)
	}
	if !repo.ownerChecked {
		t.Fatal("expected owner check")
	}
	if repo.taskBySessionChecked {
		t.Fatal("did not expect task lookup before owner check passed")
	}
	if ai.checkCalls != 0 {
		t.Fatalf("expected no AI check calls, got %d", ai.checkCalls)
	}
}

func TestStartAttachesFirstGeneratedTaskToSession(t *testing.T) {
	repo := &fakeRepo{
		sessionID: 42,
		targets:   diagnosticTargets(14),
	}
	service := NewService(repo, &fakeTaskGenerator{}, &fakeAI{configured: false})

	result, err := service.Start(context.Background(), 9)
	if err != nil {
		t.Fatalf("Start returned error: %v", err)
	}
	if result.SessionID != 42 {
		t.Fatalf("expected session id 42, got %d", result.SessionID)
	}
	if result.TotalTasks != 14 {
		t.Fatalf("expected total tasks 14, got %d", result.TotalTasks)
	}
	if result.GeneratedCount != 1 {
		t.Fatalf("expected generated count 1, got %d", result.GeneratedCount)
	}
	if len(result.Tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(result.Tasks))
	}
	if len(repo.attachedTaskIDs) != 1 {
		t.Fatalf("expected 1 attached task, got %d", len(repo.attachedTaskIDs))
	}
	if repo.attachedSessionIDs[0] != result.SessionID {
		t.Fatalf("attached task to session %d, expected %d", repo.attachedSessionIDs[0], result.SessionID)
	}
	if repo.attachedTaskIDs[0] != result.Tasks[0].ID {
		t.Fatalf("attached task id %d, expected %d", repo.attachedTaskIDs[0], result.Tasks[0].ID)
	}
}

func TestNextGeneratesOneTaskAfterExistingSessionTasks(t *testing.T) {
	repo := &fakeRepo{
		targets: diagnosticTargets(14),
		sessionTasks: []tasks.Task{
			{ID: 101, UserID: 9, Mode: tasks.ModeDiagnostic, OgeNumber: 6, SubtypeCode: "target"},
		},
	}
	service := NewService(repo, &fakeTaskGenerator{}, &fakeAI{configured: false})

	result, err := service.Next(context.Background(), 9, 77)
	if err != nil {
		t.Fatalf("Next returned error: %v", err)
	}
	if result.Task == nil {
		t.Fatal("expected generated task")
	}
	if result.Task.OgeNumber != 7 {
		t.Fatalf("expected next OGE number 7, got %d", result.Task.OgeNumber)
	}
	if result.GeneratedCount != 2 {
		t.Fatalf("expected generated count 2, got %d", result.GeneratedCount)
	}
	if len(repo.attachedTaskIDs) != 1 || repo.attachedSessionIDs[0] != 77 {
		t.Fatalf("expected one attached task for session 77, got sessions=%v tasks=%v", repo.attachedSessionIDs, repo.attachedTaskIDs)
	}
}

func TestSubmitAllowsEmptyAnswers(t *testing.T) {
	repo := &fakeRepo{}
	ai := &fakeAI{configured: true}
	service := NewService(repo, &fakeTaskGenerator{}, ai)

	result, err := service.Submit(context.Background(), 9, 77, nil)
	if err != nil {
		t.Fatalf("Submit returned error: %v", err)
	}
	if len(result.Answers) != 0 {
		t.Fatalf("expected no answers, got %#v", result.Answers)
	}
	if !repo.finished {
		t.Fatal("expected session finish")
	}
	if ai.analyzeCalls != 0 {
		t.Fatalf("expected no AI analysis for empty answers, got %d", ai.analyzeCalls)
	}
}

func TestSubmitLooksUpTaskBySession(t *testing.T) {
	repo := &fakeRepo{
		taskBySession: tasks.Task{ID: 434, UserID: 9, Mode: tasks.ModeDiagnostic, OgeNumber: 9, SubtypeCode: "linear_equation"},
	}
	service := NewService(repo, &fakeTaskGenerator{}, &fakeAI{configured: false})

	result, err := service.Submit(context.Background(), 9, 77, []AnswerInput{
		{TaskID: 434, StudentAnswer: "3"},
	})
	if err != nil {
		t.Fatalf("Submit returned error: %v", err)
	}
	if !repo.taskBySessionChecked {
		t.Fatal("expected task lookup by session")
	}
	if repo.taskBySessionSessionID != 77 || repo.taskBySessionTaskID != 434 {
		t.Fatalf("unexpected lookup session=%d task=%d", repo.taskBySessionSessionID, repo.taskBySessionTaskID)
	}
	if len(result.Answers) != 1 || result.Answers[0].TaskID != 434 {
		t.Fatalf("unexpected answers: %#v", result.Answers)
	}
	if repo.savedAnswers != 1 {
		t.Fatalf("expected saved answer, got %d", repo.savedAnswers)
	}
	if repo.progressUpdates != 1 {
		t.Fatalf("expected progress update, got %d", repo.progressUpdates)
	}
	if !repo.finished {
		t.Fatal("expected session finish")
	}
}

func TestSubmitRejectsTaskOutsideSessionBeforeAI(t *testing.T) {
	repo := &fakeRepo{taskBySessionErr: app.NotFound("task not found in session")}
	ai := &fakeAI{configured: true}
	service := NewService(repo, &fakeTaskGenerator{}, ai)

	_, err := service.Submit(context.Background(), 9, 77, []AnswerInput{
		{TaskID: 434, StudentAnswer: "3"},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	var appErr *app.Error
	if !errors.As(err, &appErr) || appErr.Code != app.CodeNotFound {
		t.Fatalf("expected not_found, got %v", err)
	}
	if !repo.taskBySessionChecked {
		t.Fatal("expected task lookup by session")
	}
	if ai.checkCalls != 0 {
		t.Fatalf("expected no AI check calls, got %d", ai.checkCalls)
	}
	if repo.savedAnswers != 0 {
		t.Fatalf("expected no saved answers, got %d", repo.savedAnswers)
	}
	if repo.progressUpdates != 0 {
		t.Fatalf("expected no progress updates, got %d", repo.progressUpdates)
	}
	if repo.finished {
		t.Fatal("did not expect session finish")
	}
}

type fakeRepo struct {
	ownerErr     error
	ownerChecked bool

	sessionID int64
	targets   []tasks.Target

	attachedSessionIDs []int64
	attachedTaskIDs    []int64
	attachErr          error
	sessionTasks       []tasks.Task

	taskBySession          tasks.Task
	taskBySessionErr       error
	taskBySessionChecked   bool
	taskBySessionSessionID int64
	taskBySessionTaskID    int64

	savedAnswers    int
	progressUpdates int
	finished        bool
}

func (r *fakeRepo) CreateDiagnosticSession(context.Context, int64) (int64, error) {
	if r.sessionID != 0 {
		return r.sessionID, nil
	}
	return 1, nil
}

func (r *fakeRepo) EnsureDiagnosticSessionOwner(context.Context, int64, int64) error {
	r.ownerChecked = true
	return r.ownerErr
}

func (r *fakeRepo) GetDiagnosticTargets(context.Context) ([]tasks.Target, error) {
	return r.targets, nil
}

func (r *fakeRepo) AttachTaskToSession(_ context.Context, sessionID, taskID int64) error {
	if r.attachErr != nil {
		return r.attachErr
	}
	r.attachedSessionIDs = append(r.attachedSessionIDs, sessionID)
	r.attachedTaskIDs = append(r.attachedTaskIDs, taskID)
	return nil
}

func (r *fakeRepo) GetTasksBySession(context.Context, int64) ([]tasks.Task, error) {
	return r.sessionTasks, nil
}

func (r *fakeRepo) GetTaskBySession(_ context.Context, sessionID, taskID int64) (tasks.Task, error) {
	r.taskBySessionChecked = true
	r.taskBySessionSessionID = sessionID
	r.taskBySessionTaskID = taskID
	if r.taskBySessionErr != nil {
		return tasks.Task{}, r.taskBySessionErr
	}
	if r.taskBySession.ID != 0 {
		return r.taskBySession, nil
	}
	return tasks.Task{ID: taskID, UserID: 1, Mode: tasks.ModeDiagnostic, OgeNumber: 9, SubtypeCode: "linear_equation"}, nil
}

func (r *fakeRepo) SaveDiagnosticAnswer(context.Context, int64, tasks.Task, string, bool, any) error {
	r.savedAnswers++
	return nil
}

func (r *fakeRepo) FinishDiagnosticSession(context.Context, int64, Analysis, []string) error {
	r.finished = true
	return nil
}

func (r *fakeRepo) UpdateProgress(context.Context, int64, tasks.Task, bool) error {
	r.progressUpdates++
	return nil
}

type fakeTaskGenerator struct {
	generated []tasks.Task
	calls     int
}

func (g *fakeTaskGenerator) Generate(_ context.Context, req tasks.GenerateRequest) (tasks.Task, error) {
	if g.calls < len(g.generated) {
		task := g.generated[g.calls]
		g.calls++
		return task, nil
	}
	g.calls++
	return tasks.Task{
		ID:          int64(100 + g.calls),
		UserID:      req.UserID,
		Mode:        tasks.ModeDiagnostic,
		OgeNumber:   valueOrZero(req.OgeNumber),
		SubtypeCode: req.SubtypeCode,
	}, nil
}

type fakeAI struct {
	configured   bool
	checkCalls   int
	analyzeCalls int
}

func (f *fakeAI) IsConfigured() bool {
	return f.configured
}

func (f *fakeAI) CheckAnswer(context.Context, tasks.Task, string) (tasks.CheckResult, error) {
	f.checkCalls++
	return tasks.CheckResult{IsCorrect: true, ShortFeedback: "ok"}, nil
}

func (f *fakeAI) AnalyzeDiagnostic(context.Context, []AnswerResult) (Analysis, error) {
	f.analyzeCalls++
	return Analysis{Summary: "ok"}, nil
}

func diagnosticTargets(count int) []tasks.Target {
	targets := make([]tasks.Target, 0, count)
	for i := 0; i < count; i++ {
		targets = append(targets, tasks.Target{
			OgeNumber:   6 + i%14,
			SubtypeCode: "target",
		})
	}
	return targets
}

func valueOrZero(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}
