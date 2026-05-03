package diagnostic

import (
	"context"
	"errors"

	"github.com/oge-math-trainer/oge-backend.git/internal/app"
	"github.com/oge-math-trainer/oge-backend.git/internal/tasks"
)

type StartResult struct {
	SessionID int64        `json:"session_id"`
	Tasks     []tasks.Task `json:"tasks"`
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
	GetGeneratedTaskForUser(ctx context.Context, userID, id int64) (tasks.Task, error)
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

	out := make([]tasks.Task, 0, 14)
	for _, target := range targets[:14] {
		oge := target.OgeNumber
		task, err := s.tasks.Generate(ctx, tasks.GenerateRequest{
			UserID:      userID,
			Mode:        tasks.ModeCustom,
			StoredMode:  tasks.ModeDiagnostic,
			OgeNumber:   &oge,
			SubtypeCode: target.SubtypeCode,
		})
		if err != nil {
			return StartResult{}, err
		}
		out = append(out, task)
	}

	sessionID, err := s.repo.CreateDiagnosticSession(ctx, userID)
	if err != nil {
		return StartResult{}, err
	}

	return StartResult{SessionID: sessionID, Tasks: out}, nil
}

func (s *Service) Submit(ctx context.Context, userID, sessionID int64, answers []AnswerInput) (SubmitResult, error) {
	if err := s.repo.EnsureDiagnosticSessionOwner(ctx, userID, sessionID); err != nil {
		return SubmitResult{}, err
	}

	if s.ai == nil || !s.ai.IsConfigured() {
		return SubmitResult{}, app.AIUnavailable(errors.New("OPENROUTER_API_KEY is empty"))
	}

	results := make([]AnswerResult, 0, len(answers))
	for _, answer := range answers {
		task, err := s.repo.GetGeneratedTaskForUser(ctx, userID, answer.TaskID)
		if err != nil {
			return SubmitResult{}, err
		}
		check, err := s.ai.CheckAnswer(ctx, task, answer.StudentAnswer)
		if err != nil {
			return SubmitResult{}, err
		}
		if err := s.repo.SaveDiagnosticAnswer(ctx, sessionID, task, answer.StudentAnswer, check.IsCorrect, check); err != nil {
			return SubmitResult{}, err
		}
		if err := s.repo.UpdateProgress(ctx, userID, task, check.IsCorrect); err != nil {
			return SubmitResult{}, err
		}
		results = append(results, AnswerResult{
			TaskID:        task.ID,
			OgeNumber:     task.OgeNumber,
			SubtypeCode:   task.SubtypeCode,
			StudentAnswer: answer.StudentAnswer,
			IsCorrect:     check.IsCorrect,
			Feedback:      check.ShortFeedback,
		})
	}

	analysis, err := s.ai.AnalyzeDiagnostic(ctx, results)
	if err != nil {
		return SubmitResult{}, err
	}
	if err := s.repo.FinishDiagnosticSession(ctx, sessionID, analysis, analysis.WeakTopics); err != nil {
		return SubmitResult{}, err
	}

	return SubmitResult{SessionID: sessionID, Answers: results, Analysis: analysis}, nil
}
