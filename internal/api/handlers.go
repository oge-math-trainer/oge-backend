package api

import (
	"log"
	"net/http"
	"strings"

	"oge-ai-trainer/backend/internal/app"
	"oge-ai-trainer/backend/internal/diagnostic"
	"oge-ai-trainer/backend/internal/tasks"
)

type credentialsRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type generateTaskRequest struct {
	Mode        string `json:"mode"`
	OgeNumber   *int   `json:"oge_number,omitempty"`
	SubtypeCode string `json:"subtype_code,omitempty"`
}

type checkTaskRequest struct {
	TaskID        int64  `json:"task_id"`
	StudentAnswer string `json:"student_answer"`
}

type taskIDRequest struct {
	TaskID int64 `json:"task_id"`
}

type diagnosticSubmitRequest struct {
	SessionID int64                    `json:"session_id"`
	Answers   []diagnostic.AnswerInput `json:"answers"`
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if s.health == nil {
		writeError(w, r, app.DBUnavailable(nil))
		return
	}
	if err := s.health.Ping(r.Context()); err != nil {
		writeError(w, r, err)
		return
	}
	writeSuccess(w, http.StatusOK, map[string]string{
		"status": "ok",
		"db":     "ok",
	})
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req credentialsRequest
	if err := decodeAuthJSON(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}
	if err := validateCredentials(req.Email, req.Password); err != nil {
		writeError(w, r, err)
		return
	}

	session, err := s.auth.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		logAuthFailure(r, "register", err)
		writeError(w, r, err)
		return
	}
	writeSuccess(w, http.StatusCreated, session)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req credentialsRequest
	if err := decodeAuthJSON(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}
	if err := validateCredentials(req.Email, req.Password); err != nil {
		writeError(w, r, err)
		return
	}

	session, err := s.auth.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		logAuthFailure(r, "login", err)
		writeError(w, r, err)
		return
	}
	writeSuccess(w, http.StatusOK, session)
}

func logAuthFailure(r *http.Request, action string, err error) {
	log.Printf("auth %s failed: request_id=%s ip=%s err=%v", action, requestIDFromContext(r.Context()), remoteIP(r), err)
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	user, err := s.auth.Me(r.Context(), userIDFromContext(r.Context()))
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeSuccess(w, http.StatusOK, user)
}

func (s *Server) handleDiagnosticStart(w http.ResponseWriter, r *http.Request) {
	result, err := s.diag.Start(r.Context(), userIDFromContext(r.Context()))
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeSuccess(w, http.StatusOK, result)
}

func (s *Server) handleDiagnosticSubmit(w http.ResponseWriter, r *http.Request) {
	var req diagnosticSubmitRequest
	if err := decodeStrictJSON(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}
	if err := validatePositiveID(req.SessionID, "session_id"); err != nil {
		writeError(w, r, err)
		return
	}
	if len(req.Answers) == 0 {
		writeError(w, r, app.Validation("answers не должен быть пустым"))
		return
	}
	for _, answer := range req.Answers {
		if err := validatePositiveID(answer.TaskID, "task_id"); err != nil {
			writeError(w, r, err)
			return
		}
		if strings.TrimSpace(answer.StudentAnswer) == "" {
			writeError(w, r, app.Validation("student_answer обязателен"))
			return
		}
	}

	result, err := s.diag.Submit(r.Context(), userIDFromContext(r.Context()), req.SessionID, req.Answers)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeSuccess(w, http.StatusOK, result)
}

func (s *Server) handleTaskGenerate(w http.ResponseWriter, r *http.Request) {
	var req generateTaskRequest
	if err := decodeStrictJSON(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}

	mode := strings.TrimSpace(req.Mode)
	if mode != tasks.ModeWeak && mode != tasks.ModeAll && mode != tasks.ModeCustom {
		writeError(w, r, app.Validation("mode должен быть weak, all или custom"))
		return
	}
	if mode == tasks.ModeCustom {
		if req.OgeNumber == nil {
			writeError(w, r, app.Validation("Для режима custom нужен oge_number"))
			return
		}
		if err := validateOgeNumber(*req.OgeNumber); err != nil {
			writeError(w, r, err)
			return
		}
		if strings.TrimSpace(req.SubtypeCode) == "" {
			writeError(w, r, app.Validation("Для режима custom нужен subtype_code"))
			return
		}
	}

	task, err := s.tasks.Generate(r.Context(), tasks.GenerateRequest{
		UserID:      userIDFromContext(r.Context()),
		Mode:        mode,
		OgeNumber:   req.OgeNumber,
		SubtypeCode: strings.TrimSpace(req.SubtypeCode),
	})
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeSuccess(w, http.StatusOK, task)
}

func (s *Server) handleTaskCheck(w http.ResponseWriter, r *http.Request) {
	var req checkTaskRequest
	if err := decodeStrictJSON(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}
	if err := validatePositiveID(req.TaskID, "task_id"); err != nil {
		writeError(w, r, err)
		return
	}
	if strings.TrimSpace(req.StudentAnswer) == "" {
		writeError(w, r, app.Validation("student_answer обязателен"))
		return
	}

	result, err := s.tasks.Check(r.Context(), tasks.CheckRequest{
		UserID:        userIDFromContext(r.Context()),
		TaskID:        req.TaskID,
		StudentAnswer: strings.TrimSpace(req.StudentAnswer),
	})
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeSuccess(w, http.StatusOK, result)
}

func (s *Server) handleTaskHint(w http.ResponseWriter, r *http.Request) {
	var req taskIDRequest
	if err := decodeStrictJSON(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}
	if err := validatePositiveID(req.TaskID, "task_id"); err != nil {
		writeError(w, r, err)
		return
	}
	result, err := s.tasks.Hint(r.Context(), userIDFromContext(r.Context()), req.TaskID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeSuccess(w, http.StatusOK, result)
}

func (s *Server) handleTaskExplain(w http.ResponseWriter, r *http.Request) {
	var req taskIDRequest
	if err := decodeStrictJSON(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}
	if err := validatePositiveID(req.TaskID, "task_id"); err != nil {
		writeError(w, r, err)
		return
	}
	result, err := s.tasks.Explain(r.Context(), userIDFromContext(r.Context()), req.TaskID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeSuccess(w, http.StatusOK, result)
}

func (s *Server) handleProgress(w http.ResponseWriter, r *http.Request) {
	items, err := s.progress.GetProgress(r.Context(), userIDFromContext(r.Context()))
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeSuccess(w, http.StatusOK, map[string]any{"items": items})
}

func (s *Server) handleProgressStats(w http.ResponseWriter, r *http.Request) {
	stats, err := s.progress.GetStats(r.Context(), userIDFromContext(r.Context()))
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeSuccess(w, http.StatusOK, stats)
}

func (s *Server) handleProgressRecommendations(w http.ResponseWriter, r *http.Request) {
	recommendations, err := s.progress.GetRecommendations(r.Context(), userIDFromContext(r.Context()))
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeSuccess(w, http.StatusOK, map[string]any{"items": recommendations})
}
