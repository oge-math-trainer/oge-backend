package tasks

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/oge-math-trainer/oge-backend.git/internal/app"
)

const (
	ModeWeak       = "weak"
	ModeAll        = "all"
	ModeCustom     = "custom"
	ModeDiagnostic = "diagnostic"
)

type Target struct {
	TaskTypeID  *int64
	OgeNumber   int
	SubtypeCode string
}

// GraphInfo - информация об одном графике от ИИ (для задания №11)
type GraphInfo struct {
	ID           string    `json:"id"`
	Type         string    `json:"type"`
	Coefficients []float64 `json:"coefficients"`
	XMin         int       `json:"x_min"`
	XMax         int       `json:"x_max"`
	YMin         int       `json:"y_min"`
	YMax         int       `json:"y_max"`
}

// GraphTaskParams - контейнер для графиков задачи
type GraphTaskParams struct {
	Graphs []GraphInfo `json:"graphs,omitempty"`
}

type Task struct {
	ID              int64            `json:"id"`
	UserID          int64            `json:"-"`
	Mode            string           `json:"mode"`
	TaskTypeID      *int64           `json:"-"`
	OgeNumber       int              `json:"oge_number"`
	SubtypeCode     string           `json:"subtype_code"`
	Question        string           `json:"question"`
	CorrectAnswer   string           `json:"-"`
	SolutionSteps   []string         `json:"-"`
	SelfCheck       string           `json:"-"`
	IsValid         bool             `json:"-"`
	ValidationNotes string           `json:"-"`
	Source          string           `json:"source"`
	CreatedAt       time.Time        `json:"created_at,omitempty"`
	GraphData       *GraphTaskParams `json:"graphs,omitempty"` // ← Фронтенд ждёт "graphs"
}

type GeneratedContent struct {
	Question        string      `json:"question"`
	CorrectAnswer   string      `json:"correct_answer"`
	SolutionSteps   []string    `json:"solution_steps"`
	SelfCheck       string      `json:"self_check"`
	IsValid         bool        `json:"is_valid"`
	ValidationNotes string      `json:"validation_notes"`
	Graphs          []GraphInfo `json:"graphs,omitempty"`
}

type CreateTask struct {
	UserID          int64
	Mode            string
	TaskTypeID      *int64
	OgeNumber       int
	SubtypeCode     string
	Question        string
	CorrectAnswer   string
	SolutionSteps   []string
	SelfCheck       string
	IsValid         bool
	ValidationNotes string
	Source          string
	GraphData       *GraphTaskParams `json:"graph_data,omitempty"`
}

type FallbackTask struct {
	ID            int64
	TaskTypeID    *int64
	OgeNumber     int
	SubtypeCode   string
	Question      string
	CorrectAnswer string
}

type GenerateRequest struct {
	UserID int64
	Mode   string
	StoredMode  string
	OgeNumber   *int
	SubtypeCode string
}

type CheckRequest struct {
	UserID        int64
	TaskID        int64
	StudentAnswer string
}

type CheckResult struct {
	IsCorrect     bool   `json:"is_correct"`
	ShortFeedback string `json:"short_feedback"`
	Reason        string `json:"reason"`
}

type HintResult struct {
	Hint string `json:"hint"`
}

type ExplainResult struct {
	Explanation string   `json:"explanation"`
	Steps       []string `json:"steps,omitempty"`
}

type Repository interface {
	GetWeakTarget(ctx context.Context, userID int64) (Target, error)
	GetRandomTaskType(ctx context.Context) (Target, error)
	ResolveTarget(ctx context.Context, target Target) (Target, error)
	CreateGeneratedTask(ctx context.Context, task CreateTask) (Task, error)
	GetGeneratedTaskForUser(ctx context.Context, userID, id int64) (Task, error)
	SaveAttempt(ctx context.Context, userID, taskID int64, mode, studentAnswer string, isCorrect bool, aiFeedback any) error
	UpdateProgress(ctx context.Context, userID int64, task Task, isCorrect bool) error
}

type AIClient interface {
	IsConfigured() bool
	GenerateTask(ctx context.Context, target Target) (GeneratedContent, error)
	CheckAnswer(ctx context.Context, task Task, studentAnswer string) (CheckResult, error)
	Hint(ctx context.Context, task Task) (HintResult, error)
	Explain(ctx context.Context, task Task) (ExplainResult, error)
}

type Service struct {
	repo Repository
	ai   AIClient
}

func NewService(repo Repository, ai AIClient) *Service {
	return &Service{repo: repo, ai: ai}
}

func (s *Service) Generate(ctx context.Context, req GenerateRequest) (Task, error) {
	target, err := s.resolveTarget(ctx, req)
	if err != nil {
		return Task{}, err
	}

	if s.ai == nil || !s.ai.IsConfigured() {
		return Task{}, app.AIUnavailable(errors.New("AITUNNEL_API_KEY is empty"))
	}

	content, err := s.ai.GenerateTask(ctx, target)
	if err != nil {
		return Task{}, app.AIUnavailable(err)
	}
	if !content.IsValid {
		return Task{}, app.AIUnavailable(errors.New("ai returned invalid task"))
	}

	// Упаковываем графики от ИИ
	var graphData *GraphTaskParams
	if len(content.Graphs) > 0 {
		graphData = &GraphTaskParams{
			Graphs: content.Graphs,
		}
	}

	return s.repo.CreateGeneratedTask(ctx, CreateTask{
		UserID:          req.UserID,
		Mode:            storedMode(req),
		TaskTypeID:      target.TaskTypeID,
		OgeNumber:       target.OgeNumber,
		SubtypeCode:     target.SubtypeCode,
		Question:        strings.TrimSpace(content.Question),
		CorrectAnswer:   strings.TrimSpace(content.CorrectAnswer),
		SolutionSteps:   content.SolutionSteps,
		SelfCheck:       strings.TrimSpace(content.SelfCheck),
		IsValid:         content.IsValid,
		ValidationNotes: strings.TrimSpace(content.ValidationNotes),
		Source:          "qwen",
		GraphData:       graphData,
	})
}

func (s *Service) Check(ctx context.Context, req CheckRequest) (CheckResult, error) {
	task, err := s.repo.GetGeneratedTaskForUser(ctx, req.UserID, req.TaskID)
	if err != nil {
		return CheckResult{}, err
	}
	if s.ai == nil || !s.ai.IsConfigured() {
		return CheckResult{}, app.AIUnavailable(errors.New("AITUNNEL_API_KEY is empty"))
	}
	return s.ai.CheckAnswer(ctx, task, req.StudentAnswer)
}

func (s *Service) Hint(ctx context.Context, userID, taskID int64) (HintResult, error) {
	task, err := s.repo.GetGeneratedTaskForUser(ctx, userID, taskID)
	if err != nil {
		return HintResult{}, err
	}
	if s.ai == nil || !s.ai.IsConfigured() {
		return HintResult{}, app.AIUnavailable(errors.New("AITUNNEL_API_KEY is empty"))
	}
	return s.ai.Hint(ctx, task)
}

func (s *Service) Explain(ctx context.Context, userID, taskID int64) (ExplainResult, error) {
	task, err := s.repo.GetGeneratedTaskForUser(ctx, userID, taskID)
	if err != nil {
		return ExplainResult{}, err
	}
	if s.ai == nil || !s.ai.IsConfigured() {
		return ExplainResult{}, app.AIUnavailable(errors.New("AITUNNEL_API_KEY is empty"))
	}
	return s.ai.Explain(ctx, task)
}

func (s *Service) resolveTarget(ctx context.Context, req GenerateRequest) (Target, error) {
	switch req.Mode {
	case ModeCustom:
		if req.OgeNumber == nil || strings.TrimSpace(req.SubtypeCode) == "" {
			return Target{}, app.Validation("Для режима custom нужны oge_number и subtype_code")
		}
		return s.repo.ResolveTarget(ctx, Target{OgeNumber: *req.OgeNumber, SubtypeCode: strings.TrimSpace(req.SubtypeCode)})
	case ModeWeak:
		return s.repo.GetWeakTarget(ctx, req.UserID)
	case ModeAll:
		return s.repo.GetRandomTaskType(ctx)
	default:
		return Target{}, app.Validation("Некорректный режим тренировки")
	}
}

func storedMode(req GenerateRequest) string {
	if strings.TrimSpace(req.StoredMode) != "" {
		return strings.TrimSpace(req.StoredMode)
	}
	return req.Mode
}
