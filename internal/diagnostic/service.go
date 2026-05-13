package diagnostic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/oge-math-trainer/oge-backend.git/internal/app"
	"github.com/oge-math-trainer/oge-backend.git/internal/tasks"
)

type StartResult struct {
	SessionID      int64        `json:"session_id"`
	Tasks          []tasks.Task `json:"tasks"`
	TotalTasks     int          `json:"total_tasks"`
	GeneratedCount int          `json:"generated_count"`
	Complete       bool         `json:"complete"`
}

type NextResult struct {
	SessionID      int64       `json:"session_id"`
	Task           *tasks.Task `json:"task,omitempty"`
	TotalTasks     int         `json:"total_tasks"`
	GeneratedCount int         `json:"generated_count"`
	Complete       bool        `json:"complete"`
}

type AnswerInput struct {
	TaskID        int64  `json:"task_id"`
	StudentAnswer string `json:"student_answer"`
}

type AnswerResult struct {
	TaskID        int64  `json:"task_id"`
	OgeNumber     int    `json:"oge_number"`
	SubtypeCode   string `json:"subtype_code"`
	StudentAnswer string `json:"student_answer"`
	IsCorrect     bool   `json:"is_correct"`
	Feedback      string `json:"feedback"`
}

type Analysis struct {
	Summary    string   `json:"summary"`
	WeakTopics []string `json:"weak_topics"`
}

type SubmitResult struct {
	SessionID int64          `json:"session_id"`
	Answers   []AnswerResult `json:"answers"`
	Analysis  Analysis       `json:"analysis"`
}

type Repository interface {
	CreateDiagnosticSession(ctx context.Context, userID int64) (int64, error)
	EnsureDiagnosticSessionOwner(ctx context.Context, userID, sessionID int64) error
	GetDiagnosticTargets(ctx context.Context) ([]tasks.Target, error)

	AttachTaskToSession(ctx context.Context, sessionID, taskID int64) error
	GetTasksBySession(ctx context.Context, sessionID int64) ([]tasks.Task, error)
	GetTaskBySession(ctx context.Context, sessionID, taskID int64) (tasks.Task, error)

	SaveDiagnosticAnswer(ctx context.Context, sessionID int64, task tasks.Task, studentAnswer string, isCorrect bool, feedback any) error
	FinishDiagnosticSession(ctx context.Context, sessionID int64, analysis Analysis, weakTopics []string) error
	UpdateProgress(ctx context.Context, userID int64, task tasks.Task, isCorrect bool) error
}

type TaskGenerator interface {
	Generate(ctx context.Context, req tasks.GenerateRequest) (tasks.Task, error)
}

type AIClient interface {
	IsConfigured() bool
	CheckAnswer(ctx context.Context, task tasks.Task, studentAnswer string) (tasks.CheckResult, error)
	AnalyzeDiagnostic(ctx context.Context, answers []AnswerResult) (Analysis, error)
}

type Service struct {
	repo  Repository
	tasks TaskGenerator
	ai    AIClient
}

func NewService(repo Repository, taskGenerator TaskGenerator, ai AIClient) *Service {
	return &Service{repo: repo, tasks: taskGenerator, ai: ai}
}

func (s *Service) Start(ctx context.Context, userID int64) (StartResult, error) {
	targets, err := s.repo.GetDiagnosticTargets(ctx)
	if err != nil {
		return StartResult{}, err
	}

	if len(targets) < 14 {
		return StartResult{}, app.NotFound("Не найдены 14 типов заданий для диагностики")
	}

	sessionID, err := s.repo.CreateDiagnosticSession(ctx, userID)
	if err != nil {
		return StartResult{}, err
	}

	task, generatedCount, complete, err := s.generateNextTask(ctx, userID, sessionID, targets[:14], nil)
	if err != nil {
		return StartResult{}, err
	}

	return StartResult{
		SessionID:      sessionID,
		Tasks:          []tasks.Task{task},
		TotalTasks:     len(targets[:14]),
		GeneratedCount: generatedCount,
		Complete:       complete,
	}, nil
}

func (s *Service) Next(ctx context.Context, userID, sessionID int64) (NextResult, error) {
	if err := s.repo.EnsureDiagnosticSessionOwner(ctx, userID, sessionID); err != nil {
		return NextResult{}, err
	}

	targets, err := s.repo.GetDiagnosticTargets(ctx)
	if err != nil {
		return NextResult{}, err
	}
	if len(targets) < 14 {
		return NextResult{}, app.NotFound("Не найдены 14 типов заданий для диагностики")
	}

	existingTasks, err := s.repo.GetTasksBySession(ctx, sessionID)
	if err != nil {
		return NextResult{}, err
	}

	task, generatedCount, complete, err := s.generateNextTask(ctx, userID, sessionID, targets[:14], existingTasks)
	if err != nil {
		return NextResult{}, err
	}

	result := NextResult{
		SessionID:      sessionID,
		TotalTasks:     len(targets[:14]),
		GeneratedCount: generatedCount,
		Complete:       complete,
	}
	if !complete {
		result.Task = &task
	}
	return result, nil
}

func (s *Service) generateNextTask(
	ctx context.Context,
	userID, sessionID int64,
	targets []tasks.Target,
	existingTasks []tasks.Task,
) (tasks.Task, int, bool, error) {
	generatedOgeNumbers := make(map[int]struct{}, len(existingTasks))
	for _, task := range existingTasks {
		generatedOgeNumbers[task.OgeNumber] = struct{}{}
	}
	if len(generatedOgeNumbers) >= len(targets) {
		return tasks.Task{}, len(existingTasks), true, nil
	}

	var lastErr error
	for _, target := range targets {
		if _, exists := generatedOgeNumbers[target.OgeNumber]; exists {
			continue
		}

		task, err := s.generateTaskForTarget(ctx, userID, target)
		if err != nil {
			lastErr = err
			continue
		}

		if err := s.repo.AttachTaskToSession(ctx, sessionID, task.ID); err != nil {
			return tasks.Task{}, len(existingTasks), false, err
		}

		generatedCount := len(existingTasks) + 1
		return task, generatedCount, generatedCount >= len(targets), nil
	}

	if lastErr != nil {
		return tasks.Task{}, len(existingTasks), false, lastErr
	}
	return tasks.Task{}, len(existingTasks), true, nil
}

func (s *Service) generateTaskForTarget(ctx context.Context, userID int64, target tasks.Target) (tasks.Task, error) {
	oge := target.OgeNumber

	var task tasks.Task
	var lastErr error
	const maxRetries = 3

	for attempt := 1; attempt <= maxRetries; attempt++ {
		var err error
		task, err = s.tasks.Generate(ctx, tasks.GenerateRequest{
			UserID:      userID,
			Mode:        tasks.ModeCustom,
			StoredMode:  tasks.ModeDiagnostic,
			OgeNumber:   &oge,
			SubtypeCode: target.SubtypeCode,
		})
		if err == nil {
			return task, nil
		}

		lastErr = err
		fmt.Printf("diag generate fail OGE#%d attempt %d/%d: %v\n",
			target.OgeNumber, attempt, maxRetries, err)

		select {
		case <-ctx.Done():
			return tasks.Task{}, ctx.Err()
		case <-time.After(300 * time.Millisecond):
		}
	}

	if lastErr != nil {
		return tasks.Task{}, lastErr
	}
	return tasks.Task{}, app.AIUnavailable(errors.New("no diagnostic task generated"))
}

func (s *Service) Submit(
	ctx context.Context,
	userID, sessionID int64,
	answers []AnswerInput,
) (SubmitResult, error) {

	if err := s.repo.EnsureDiagnosticSessionOwner(ctx, userID, sessionID); err != nil {
		return SubmitResult{}, err
	}

	aiEnabled := s.ai != nil && s.ai.IsConfigured()

	results := make([]AnswerResult, 0, len(answers))

	for _, answer := range answers {

		task, err := s.repo.GetTaskBySession(ctx, sessionID, answer.TaskID)
		if err != nil {
			return SubmitResult{}, err
		}

		var (
			isCorrect bool
			feedback  string
		)

		if aiEnabled {
			check, err := s.ai.CheckAnswer(ctx, task, answer.StudentAnswer)
			if err != nil {
				return SubmitResult{}, err
			}

			isCorrect = check.IsCorrect
			feedback = check.ShortFeedback
		} else {
			isCorrect = false
			feedback = "AI недоступен"
		}

		if err := s.repo.SaveDiagnosticAnswer(
			ctx,
			sessionID,
			task,
			answer.StudentAnswer,
			isCorrect,
			nil,
		); err != nil {
			return SubmitResult{}, err
		}

		if err := s.repo.UpdateProgress(ctx, userID, task, isCorrect); err != nil {
			return SubmitResult{}, err
		}

		results = append(results, AnswerResult{
			TaskID:        task.ID,
			OgeNumber:     task.OgeNumber,
			SubtypeCode:   task.SubtypeCode,
			StudentAnswer: answer.StudentAnswer,
			IsCorrect:     isCorrect,
			Feedback:      feedback,
		})
	}

	var analysis Analysis
	var err error

	if len(results) == 0 {
		analysis = Analysis{
			Summary: "Диагностика завершена без ответов",
		}
	} else if aiEnabled {
		analysis, err = s.ai.AnalyzeDiagnostic(ctx, results)
		if err != nil {
			return SubmitResult{}, err
		}
	} else {
		analysis = Analysis{
			Summary: "AI выключен",
		}
	}

	if err := s.repo.FinishDiagnosticSession(ctx, sessionID, analysis, analysis.WeakTopics); err != nil {
		return SubmitResult{}, err
	}

	return SubmitResult{
		SessionID: sessionID,
		Answers:   results,
		Analysis:  analysis,
	}, nil
}
