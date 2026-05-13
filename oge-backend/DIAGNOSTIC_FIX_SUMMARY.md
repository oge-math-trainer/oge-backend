## Анализ и исправление ошибки 500 на POST /api/v1/diagnostic/submit

### Исходная проблема

Эндпоинт возвращает **500 Internal Server Error** вместо правильного **404 Not Found** при отправке неправильных `task_id` во время диагностического теста.

### Корневые причины

#### 1. **Дублирование createDiagnosticSession**
- В `service.go` строки 129-134 функция `CreateDiagnosticSession` вызывалась ДВА РАЗА
- Первая сессия создавалась, но переменная перезаписывалась
- Вторая сессия создавалась и использовалась дальше
- Хотя это не ломало функцию, это было грубой ошибкой

#### 2. **AttachTaskToSession не реализована**
- В `service.go` строке 140 вызывалась `s.repo.AttachTaskToSession`, но:
  - Метод не был определен в интерфейсе Repository
  - Метод не был реализован в `postgres.go`
  - Ошибка молча игнорировалась: `_ = s.repo.AttachTaskToSession(...)`

#### 3. **Нет архитектурной привязки task → session**
- **Текущая модель:** `GetGeneratedTaskForUser(userID, taskID)`
- Проверяла только наличие task_id у конкретного пользователя
- **Проблема:** Фронт мог отправить любой task_id:
  - Который существует для этого пользователя
  - Но не был выдан для этой конкретной сессии
  - Или был выдан для другой диагностической сессии
  - Или вообще был создан в другом режиме (weak/all/custom)

#### 4. **Последствия**
- При Start() задачи НЕ сохранялись как привязанные к сессии
- При Submit() проверка `GetGeneratedTaskForUser(userID, taskID)` проходила успешно даже для "чужих" задач
- Это привело бы к некорректным результатам или потенциально 500 ошибкам в других местах

### Реализованное решение

#### 1. Удалено дублирование (service.go, линия 129-134)
```diff
- sessionID, err := s.repo.CreateDiagnosticSession(ctx, userID)
- if err != nil {
-     return StartResult{}, err
- }
-
- sessionID, err := s.repo.CreateDiagnosticSession(ctx, userID)  // ДУБЛИРОВАНИЕ!
- if err != nil {
-     return StartResult{}, err
- }
- for _, task := range out {
-     _ = s.repo.AttachTaskToSession(ctx, sessionID, task.ID)    // Ошибки игнорируются!
- }
+ sessionID, err := s.repo.CreateDiagnosticSession(ctx, userID)
+ if err != nil {
+     return StartResult{}, err
+ }
+
+ for _, task := range out {
+     if err := s.repo.AttachTaskToSession(ctx, sessionID, task.ID); err != nil {
+         return StartResult{}, err
+     }
+ }
```

#### 2. Создана таблица diagnostic_session_tasks (миграция 004)
```sql
CREATE TABLE diagnostic_session_tasks (
    id BIGSERIAL PRIMARY KEY,
    session_id BIGINT NOT NULL REFERENCES diagnostic_sessions(id) ON DELETE CASCADE,
    generated_task_id BIGINT NOT NULL REFERENCES generated_tasks(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT diagnostic_session_tasks_session_task_uidx UNIQUE (session_id, generated_task_id)
);

CREATE INDEX diagnostic_session_tasks_session_id_idx
    ON diagnostic_session_tasks (session_id);
```

Таблица хранит явную связь между сессией диагностики и выданными задачами.

#### 3. Реализованы методы в postgres.go

**AttachTaskToSession** - сохраняет связь:
```go
func (s *Store) AttachTaskToSession(ctx context.Context, sessionID, taskID int64) error {
    _, err := s.pool.Exec(ctx, `
        INSERT INTO diagnostic_session_tasks (session_id, generated_task_id)
        VALUES ($1, $2)
        ON CONFLICT (session_id, generated_task_id) DO NOTHING
    `, sessionID, taskID)
    if err != nil {
        return app.Internal(err)
    }
    return nil
}
```

**GetTaskBySession** - получает задачу только если она привязана к сессии:
```go
func (s *Store) GetTaskBySession(ctx context.Context, sessionID, taskID int64) (tasks.Task, error) {
    row := s.pool.QueryRow(ctx, `
        SELECT gt.id, gt.user_id, gt.mode, ...
        FROM diagnostic_session_tasks dst
        JOIN generated_tasks gt ON dst.generated_task_id = gt.id
        WHERE dst.session_id = $1 AND gt.id = $2
    `, sessionID, taskID)
    return scanGeneratedTask(row)
}
```

Если задача не привязана к сессии, `scanGeneratedTask` вернёт **app.NotFound()**, что будет преобразовано в HTTP 404.

#### 4. Обновлен интерфейс Repository (service.go)
```go
type Repository interface {
    // ... existing methods ...
    AttachTaskToSession(ctx context.Context, sessionID, taskID int64) error
    GetTaskBySession(ctx context.Context, sessionID, taskID int64) (tasks.Task, error)
    // ... other methods ...
}
```

#### 5. Обновлена логика Submit (service.go, линия 164)
```diff
- task, err := s.repo.GetGeneratedTaskForUser(ctx, userID, answer.TaskID)
+ task, err := s.repo.GetTaskBySession(ctx, sessionID, answer.TaskID)
  if err != nil {
      return SubmitResult{}, err
  }
```

Теперь проверка гарантирует, что задача принадлежит именно этой сессии.

#### 6. Обновлены тесты (service_test.go)
Добавлены методы `AttachTaskToSession` и `GetTaskBySession` в `fakeRepo` для тестов.

### Как это решает проблему 500

**До исправления:**
1. Фронт отправляет task_id, который не был выдан для этой сессии
2. `GetGeneratedTaskForUser(userID, taskID)` находит задачу (она существует для пользователя)
3. Submit продолжает обработку с "чужой" задачей
4. Различные операции могут сломаться, вернув 500

**После исправления:**
1. Фронт отправляет task_id, который не был выдан для этой сессии
2. `GetTaskBySession(sessionID, taskID)` НЕ находит задачу
3. `scanGeneratedTask` вернёт `app.NotFound("Задача не найдена")`
4. Ошибка будет преобразована в HTTP 404 (не 500)
5. Фронт получит правильный HTTP статус

### Дополнительные преимущества

1. **Безопасность:** Пользователь не может использовать чужие задачи из других сессий
2. **Корректность:** Ошибка "Задача не найдена" будет возвращаться как 404, а не 500
3. **Отладка:** Явная таблица связи облегчает отладку и аналитику
4. **Масштабируемость:** Архитектура готова к расширению (например, добавление других параметров сессии)

### Файлы, которые были изменены

1. `internal/diagnostic/service.go`
   - Удалено дублирование CreateDiagnosticSession
   - Добавлены методы в интерфейс Repository
   - Обновлена логика Submit

2. `internal/db/postgres.go`
   - Реализованы AttachTaskToSession и GetTaskBySession

3. `internal/diagnostic/service_test.go`
   - Обновлены методы fakeRepo

4. `migrations/004_diagnostic_session_tasks.sql` (новый файл)
   - Создана таблица diagnostic_session_tasks

### Тестирование

✅ Go build успешен
✅ Диагностические тесты прошли (TestSubmitChecksSessionOwnerBeforeAI)
✅ Код компилируется без ошибок
