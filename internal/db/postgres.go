package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/oge-math-trainer/oge-backend.git/internal/app"
	"github.com/oge-math-trainer/oge-backend.git/internal/auth"
	"github.com/oge-math-trainer/oge-backend.git/internal/diagnostic"
	"github.com/oge-math-trainer/oge-backend.git/internal/progress"
	"github.com/oge-math-trainer/oge-backend.git/internal/tasks"
)

type Store struct {
	pool *pgxpool.Pool
}

const preparedTasksWorkerLockKey int64 = 2026052401

func New(ctx context.Context, databaseURL string) (*Store, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}
	return &Store{pool: pool}, nil
}

func (s *Store) Close() {
	if s != nil && s.pool != nil {
		s.pool.Close()
	}
}

func (s *Store) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if s == nil || s.pool == nil {
		return app.DBUnavailable(errors.New("postgres pool is nil"))
	}
	if err := s.pool.Ping(ctx); err != nil {
		return app.DBUnavailable(err)
	}
	return nil
}

func (s *Store) TryAcquirePreparationLock(ctx context.Context) (func(), bool, error) {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return nil, false, app.Internal(err)
	}

	var locked bool
	if err := conn.QueryRow(ctx, `SELECT pg_try_advisory_lock($1)`, preparedTasksWorkerLockKey).Scan(&locked); err != nil {
		conn.Release()
		return nil, false, app.Internal(err)
	}
	if !locked {
		conn.Release()
		return nil, false, nil
	}

	release := func() {
		_, _ = conn.Exec(context.Background(), `SELECT pg_advisory_unlock($1)`, preparedTasksWorkerLockKey)
		conn.Release()
	}
	return release, true, nil
}

func (s *Store) CreateUser(ctx context.Context, email, passwordHash string) (auth.User, error) {
	row := s.pool.QueryRow(ctx, `
		INSERT INTO users (email, password_hash)
		VALUES ($1, $2)
		RETURNING id, email, created_at
	`, email, passwordHash)

	var user auth.User
	if err := row.Scan(&user.ID, &user.Email, &user.CreatedAt); err != nil {
		if isUniqueViolation(err) {
			return auth.User{}, app.Conflict("Пользователь с таким email уже существует")
		}
		return auth.User{}, app.Internal(err)
	}
	return user, nil
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (auth.UserWithPassword, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, email, COALESCE(password_hash, ''), created_at
		FROM users
		WHERE lower(email) = lower($1)
	`, email)

	var user auth.UserWithPassword
	if err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return auth.UserWithPassword{}, app.NotFound("Пользователь не найден")
		}
		return auth.UserWithPassword{}, app.Internal(err)
	}
	return user, nil
}

func (s *Store) FindOrCreateOAuthUser(ctx context.Context, identity auth.OAuthIdentity) (auth.User, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return auth.User{}, app.Internal(err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	user, err := selectOAuthUser(ctx, tx, identity.Provider, identity.ProviderUserID)
	if err == nil {
		if err := tx.Commit(ctx); err != nil {
			return auth.User{}, app.Internal(err)
		}
		return user, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return auth.User{}, app.Internal(err)
	}

	user, err = selectUserByEmail(ctx, tx, identity.Email)
	if errors.Is(err, pgx.ErrNoRows) {
		user, err = insertOAuthUser(ctx, tx, identity.Email)
		if err != nil && isUniqueViolation(err) {
			user, err = selectUserByEmail(ctx, tx, identity.Email)
		}
	}
	if err != nil {
		return auth.User{}, app.Internal(err)
	}

	var linkedUserID int64
	if err := tx.QueryRow(ctx, `
		INSERT INTO user_identities (user_id, provider, provider_user_id, email)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (provider, provider_user_id)
		DO UPDATE SET email = EXCLUDED.email
		RETURNING user_id
	`, user.ID, identity.Provider, identity.ProviderUserID, identity.Email).Scan(&linkedUserID); err != nil {
		return auth.User{}, app.Internal(err)
	}
	if linkedUserID != user.ID {
		user, err = selectUserByID(ctx, tx, linkedUserID)
		if err != nil {
			return auth.User{}, app.Internal(err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return auth.User{}, app.Internal(err)
	}
	return user, nil
}

func selectOAuthUser(ctx context.Context, tx pgx.Tx, provider, providerUserID string) (auth.User, error) {
	row := tx.QueryRow(ctx, `
		SELECT u.id, u.email, u.created_at
		FROM user_identities ui
		JOIN users u ON u.id = ui.user_id
		WHERE ui.provider = $1 AND ui.provider_user_id = $2
	`, provider, providerUserID)

	var user auth.User
	if err := row.Scan(&user.ID, &user.Email, &user.CreatedAt); err != nil {
		return auth.User{}, err
	}
	return user, nil
}

func selectUserByEmail(ctx context.Context, tx pgx.Tx, email string) (auth.User, error) {
	row := tx.QueryRow(ctx, `
		SELECT id, email, created_at
		FROM users
		WHERE lower(email) = lower($1)
	`, email)

	var user auth.User
	if err := row.Scan(&user.ID, &user.Email, &user.CreatedAt); err != nil {
		return auth.User{}, err
	}
	return user, nil
}

func selectUserByID(ctx context.Context, tx pgx.Tx, id int64) (auth.User, error) {
	row := tx.QueryRow(ctx, `
		SELECT id, email, created_at
		FROM users
		WHERE id = $1
	`, id)

	var user auth.User
	if err := row.Scan(&user.ID, &user.Email, &user.CreatedAt); err != nil {
		return auth.User{}, err
	}
	return user, nil
}

func insertOAuthUser(ctx context.Context, tx pgx.Tx, email string) (auth.User, error) {
	row := tx.QueryRow(ctx, `
		INSERT INTO users (email, password_hash)
		VALUES ($1, NULL)
		RETURNING id, email, created_at
	`, email)

	var user auth.User
	if err := row.Scan(&user.ID, &user.Email, &user.CreatedAt); err != nil {
		return auth.User{}, err
	}
	return user, nil
}

func (s *Store) GetUserByID(ctx context.Context, id int64) (auth.User, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, email, created_at
		FROM users
		WHERE id = $1
	`, id)

	var user auth.User
	if err := row.Scan(&user.ID, &user.Email, &user.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return auth.User{}, app.NotFound("Пользователь не найден")
		}
		return auth.User{}, app.Internal(err)
	}
	return user, nil
}

func (s *Store) GetWeakTarget(ctx context.Context, userID int64) (tasks.Target, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT COALESCE(p.task_type_id, tt.id), p.oge_number, p.subtype_code
		FROM progress p
		LEFT JOIN task_types tt
			ON tt.oge_number = p.oge_number
			AND tt.subtype_code = p.subtype_code
		WHERE p.user_id = $1
		ORDER BY p.mastery_score ASC, p.attempts_count DESC, p.updated_at ASC
		LIMIT 1
	`, userID)
	return scanTarget(row, "Нет данных прогресса для режима weak")
}

func (s *Store) GetRandomTaskType(ctx context.Context) (tasks.Target, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, oge_number, subtype_code
		FROM task_types
		WHERE oge_number BETWEEN 6 AND 19
		ORDER BY random()
		LIMIT 1
	`)
	return scanTarget(row, "Нет готовых заданий для режима all")
}

func (s *Store) GetPreparationTarget(ctx context.Context, minReady int) (tasks.Target, int, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT tt.id, tt.oge_number, tt.subtype_code, COUNT(pt.id)::int
		FROM task_types tt
		LEFT JOIN prepared_tasks pt
			ON pt.used = false
			AND pt.shown_count < 20
			AND pt.oge_number = tt.oge_number
			AND pt.subtype_code = tt.subtype_code
		WHERE tt.oge_number BETWEEN 6 AND 19
		GROUP BY tt.id, tt.oge_number, tt.subtype_code
		HAVING COUNT(pt.id) < $1
		ORDER BY COUNT(pt.id) ASC, random()
		LIMIT 1
	`, minReady)

	target, err := scanTargetWithReadyCount(row, "Все категории заполнены")
	if err != nil {
		return tasks.Target{}, 0, err
	}
	return target.Target, target.ReadyCount, nil
}

func (s *Store) ResolveTarget(ctx context.Context, target tasks.Target) (tasks.Target, error) {
	if target.TaskTypeID != nil {
		return target, nil
	}

	row := s.pool.QueryRow(ctx, `
		SELECT id, oge_number, subtype_code
		FROM task_types
		WHERE oge_number = $1 AND subtype_code = $2
		LIMIT 1
	`, target.OgeNumber, target.SubtypeCode)

	resolved, err := scanTarget(row, "")
	if err != nil {
		var appErr *app.Error
		if errors.As(err, &appErr) && appErr.Code == app.CodeNotFound {
			return target, nil
		}
		return tasks.Target{}, err
	}
	return resolved, nil
}

func (s *Store) GetPreparedTask(ctx context.Context, userID int64, target tasks.Target) (tasks.PreparedTask, error) {
	row := s.pool.QueryRow(ctx, `
		WITH selected AS (
			SELECT pt.id
			FROM prepared_tasks pt
			WHERE pt.used = false
				AND pt.shown_count < 20
				AND pt.oge_number = $1
				AND pt.subtype_code = $2
				AND NOT EXISTS (
					SELECT 1
					FROM prepared_task_views pv
					WHERE pv.prepared_task_id = pt.id
						AND pv.user_id = $3
				)
				AND NOT EXISTS (
					SELECT 1
					FROM generated_tasks gt
					WHERE gt.user_id = $3
						AND gt.oge_number = pt.oge_number
						AND gt.subtype_code = pt.subtype_code
						AND md5(regexp_replace(lower(trim(gt.question)), '\s+', ' ', 'g')) =
						    md5(regexp_replace(lower(trim(pt.question)), '\s+', ' ', 'g'))
				)
			ORDER BY pt.shown_count ASC, pt.created_at ASC
			LIMIT 1
			FOR UPDATE SKIP LOCKED
		),
		viewed AS (
			INSERT INTO prepared_task_views (prepared_task_id, user_id)
			SELECT selected.id, $3
			FROM selected
			ON CONFLICT (prepared_task_id, user_id) DO NOTHING
			RETURNING prepared_task_id
		),
		updated AS (
			UPDATE prepared_tasks pt
			SET shown_count = pt.shown_count + 1,
				used = (pt.shown_count + 1) >= 20
			FROM viewed
			WHERE pt.id = viewed.prepared_task_id
			RETURNING pt.id, pt.mode, pt.task_type_id, pt.oge_number, pt.subtype_code, pt.question, pt.correct_answer,
				COALESCE(pt.solution_steps, '[]'::jsonb) AS solution_steps,
				COALESCE(pt.self_check, '') AS self_check,
				pt.is_valid,
				COALESCE(pt.validation_notes, '') AS validation_notes,
				pt.generation_source,
				COALESCE(pt.visual_data, '{}'::jsonb) AS visual_data,
				COALESCE(pt.graphs, '[]'::jsonb) AS graphs,
				pt.created_at,
				pt.shown_count
		)
		SELECT id, mode, task_type_id, oge_number, subtype_code, question, correct_answer,
			COALESCE(solution_steps, '[]'::jsonb), COALESCE(self_check, ''), is_valid,
			COALESCE(validation_notes, ''), generation_source,
			COALESCE(visual_data, '{}'::jsonb), COALESCE(graphs, '[]'::jsonb), created_at
		FROM updated
	`, target.OgeNumber, target.SubtypeCode, userID)

	prepared, err := scanPreparedTask(row)
	if err != nil {
		return tasks.PreparedTask{}, err
	}
	if err := s.deleteExhaustedPreparedTasks(ctx); err != nil {
		return tasks.PreparedTask{}, err
	}
	return prepared, nil
}

func (s *Store) deleteExhaustedPreparedTasks(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `
		DELETE FROM prepared_tasks
		WHERE used = true OR shown_count >= 20
	`)
	if err != nil {
		return app.Internal(err)
	}
	return nil
}

func (s *Store) CountPreparedTasks(ctx context.Context) (int, error) {
	var count int
	if err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM prepared_tasks
		WHERE used = false AND shown_count < 20
	`).Scan(&count); err != nil {
		return 0, app.Internal(err)
	}
	return count, nil
}

func (s *Store) CreatePreparedTask(ctx context.Context, task tasks.CreateTask) (tasks.PreparedTask, error) {
	steps, err := json.Marshal(task.SolutionSteps)
	if err != nil {
		return tasks.PreparedTask{}, app.Internal(err)
	}
	var visualDataJSON json.RawMessage
	if task.VisualData != nil {
		b, err := json.Marshal(task.VisualData)
		if err != nil {
			return tasks.PreparedTask{}, app.Internal(err)
		}
		visualDataJSON = json.RawMessage(b)
	}
	var graphsJSON json.RawMessage
	if len(task.Graphs) > 0 {
		b, err := json.Marshal(task.Graphs)
		if err != nil {
			return tasks.PreparedTask{}, app.Internal(err)
		}
		graphsJSON = json.RawMessage(b)
	}

	row := s.pool.QueryRow(ctx, `
		INSERT INTO prepared_tasks (
			mode,
			task_type_id,
			oge_number,
			subtype_code,
			question,
			correct_answer,
			solution_steps,
			self_check,
			is_valid,
			validation_notes,
			visual_data,
			graphs,
			generation_source,
			used
		)
		SELECT $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, false
		WHERE NOT EXISTS (
			SELECT 1
			FROM prepared_tasks
			WHERE used = false
				AND shown_count < 20
				AND oge_number = $3
				AND subtype_code = $4
				AND md5(regexp_replace(lower(trim(question)), '\s+', ' ', 'g')) =
				    md5(regexp_replace(lower(trim($5::text)), '\s+', ' ', 'g'))
		)
		RETURNING id, mode, task_type_id, oge_number, subtype_code, question, correct_answer,
			COALESCE(solution_steps, '[]'::jsonb), COALESCE(self_check, ''), is_valid,
			COALESCE(validation_notes, ''), generation_source,
			COALESCE(visual_data, '{}'::jsonb), COALESCE(graphs, '[]'::jsonb), created_at
	`, task.Mode, task.TaskTypeID, task.OgeNumber, task.SubtypeCode, task.Question, task.CorrectAnswer,
		string(steps), task.SelfCheck, task.IsValid, task.ValidationNotes, visualDataJSON, graphsJSON, task.Source)

	prepared, err := scanPreparedTask(row)
	if err != nil {
		if isAppCode(err, app.CodeNotFound) || isUniqueViolation(err) {
			return tasks.PreparedTask{}, app.Conflict("duplicate prepared task")
		}
		return tasks.PreparedTask{}, err
	}
	return prepared, nil
}

func (s *Store) CreateGeneratedTask(ctx context.Context, task tasks.CreateTask) (tasks.Task, error) {
	steps, err := json.Marshal(task.SolutionSteps)
	if err != nil {
		return tasks.Task{}, app.Internal(err)
	}
	var visualDataJSON json.RawMessage
	if task.VisualData != nil {
		b, err := json.Marshal(task.VisualData)
		if err != nil {
			return tasks.Task{}, app.Internal(err)
		}
		visualDataJSON = json.RawMessage(b)
	}
	var graphsJSON json.RawMessage
	if len(task.Graphs) > 0 {
		b, err := json.Marshal(task.Graphs)
		if err != nil {
			return tasks.Task{}, app.Internal(err)
		}
		graphsJSON = json.RawMessage(b)
	}

	row := s.pool.QueryRow(ctx, `
		INSERT INTO generated_tasks (
			user_id,
			mode,
			task_type_id,
			oge_number,
			subtype_code,
			question,
			correct_answer,
			solution_steps,
			self_check,
			is_valid,
			validation_notes,
			generation_source,
			visual_data,
			graphs
		)
		SELECT $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		WHERE NOT EXISTS (
			SELECT 1
			FROM generated_tasks
			WHERE user_id = $1
				AND mode = $2
				AND oge_number = $4
				AND subtype_code = $5
				AND md5(regexp_replace(lower(trim(question)), '\s+', ' ', 'g')) =
				    md5(regexp_replace(lower(trim($6::text)), '\s+', ' ', 'g'))
		)
		RETURNING id, user_id, mode, task_type_id, oge_number, subtype_code, question, correct_answer,
		          COALESCE(solution_steps, '[]'::jsonb), COALESCE(self_check, ''),
		          is_valid, COALESCE(validation_notes, ''), generation_source,
		          COALESCE(visual_data, '{}'::jsonb), COALESCE(graphs, '[]'::jsonb), created_at
	`, task.UserID, task.Mode, task.TaskTypeID, task.OgeNumber, task.SubtypeCode, task.Question, task.CorrectAnswer,
		string(steps), task.SelfCheck, task.IsValid, task.ValidationNotes, task.Source, visualDataJSON, graphsJSON)

	generated, err := scanGeneratedTask(row)
	if err != nil {
		if isAppCode(err, app.CodeNotFound) {
			return tasks.Task{}, app.Conflict("duplicate generated task")
		}
		return tasks.Task{}, err
	}
	return generated, nil
}

func (s *Store) GetGeneratedTaskForUser(ctx context.Context, userID, id int64) (tasks.Task, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, user_id, mode, task_type_id, oge_number, subtype_code, question, correct_answer,
		       COALESCE(solution_steps, '[]'::jsonb), COALESCE(self_check, ''),
		       is_valid, COALESCE(validation_notes, ''), generation_source,
		       COALESCE(visual_data, '{}'::jsonb), COALESCE(graphs, '[]'::jsonb), created_at
		FROM generated_tasks
		WHERE id = $1 AND user_id = $2
	`, id, userID)
	return scanGeneratedTask(row)
}

func (s *Store) SaveAttempt(ctx context.Context, userID, taskID int64, mode, studentAnswer string, isCorrect bool, aiFeedback any) error {
	var feedbackJSON json.RawMessage
	if aiFeedback != nil {
		b, err := json.Marshal(aiFeedback)
		if err != nil {
			return app.Internal(err)
		}
		feedbackJSON = json.RawMessage(b)
	}

	_, err := s.pool.Exec(ctx, `
		INSERT INTO attempts (user_id, generated_task_id, mode, student_answer, is_correct, ai_feedback)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, userID, taskID, mode, studentAnswer, isCorrect, feedbackJSON)
	if err != nil {
		return app.Internal(err)
	}
	return nil
}

func (s *Store) UpdateProgress(ctx context.Context, userID int64, task tasks.Task, isCorrect bool) error {
	correctDelta := 0
	if isCorrect {
		correctDelta = 1
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO progress (
			user_id,
			task_type_id,
			oge_number,
			subtype_code,
			attempts_count,
			correct_count,
			mastery_score
		)
		VALUES ($1, $2, $3, $4, 1, $5::int, $5::int::numeric)
		ON CONFLICT (user_id, oge_number, subtype_code)
		DO UPDATE SET
			task_type_id = COALESCE(EXCLUDED.task_type_id, progress.task_type_id),
			attempts_count = progress.attempts_count + 1,
			correct_count = progress.correct_count + $5::int,
			mastery_score = ((progress.correct_count + $5::int)::numeric / (progress.attempts_count + 1)::numeric),
			updated_at = now()
	`, userID, task.TaskTypeID, task.OgeNumber, task.SubtypeCode, correctDelta)
	if err != nil {
		return app.Internal(err)
	}
	return nil
}

func (s *Store) CreateDiagnosticSession(ctx context.Context, userID int64) (int64, error) {
	var id int64
	if err := s.pool.QueryRow(ctx, `
		INSERT INTO diagnostic_sessions (user_id, status)
		VALUES ($1, 'started')
		RETURNING id
	`, userID).Scan(&id); err != nil {
		return 0, app.Internal(err)
	}
	return id, nil
}

func (s *Store) EnsureDiagnosticSessionOwner(ctx context.Context, userID, sessionID int64) error {
	var exists bool
	if err := s.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM diagnostic_sessions
			WHERE id = $1 AND user_id = $2
		)
	`, sessionID, userID).Scan(&exists); err != nil {
		return app.Internal(err)
	}
	if !exists {
		return app.NotFound("Диагностическая сессия не найдена")
	}
	return nil
}

func (s *Store) GetDiagnosticTargets(ctx context.Context) ([]tasks.Target, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, oge_number, subtype_code
		FROM (
			SELECT DISTINCT ON (oge_number) id, oge_number, subtype_code
			FROM task_types
			WHERE oge_number BETWEEN 6 AND 19
			ORDER BY oge_number, random()
		) selected
		ORDER BY oge_number
	`)
	if err != nil {
		return nil, app.Internal(err)
	}
	defer rows.Close()

	targets, err := collectTargets(rows)
	if err != nil {
		return nil, err
	}
	if len(targets) > 0 {
		return targets, nil
	}

	return targets, nil
}

func (s *Store) SaveDiagnosticAnswer(ctx context.Context, sessionID int64, task tasks.Task, studentAnswer string, isCorrect bool, feedback any) error {
	var feedbackJSON json.RawMessage
	if feedback != nil {
		b, err := json.Marshal(feedback)
		if err != nil {
			return app.Internal(err)
		}
		feedbackJSON = json.RawMessage(b)
	}

	_, err := s.pool.Exec(ctx, `
		INSERT INTO diagnostic_answers (session_id, generated_task_id, student_answer, is_correct, ai_feedback)
		VALUES ($1, $2, $3, $4, $5)
	`, sessionID, task.ID, studentAnswer, isCorrect, feedbackJSON)
	if err != nil {
		return app.Internal(err)
	}
	return nil
}

func (s *Store) FinishDiagnosticSession(ctx context.Context, sessionID int64, analysis diagnostic.Analysis, weakTopics []string) error {
	analysisJSON, err := json.Marshal(analysis)
	if err != nil {
		return app.Internal(err)
	}
	weakTopicsJSON, err := json.Marshal(weakTopics)
	if err != nil {
		return app.Internal(err)
	}

	_, err = s.pool.Exec(ctx, `
		UPDATE diagnostic_sessions
		SET status = 'finished',
		    ai_analysis = $2,
		    weak_topics = $3,
		    finished_at = now()
		WHERE id = $1
	`, sessionID, json.RawMessage(analysisJSON), json.RawMessage(weakTopicsJSON))
	if err != nil {
		return app.Internal(err)
	}
	return nil
}

func (s *Store) AttachTaskToSession(ctx context.Context, sessionID, taskID int64) error {
	tag, err := s.pool.Exec(ctx, `
		INSERT INTO diagnostic_session_tasks (session_id, generated_task_id)
		SELECT $1, $2
		WHERE EXISTS (
			SELECT 1
			FROM diagnostic_sessions
			WHERE id = $1 AND status = 'started'
		)
		ON CONFLICT (session_id, generated_task_id) DO NOTHING
	`, sessionID, taskID)
	if err != nil {
		return app.Internal(err)
	}
	if tag.RowsAffected() == 0 {
		return app.NotFound("Диагностическая сессия завершена")
	}
	return nil
}

func (s *Store) GetTasksBySession(ctx context.Context, sessionID int64) ([]tasks.Task, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT gt.id, gt.user_id, gt.mode, gt.task_type_id, gt.oge_number, gt.subtype_code, 
		       gt.question, gt.correct_answer, COALESCE(gt.solution_steps, '[]'::jsonb), 
		       COALESCE(gt.self_check, ''), gt.is_valid, COALESCE(gt.validation_notes, ''), 
		       gt.generation_source, COALESCE(gt.visual_data, '{}'::jsonb), COALESCE(gt.graphs, '[]'::jsonb), gt.created_at
		FROM diagnostic_session_tasks dst
		JOIN generated_tasks gt ON dst.generated_task_id = gt.id
		WHERE dst.session_id = $1
		  AND gt.mode = 'diagnostic'
		ORDER BY gt.oge_number, dst.created_at
	`, sessionID)
	if err != nil {
		return nil, app.Internal(err)
	}
	defer rows.Close()

	out := make([]tasks.Task, 0)
	for rows.Next() {
		task, err := scanGeneratedTask(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, task)
	}
	if err := rows.Err(); err != nil {
		return nil, app.Internal(err)
	}
	return out, nil
}

func (s *Store) GetTaskBySession(ctx context.Context, sessionID, taskID int64) (tasks.Task, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT gt.id, gt.user_id, gt.mode, gt.task_type_id, gt.oge_number, gt.subtype_code, 
		       gt.question, gt.correct_answer, COALESCE(gt.solution_steps, '[]'::jsonb), 
		       COALESCE(gt.self_check, ''), gt.is_valid, COALESCE(gt.validation_notes, ''), 
		       gt.generation_source, COALESCE(gt.visual_data, '{}'::jsonb), COALESCE(gt.graphs, '[]'::jsonb), gt.created_at
		FROM diagnostic_session_tasks dst
		JOIN generated_tasks gt ON dst.generated_task_id = gt.id
		WHERE dst.session_id = $1
		  AND gt.id = $2
		  AND gt.mode = 'diagnostic'
	`, sessionID, taskID)
	return scanGeneratedTask(row)
}

func (s *Store) GetProgress(ctx context.Context, userID int64) ([]progress.Item, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT
			p.oge_number,
			p.subtype_code,
			COALESCE(t.title, ''),
			COALESCE(st.title, ''),
			COALESCE(tt.title, ''),
			p.attempts_count,
			p.correct_count,
			p.mastery_score::float8
		FROM progress p
		LEFT JOIN task_types tt ON tt.id = p.task_type_id
		LEFT JOIN topics t ON t.id = tt.topic_id
		LEFT JOIN subtopics st ON st.id = tt.subtopic_id
		WHERE p.user_id = $1
		ORDER BY p.mastery_score ASC, p.oge_number ASC, p.subtype_code ASC
	`, userID)
	if err != nil {
		return nil, app.Internal(err)
	}
	defer rows.Close()

	items := make([]progress.Item, 0)
	for rows.Next() {
		var item progress.Item
		if err := rows.Scan(&item.OgeNumber, &item.SubtypeCode, &item.TopicTitle, &item.SubtopicTitle, &item.TaskTitle, &item.AttemptsCount, &item.CorrectCount, &item.MasteryScore); err != nil {
			return nil, app.Internal(err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, app.Internal(err)
	}
	return items, nil
}

func (s *Store) GetProgressStats(ctx context.Context, userID int64) (progress.Stats, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT
			COALESCE(SUM(attempts_count), 0)::int,
			COALESCE(SUM(correct_count), 0)::int,
			COALESCE(AVG(mastery_score), 0)::float8
		FROM progress
		WHERE user_id = $1
	`, userID)

	var stats progress.Stats
	if err := row.Scan(&stats.TotalAttempts, &stats.CorrectCount, &stats.AverageMastery); err != nil {
		return progress.Stats{}, app.Internal(err)
	}
	if stats.TotalAttempts > 0 {
		stats.Accuracy = float64(stats.CorrectCount) / float64(stats.TotalAttempts)
	}
	return stats, nil
}

func (s *Store) GetRecommendations(ctx context.Context, userID int64) ([]progress.Recommendation, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT oge_number, subtype_code, mastery_score::float8
		FROM progress
		WHERE user_id = $1
		ORDER BY mastery_score ASC, attempts_count DESC, updated_at ASC
		LIMIT 5
	`, userID)
	if err != nil {
		return nil, app.Internal(err)
	}
	defer rows.Close()

	recommendations := make([]progress.Recommendation, 0)
	for rows.Next() {
		var rec progress.Recommendation
		if err := rows.Scan(&rec.OgeNumber, &rec.SubtypeCode, &rec.MasteryScore); err != nil {
			return nil, app.Internal(err)
		}
		rec.Message = "Рекомендуем потренировать эту тему"
		recommendations = append(recommendations, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, app.Internal(err)
	}
	return recommendations, nil
}

func scanTarget(row pgx.Row, notFoundMessage string) (tasks.Target, error) {
	var target tasks.Target
	var taskTypeID sql.NullInt64
	if err := row.Scan(&taskTypeID, &target.OgeNumber, &target.SubtypeCode); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return tasks.Target{}, app.NotFound(notFoundMessage)
		}
		return tasks.Target{}, app.Internal(err)
	}
	target.TaskTypeID = ptrFromNullInt64(taskTypeID)
	return target, nil
}

type targetWithReadyCount struct {
	Target     tasks.Target
	ReadyCount int
}

func scanTargetWithReadyCount(row pgx.Row, notFoundMessage string) (targetWithReadyCount, error) {
	var out targetWithReadyCount
	var taskTypeID sql.NullInt64
	if err := row.Scan(&taskTypeID, &out.Target.OgeNumber, &out.Target.SubtypeCode, &out.ReadyCount); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return targetWithReadyCount{}, app.NotFound(notFoundMessage)
		}
		return targetWithReadyCount{}, app.Internal(err)
	}
	out.Target.TaskTypeID = ptrFromNullInt64(taskTypeID)
	return out, nil
}

func collectTargets(rows pgx.Rows) ([]tasks.Target, error) {
	targets := make([]tasks.Target, 0, 14)
	for rows.Next() {
		var target tasks.Target
		var taskTypeID sql.NullInt64
		if err := rows.Scan(&taskTypeID, &target.OgeNumber, &target.SubtypeCode); err != nil {
			return nil, app.Internal(err)
		}
		target.TaskTypeID = ptrFromNullInt64(taskTypeID)
		targets = append(targets, target)
	}
	if err := rows.Err(); err != nil {
		return nil, app.Internal(err)
	}
	return targets, nil
}

func scanGeneratedTask(row pgx.Row) (tasks.Task, error) {
	var task tasks.Task
	var stepsJSON []byte
	var visualDataJSON []byte
	var graphsJSON []byte
	var taskTypeID sql.NullInt64
	if err := row.Scan(
		&task.ID,
		&task.UserID,
		&task.Mode,
		&taskTypeID,
		&task.OgeNumber,
		&task.SubtypeCode,
		&task.Question,
		&task.CorrectAnswer,
		&stepsJSON,
		&task.SelfCheck,
		&task.IsValid,
		&task.ValidationNotes,
		&task.Source,
		&visualDataJSON,
		&graphsJSON,
		&task.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return tasks.Task{}, app.NotFound("Задача не найдена")
		}
		return tasks.Task{}, app.Internal(err)
	}
	if len(stepsJSON) > 0 {
		_ = json.Unmarshal(stepsJSON, &task.SolutionSteps)
	}
	if len(visualDataJSON) > 0 && string(visualDataJSON) != "{}" && string(visualDataJSON) != "null" {
		var visualData tasks.VisualData
		if err := json.Unmarshal(visualDataJSON, &visualData); err != nil {
			return tasks.Task{}, app.Internal(err)
		}
		task.VisualData = visualData
	}
	if len(graphsJSON) > 0 && string(graphsJSON) != "[]" && string(graphsJSON) != "null" {
		var graphs []tasks.GraphInfo
		if err := json.Unmarshal(graphsJSON, &graphs); err != nil {
			return tasks.Task{}, app.Internal(err)
		}
		if len(graphs) > 0 {
			task.Graphs = graphs
		}
	}
	task.TaskTypeID = ptrFromNullInt64(taskTypeID)
	return task, nil
}

func scanPreparedTask(row pgx.Row) (tasks.PreparedTask, error) {
	var task tasks.PreparedTask
	var stepsJSON []byte
	var visualDataJSON []byte
	var graphsJSON []byte
	var taskTypeID sql.NullInt64
	if err := row.Scan(
		&task.ID,
		&task.Mode,
		&taskTypeID,
		&task.OgeNumber,
		&task.SubtypeCode,
		&task.Question,
		&task.CorrectAnswer,
		&stepsJSON,
		&task.SelfCheck,
		&task.IsValid,
		&task.ValidationNotes,
		&task.Source,
		&visualDataJSON,
		&graphsJSON,
		&task.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return tasks.PreparedTask{}, app.NotFound("Готовая задача не найдена")
		}
		return tasks.PreparedTask{}, app.Internal(err)
	}
	if len(stepsJSON) > 0 {
		_ = json.Unmarshal(stepsJSON, &task.SolutionSteps)
	}
	if len(visualDataJSON) > 0 && string(visualDataJSON) != "{}" && string(visualDataJSON) != "null" {
		var visualData tasks.VisualData
		if err := json.Unmarshal(visualDataJSON, &visualData); err != nil {
			return tasks.PreparedTask{}, app.Internal(err)
		}
		task.VisualData = visualData
	}
	if len(graphsJSON) > 0 && string(graphsJSON) != "[]" && string(graphsJSON) != "null" {
		var graphs []tasks.GraphInfo
		if err := json.Unmarshal(graphsJSON, &graphs); err != nil {
			return tasks.PreparedTask{}, app.Internal(err)
		}
		if len(graphs) > 0 {
			task.Graphs = graphs
		}
	}
	task.TaskTypeID = ptrFromNullInt64(taskTypeID)
	return task, nil
}

func ptrFromNullInt64(value sql.NullInt64) *int64 {
	if !value.Valid {
		return nil
	}
	return &value.Int64
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func isAppCode(err error, code string) bool {
	var appErr *app.Error
	return errors.As(err, &appErr) && appErr.Code == code
}
