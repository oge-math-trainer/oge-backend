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
// type GraphTaskParams struct {
// 	Graphs []GraphInfo `json:"graphs,omitempty"`
// }

type VisualData map[string]any

type Task struct {
	ID              int64       `json:"id"`
	UserID          int64       `json:"-"`
	Mode            string      `json:"mode"`
	TaskTypeID      *int64      `json:"-"`
	OgeNumber       int         `json:"oge_number"`
	SubtypeCode     string      `json:"subtype_code"`
	Question        string      `json:"question"`
	CorrectAnswer   string      `json:"-"`
	SolutionSteps   []string    `json:"-"`
	SelfCheck       string      `json:"-"`
	IsValid         bool        `json:"-"`
	ValidationNotes string      `json:"-"`
	Source          string      `json:"source"`
	CreatedAt       time.Time   `json:"created_at,omitempty"`
	VisualData      VisualData  `json:"visual_data,omitempty"`
	Graphs          []GraphInfo `json:"graphs,omitempty"` // ← Фронтенд ждёт "graphs"
}

type PreparedTask struct {
	ID              int64       `json:"id"`
	Mode            string      `json:"mode"`
	TaskTypeID      *int64      `json:"-"`
	OgeNumber       int         `json:"oge_number"`
	SubtypeCode     string      `json:"subtype_code"`
	Question        string      `json:"question"`
	CorrectAnswer   string      `json:"-"`
	SolutionSteps   []string    `json:"-"`
	SelfCheck       string      `json:"-"`
	IsValid         bool        `json:"-"`
	ValidationNotes string      `json:"-"`
	Source          string      `json:"source"`
	CreatedAt       time.Time   `json:"created_at,omitempty"`
	VisualData      VisualData  `json:"visual_data,omitempty"`
	Graphs          []GraphInfo `json:"graphs,omitempty"`
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

type GenerationReview struct {
	IsValid       bool   `json:"is_valid"`
	Reason        string `json:"reason"`
	CorrectAnswer string `json:"correct_answer"`
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
	Graphs          []GraphInfo
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
	TryAcquirePreparationLock(ctx context.Context) (func(), bool, error)
	GetWeakTarget(ctx context.Context, userID int64) (Target, error)
	GetRandomTaskType(ctx context.Context) (Target, error)
	GetPreparationTarget(ctx context.Context, minReady int) (Target, int, error)
	ResolveTarget(ctx context.Context, target Target) (Target, error)
	GetPreparedTask(ctx context.Context, userID int64, target Target) (PreparedTask, error)
	CountPreparedTasks(ctx context.Context) (int, error)
	CreatePreparedTask(ctx context.Context, task CreateTask) (PreparedTask, error)
	CreateGeneratedTask(ctx context.Context, task CreateTask) (Task, error)
	GetGeneratedTaskForUser(ctx context.Context, userID, id int64) (Task, error)
	SaveAttempt(ctx context.Context, userID, taskID int64, mode, studentAnswer string, isCorrect bool, aiFeedback any) error
	UpdateProgress(ctx context.Context, userID int64, task Task, isCorrect bool) error
}

type AIClient interface {
	IsConfigured() bool
	GenerateTask(ctx context.Context, target Target, feedback string) (GeneratedContent, error)
	ReviewGeneratedTask(ctx context.Context, target Target, content GeneratedContent) (GenerationReview, error)
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

	const maxPreparedAttempts = 5
	for attempt := 1; attempt <= maxPreparedAttempts; attempt++ {
		preparedTask, err := s.repo.GetPreparedTask(ctx, req.UserID, target)
		if err != nil {
			var appErr *app.Error
			if errors.As(err, &appErr) && appErr.Code == app.CodeNotFound {
				log.Printf("task generate cache miss: user_id=%d mode=%s target=%d/%s",
					req.UserID, req.Mode, target.OgeNumber, target.SubtypeCode)
				return s.generateOnDemand(ctx, req, target)
			}
			log.Printf("task generate cache error: user_id=%d mode=%s target=%d/%s error=%v",
				req.UserID, req.Mode, target.OgeNumber, target.SubtypeCode, err)
			return Task{}, err
		}

		log.Printf("task generate cache hit: user_id=%d mode=%s target=%d/%s prepared_id=%d",
			req.UserID, req.Mode, target.OgeNumber, target.SubtypeCode, preparedTask.ID)
		content := GeneratedContent{
			Question:        preparedTask.Question,
			CorrectAnswer:   preparedTask.CorrectAnswer,
			SolutionSteps:   preparedTask.SolutionSteps,
			SelfCheck:       preparedTask.SelfCheck,
			IsValid:         preparedTask.IsValid,
			ValidationNotes: preparedTask.ValidationNotes,
			VisualData:      preparedTask.VisualData,
			Graphs:          preparedTask.Graphs,
		}
		preparedTarget := Target{
			TaskTypeID:  preparedTask.TaskTypeID,
			OgeNumber:   preparedTask.OgeNumber,
			SubtypeCode: preparedTask.SubtypeCode,
		}
		if err := validateGeneratedContent(preparedTarget, &content); err != nil {
			log.Printf("task generate cache item rejected: user_id=%d mode=%s target=%d/%s prepared_id=%d error=%v",
				req.UserID, req.Mode, target.OgeNumber, target.SubtypeCode, preparedTask.ID, err)
			return Task{}, app.TaskUnavailable("Готовые уникальные задачи для этой темы еще готовятся")
		}

		task, err := s.repo.CreateGeneratedTask(ctx, createGeneratedTask(req.UserID, storedMode(req), preparedTarget, content, preparedTask.Source))
		if err == nil {
			return task, nil
		}
		if isAppConflict(err) {
			log.Printf("task generate duplicate skipped: user_id=%d mode=%s target=%d/%s prepared_id=%d attempt=%d/%d",
				req.UserID, req.Mode, target.OgeNumber, target.SubtypeCode, preparedTask.ID, attempt, maxPreparedAttempts)
			continue
		}
		return Task{}, err
	}

	log.Printf("task generate cache exhausted by duplicates: user_id=%d mode=%s target=%d/%s attempts=%d",
		req.UserID, req.Mode, target.OgeNumber, target.SubtypeCode, maxPreparedAttempts)
	return Task{}, app.TaskUnavailable("Готовые уникальные задачи для этой темы еще готовятся")
}

func (s *Service) generateOnDemand(ctx context.Context, req GenerateRequest, target Target) (Task, error) {
	const maxAttempts = 3
	feedback := ""
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		content, source, err := s.generateContent(ctx, target, feedback)
		if err != nil {
			return Task{}, err
		}
		log.Printf("task generate on-demand: user_id=%d mode=%s target=%d/%s source=%s attempt=%d/%d",
			req.UserID, req.Mode, target.OgeNumber, target.SubtypeCode, source, attempt, maxAttempts)
		task, err := s.repo.CreateGeneratedTask(ctx, createGeneratedTask(req.UserID, storedMode(req), target, content, source))
		if err == nil {
			s.cacheOnDemandTask(ctx, target, content, source)
			return task, nil
		}
		if isAppConflict(err) {
			log.Printf("task generate on-demand duplicate skipped: user_id=%d mode=%s target=%d/%s attempt=%d/%d",
				req.UserID, req.Mode, target.OgeNumber, target.SubtypeCode, attempt, maxAttempts)
			feedback = duplicateGeneratedFeedback
			if source == "fallback" {
				break
			}
			continue
		}
		return Task{}, err
	}
	return Task{}, app.TaskUnavailable("Не удалось сгенерировать новую уникальную задачу")
}

func (s *Service) cacheOnDemandTask(ctx context.Context, target Target, content GeneratedContent, source string) {
	prepared, err := s.repo.CreatePreparedTask(ctx, createPreparedTask(target, content, source))
	if err == nil {
		log.Printf("task generate on-demand cached: prepared_id=%d target=%d/%s source=%s",
			prepared.ID, target.OgeNumber, target.SubtypeCode, source)
		return
	}
	if isAppConflict(err) {
		log.Printf("task generate on-demand already cached: target=%d/%s",
			target.OgeNumber, target.SubtypeCode)
		return
	}
	log.Printf("task generate on-demand cache store failed: target=%d/%s error=%v",
		target.OgeNumber, target.SubtypeCode, err)
}

func createGeneratedTask(userID int64, mode string, target Target, content GeneratedContent, source string) CreateTask {
	return CreateTask{
		UserID:          userID,
		Mode:            mode,
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
		Graphs:          content.Graphs,
	}
}

func createPreparedTask(target Target, content GeneratedContent, source string) CreateTask {
	return CreateTask{
		Mode:            ModeCustom,
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
		Graphs:          content.Graphs,
	}
}

func (s *Service) PrepareTask(ctx context.Context, target Target) (PreparedTask, error) {
	const maxAttempts = 5
	feedback := ""
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		content, source, err := s.generateContent(ctx, target, feedback)
		if err != nil {
			return PreparedTask{}, err
		}
		if err := s.reviewPreparedTask(ctx, target, content, source); err != nil {
			log.Printf("prepared task review rejected: target=%d/%s attempt=%d/%d source=%s error=%v",
				target.OgeNumber, target.SubtypeCode, attempt, maxAttempts, source, err)
			feedback = mergeGenerationFeedback(workerReviewFeedback, err)
			continue
		}
		prepared, err := s.repo.CreatePreparedTask(ctx, createPreparedTask(target, content, source))
		if err == nil {
			return prepared, nil
		}
		if isAppConflict(err) {
			log.Printf("prepared task duplicate skipped: target=%d/%s attempt=%d/%d",
				target.OgeNumber, target.SubtypeCode, attempt, maxAttempts)
			feedback = duplicatePreparedFeedback
			if source == "fallback" {
				break
			}
			continue
		}
		return PreparedTask{}, err
	}
	return PreparedTask{}, app.TaskUnavailable("Не удалось подготовить уникальную задачу")
}

func (s *Service) reviewPreparedTask(ctx context.Context, target Target, content GeneratedContent, source string) error {
	if s.ai == nil || !s.ai.IsConfigured() || source == "fallback" {
		return nil
	}
	review, err := s.ai.ReviewGeneratedTask(ctx, target, content)
	if err != nil {
		return err
	}
	review.Reason = strings.TrimSpace(review.Reason)
	review.CorrectAnswer = strings.TrimSpace(review.CorrectAnswer)
	if !review.IsValid {
		if review.Reason == "" {
			return errors.New("AI reviewer rejected generated task")
		}
		return fmt.Errorf("AI reviewer rejected generated task: %s", review.Reason)
	}
	if review.CorrectAnswer != "" && strings.TrimSpace(content.CorrectAnswer) != review.CorrectAnswer {
		return fmt.Errorf("AI reviewer expected correct_answer %q, got %q", review.CorrectAnswer, content.CorrectAnswer)
	}
	return nil
}

func (s *Service) generateContent(ctx context.Context, target Target, feedback string) (GeneratedContent, string, error) {
	if s.ai == nil || !s.ai.IsConfigured() {
		if RequiresVisualData(target) {
			err := errors.New("AI client is not configured")
			log.Printf("ai generation fallback: oge_number=%d subtype_code=%s reason=%v", target.OgeNumber, target.SubtypeCode, err)
			return FallbackGeneratedContent(target, err), "fallback", nil
		}
		return GeneratedContent{}, "", app.AIUnavailable(errors.New("AITUNNEL_API_KEY is empty"))
	}

	var lastErr error
	const maxAttempts = 3
	attemptFeedback := strings.TrimSpace(feedback)
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		started := time.Now()
		content, err := s.ai.GenerateTask(ctx, target, attemptFeedback)
		duration := time.Since(started)
		if err != nil {
			lastErr = err
			attemptFeedback = mergeGenerationFeedback(feedback, err)
			log.Printf("ai generation result: oge_number=%d subtype_code=%s attempt=%d duration_ms=%d validation=error error=%v",
				target.OgeNumber, target.SubtypeCode, attempt, duration.Milliseconds(), err)
			continue
		}

		if err := validateGeneratedContent(target, &content); err != nil {
			lastErr = err
			attemptFeedback = mergeGenerationFeedback(feedback, err)
			log.Printf("ai generation result: oge_number=%d subtype_code=%s attempt=%d duration_ms=%d validation=error error=%v",
				target.OgeNumber, target.SubtypeCode, attempt, duration.Milliseconds(), err)
			continue
		}

		log.Printf("ai generation result: oge_number=%d subtype_code=%s attempt=%d duration_ms=%d validation=success",
			target.OgeNumber, target.SubtypeCode, attempt, duration.Milliseconds())
		return content, "gpt-4o-mini", nil
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

const (
	duplicateGeneratedFeedback = "The previous generated task was rejected as a duplicate for this user. Generate a clearly different task: change the wording, scenario, all key numeric values, final answer, and visual_data if present."
	duplicatePreparedFeedback  = "The previous prepared task was rejected as a duplicate already stored for this target. Generate a clearly different task: change the wording, scenario, all key numeric values, final answer, and visual_data if present."
	workerReviewFeedback       = "The worker reviewer rejected the previous task. Fix the mathematical correctness, make correct_answer match the solved result exactly, and keep the task in the requested OGE subtype."
)

func mergeGenerationFeedback(base string, err error) string {
	base = strings.TrimSpace(base)
	errText := ""
	if err != nil {
		errText = strings.TrimSpace(err.Error())
	}
	switch {
	case base != "" && errText != "":
		return base + "; " + errText
	case base != "":
		return base
	default:
		return errText
	}
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
	if err := ValidateUserFacingText(*content); err != nil {
		return err
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
		target, err := s.repo.GetRandomTaskType(ctx)
		if isAppNotFound(err) {
			return Target{}, app.TaskUnavailable("Готовые задачи еще готовятся")
		}
		return target, err
	default:
		return Target{}, app.Validation("Некорректный режим тренировки")
	}
}

func isAppNotFound(err error) bool {
	var appErr *app.Error
	return errors.As(err, &appErr) && appErr.Code == app.CodeNotFound
}

func isAppConflict(err error) bool {
	var appErr *app.Error
	return errors.As(err, &appErr) && appErr.Code == app.CodeConflict
}

func storedMode(req GenerateRequest) string {
	if strings.TrimSpace(req.StoredMode) != "" {
		return strings.TrimSpace(req.StoredMode)
	}
	return req.Mode
}
