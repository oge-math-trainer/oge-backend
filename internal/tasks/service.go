package tasks

import (
	"context"
	"errors"
	"fmt"
	"log"
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

type VisualData map[string]any

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
	VisualData      VisualData       `json:"visual_data,omitempty"`
	GraphData       *GraphTaskParams `json:"graphs,omitempty"` // ← Фронтенд ждёт "graphs"
}

type GeneratedContent struct {
	Question        string      `json:"question"`
	CorrectAnswer   string      `json:"correct_answer"`
	SolutionSteps   []string    `json:"solution_steps"`
	SelfCheck       string      `json:"self_check"`
	IsValid         bool        `json:"is_valid"`
	ValidationNotes string      `json:"validation_notes"`
	VisualData      VisualData  `json:"visual_data,omitempty"`
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
	VisualData      VisualData
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
	UserID      int64
	Mode        string
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

	content, source, err := s.generateContent(ctx, target)
	if err != nil {
		return Task{}, err
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
		Source:          source,
		VisualData:      content.VisualData,
	})
}

func (s *Service) generateContent(ctx context.Context, target Target) (GeneratedContent, string, error) {
	if s.ai == nil || !s.ai.IsConfigured() {
		if RequiresVisualData(target) {
			err := errors.New("AI client is not configured")
			log.Printf("ai generation fallback: oge_number=%d subtype_code=%s reason=%v", target.OgeNumber, target.SubtypeCode, err)
			return FallbackGeneratedContent(target, err), "fallback", nil
		}
		return GeneratedContent{}, "", app.AIUnavailable(errors.New("AITUNNEL_API_KEY is empty"))
	}

	var lastErr error
	const maxAttempts = 2
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		started := time.Now()
		content, err := s.ai.GenerateTask(ctx, target)
		duration := time.Since(started)
		if err != nil {
			lastErr = err
			log.Printf("ai generation result: oge_number=%d subtype_code=%s attempt=%d duration_ms=%d validation=error error=%v",
				target.OgeNumber, target.SubtypeCode, attempt, duration.Milliseconds(), err)
			continue
		}

		if err := validateGeneratedContent(target, &content); err != nil {
			lastErr = err
			log.Printf("ai generation result: oge_number=%d subtype_code=%s attempt=%d duration_ms=%d validation=error error=%v",
				target.OgeNumber, target.SubtypeCode, attempt, duration.Milliseconds(), err)
			continue
		}

		log.Printf("ai generation result: oge_number=%d subtype_code=%s attempt=%d duration_ms=%d validation=success",
			target.OgeNumber, target.SubtypeCode, attempt, duration.Milliseconds())
		return content, "qwen", nil
	}

	if RequiresVisualData(target) {
		log.Printf("ai generation fallback: oge_number=%d subtype_code=%s reason=%v", target.OgeNumber, target.SubtypeCode, lastErr)
		return FallbackGeneratedContent(target, lastErr), "fallback", nil
	}

	if lastErr == nil {
		lastErr = errors.New("ai returned no generated content")
	}
	return GeneratedContent{}, "", app.AIUnavailable(lastErr)
}

func validateGeneratedContent(target Target, content *GeneratedContent) error {
	if content == nil {
		return errors.New("AI returned empty content")
	}
	if strings.TrimSpace(content.Question) == "" {
		return errors.New("AI returned empty question")
	}
	if strings.TrimSpace(content.CorrectAnswer) == "" {
		return errors.New("AI returned empty correct_answer")
	}
	if len(content.SolutionSteps) < 3 {
		return fmt.Errorf("AI returned too few solution_steps: got %d, need at least 3", len(content.SolutionSteps))
	}
	if !content.IsValid {
		return errors.New("AI returned is_valid=false")
	}
	if !RequiresVisualData(target) {
		content.VisualData = nil
		return nil
	}

	visualData, err := NormalizeAndValidateVisualData(target, content)
	if err != nil {
		return err
	}
	content.VisualData = visualData
	return nil
}

func (s *Service) Check(ctx context.Context, req CheckRequest) (CheckResult, error) {
	task, err := s.repo.GetGeneratedTaskForUser(ctx, req.UserID, req.TaskID)
	if err != nil {
		return CheckResult{}, err
	}
	if s.ai == nil || !s.ai.IsConfigured() {
		return CheckResult{}, app.AIUnavailable(errors.New("AITUNNEL_API_KEY is empty"))
	}
	check, err := s.ai.CheckAnswer(ctx, task, req.StudentAnswer)
	if err != nil {
		return CheckResult{}, err
	}
	if err := s.repo.SaveAttempt(ctx, req.UserID, task.ID, task.Mode, req.StudentAnswer, check.IsCorrect, check); err != nil {
		return CheckResult{}, err
	}
	if err := s.repo.UpdateProgress(ctx, req.UserID, task, check.IsCorrect); err != nil {
		return CheckResult{}, err
	}
	return check, nil
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
