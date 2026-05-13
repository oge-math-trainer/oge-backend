# Реализация API endpoints для oge-ai-trainer

Этот файл нужен, чтобы можно было объяснить, что было сделано в backend, какие файлы за что отвечают и как проверить API руками.

## 1. Общая логика реализации

Backend теперь разделён на слои:

- `cmd/server` запускает приложение.
- `internal/config` читает настройки из `.env`.
- `internal/db` работает с PostgreSQL.
- `internal/auth`, `internal/tasks`, `internal/diagnostic`, `internal/progress`, `internal/ai` содержат основную логику.
- `internal/api` содержит только HTTP-слой: маршруты, handlers, middleware, валидацию входа и JSON-ответы.

Главное правило: API-слой не хранит бизнес-логику. Например, handler `/api/v1/tasks/check` только читает `task_id` и `student_answer`, проверяет формат и вызывает `tasks.Service.Check`.

Все пользовательские маршруты идут через префикс `/api/v1`. Технический маршрут `/health` оставлен без префикса и без авторизации.

Security-усиления после ревью:

- JWT проверяется стандартной библиотекой `github.com/golang-jwt/jwt/v5`, только с алгоритмом `HS256`.
- На `/api/v1/auth/register` и `/api/v1/auth/login` добавлен простой in-memory rate limiting по IP и path. Лимиты настраиваются через `.env`.
- Каждый ответ получает `X-Request-ID`; этот id пишется в server logs и помогает искать конкретный запрос.
- Добавлены базовые security headers: `X-Content-Type-Options`, `X-Frame-Options`, `Referrer-Policy`, `Permissions-Policy`, `Content-Security-Policy`.
- При login для несуществующего email выполняется dummy bcrypt compare, чтобы уменьшить timing leak между "email не найден" и "пароль неверный".
- Сгенерированная задача хранит `user_id` и `mode`, поэтому `/tasks/check`, `/tasks/hint`, `/tasks/explain` работают только с задачами текущего пользователя.
- Если `OPENROUTER_API_KEY` отсутствует, AI endpoints возвращают `ai_unavailable`; fake responses и fallback-генерация для MVP отключены.
- `diagnostic/submit` проверяет, что `session_id` принадлежит текущему пользователю.
- `OPENROUTER_MODEL` задаётся через `.env`; hardcoded default model убран.
- Пароль для MVP: от 8 до 72 байт, потому что bcrypt безопасно обрабатывает только первые 72 байта.

## 2. Созданные и изменённые файлы

### `cmd/server/main.go`

Это точка входа в backend.

Что делает файл:

1. Загружает `.env` через `config.Load(".env")`.
2. Проверяет обязательные настройки: `DATABASE_URL`, `AUTH_SECRET`, `TOKEN_TTL`, `PORT`.
3. Создаёт PostgreSQL pool через `db.New`. Если сама БД временно недоступна, сервер всё равно может подняться, а `/health` покажет `db_unavailable`.
4. Создаёт сервисы:
   - `authService` для регистрации, входа и JWT;
   - `taskService` для генерации, проверки, подсказок и объяснений;
   - `diagnosticService` для диагностики;
   - `progressService` для личного прогресса;
   - `aiClient` для OpenRouter.
5. Передаёт все сервисы в `api.NewRouter`.
6. Запускает HTTP-сервер с timeouts.

Как объяснить: `main.go` не обрабатывает запросы сам. Он только собирает приложение из готовых модулей и запускает сервер.

### `.env.example`

Пример переменных окружения без реальных секретов.

Ключевые переменные:

- `PORT` - порт backend.
- `DATABASE_URL` - строка подключения к PostgreSQL.
- `AUTH_SECRET` - секрет для подписи JWT. Его нельзя хранить в коде.
- `TOKEN_TTL` - время жизни токена, например `24h`.
- `AUTH_RATE_LIMIT_REQUESTS` - сколько auth-запросов разрешено за окно.
- `AUTH_RATE_LIMIT_WINDOW` - окно rate limit, например `1m`.
- `CORS_ALLOWED_ORIGINS` - frontend origin для CORS.
- `OPENROUTER_API_KEY` - ключ OpenRouter. Его нельзя хранить в коде.
- `OPENROUTER_MODEL` - модель для AI-запросов. Если `OPENROUTER_API_KEY` задан, модель тоже обязательна.

Как объяснить: реальные пароли и API-ключи должны быть только в `.env`, а в репозитории лежит только безопасный пример.

### `go.mod` и `go.sum`

Добавлены зависимости:

- `github.com/jackc/pgx/v5` - подключение к PostgreSQL.
- `github.com/golang-jwt/jwt/v5` - стандартная проверка и выпуск JWT.
- `golang.org/x/crypto/bcrypt` - безопасное хэширование паролей.

Как объяснить: pgx нужен для работы с БД, jwt/v5 нужен, чтобы не писать JWT вручную, bcrypt нужен, чтобы никогда не хранить пароль в открытом виде.

### `internal/app/errors.go`

Общий файл с кодами ошибок приложения.

Основные коды:

- `validation_error` - неверные входные данные.
- `unauthorized` - нет токена или неверный токен.
- `not_found` - объект не найден.
- `conflict` - конфликт, например email уже занят.
- `ai_unavailable` - AI недоступен.
- `db_unavailable` - PostgreSQL недоступен.
- `rate_limited` - слишком много auth-запросов с одного IP/path.
- `internal_error` - внутренняя ошибка сервера.

Как работает:

- сервисы возвращают `app.Error`;
- API-слой превращает эту ошибку в единый JSON;
- внутренний текст Go/SQL ошибки пользователю не отдаётся.

Как объяснить: этот файл нужен, чтобы все части backend говорили об ошибках на одном языке.

### `internal/config/config.go`

Читает настройки из `.env` и переменных окружения.

Что важно:

- `.env` загружается безопасно: если переменная уже задана в окружении, она не перетирается.
- `AUTH_SECRET`, `DATABASE_URL`, `OPENROUTER_API_KEY` не записаны в код.
- `AUTH_SECRET` должен быть не короче 32 символов.
- если задан `OPENROUTER_API_KEY`, то `OPENROUTER_MODEL` тоже обязателен.
- `AUTH_RATE_LIMIT_REQUESTS` и `AUTH_RATE_LIMIT_WINDOW` позволяют поменять лимиты без правки кода.
- `CORS_ALLOWED_ORIGINS` читается как список через запятую.
- если `TOKEN_TTL` не задан, используется `24h`.

Как объяснить: config отделяет настройки от кода, поэтому секреты можно менять без изменения исходников.

### `internal/api/router.go`

Подключает все маршруты.

Публичные маршруты:

- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `GET /health`

Защищённые маршруты:

- `GET /api/v1/auth/me`
- `POST /api/v1/diagnostic/start`
- `POST /api/v1/diagnostic/submit`
- `POST /api/v1/tasks/generate`
- `POST /api/v1/tasks/check`
- `POST /api/v1/tasks/hint`
- `POST /api/v1/tasks/explain`
- `GET /api/v1/progress`
- `GET /api/v1/progress/stats`
- `GET /api/v1/progress/recommendations`

Как работает:

- сначала создаётся `http.ServeMux`;
- публичный `/health` подключается напрямую;
- auth routes подключаются через rate limiter;
- защищённые маршруты оборачиваются в `requireAuth`;
- сверху подключаются CORS и recover middleware.

Как объяснить: `router.go` - это карта API. В нём видно, какие URL существуют и какие из них требуют токен.

### `internal/api/handlers.go`

Содержит HTTP-handlers.

Что делает handler:

1. Читает JSON через `decodeStrictJSON`, а для auth routes через `decodeAuthJSON`.
2. Валидирует поля.
3. Берёт `userID` из JWT-контекста, если маршрут защищённый.
4. Вызывает нужный сервис.
5. Возвращает `writeSuccess` или `writeError`.

Важные решения:

- `user_id` не принимается из body.
- `/api/v1/tasks/check` принимает только `task_id` и `student_answer`.
- `/api/v1/tasks/generate`:
  - `custom` требует `oge_number` и `subtype_code`;
  - `weak` не требует их, тема выбирается из `progress`;
  - `all` не требует их, тип задания выбирает backend.
- `/api/v1/diagnostic/submit` принимает `session_id`, но сервис проверяет владельца session по JWT userID.

Как объяснить: handlers - это входные двери API. Они проверяют запрос, но не решают учебные задачи сами.

### `internal/api/json.go`

Отвечает за строгий JSON.

Проверки:

- `Content-Type` должен быть `application/json`;
- для `/api/v1/auth/register` и `/api/v1/auth/login` пустой `Content-Type` разрешён, но если заголовок указан, он должен быть `application/json`;
- тело не должно быть пустым;
- JSON должен быть корректным;
- неизвестные поля запрещены через `DisallowUnknownFields`;
- после первого JSON-объекта не должно быть лишних данных;
- размер body ограничен.

Пример: если отправить в `/tasks/check` поле `mode`, backend вернёт `validation_error`, потому что этот endpoint не принимает `mode`.

Как объяснить: строгий JSON защищает frontend от неожиданных форматов и помогает быстро находить ошибки в запросах.

### `internal/api/middleware.go`

Middleware выполняются до handlers.

Что есть:

- `corsMiddleware` разрешает frontend с `localhost` и `127.0.0.1`.
- `recoverMiddleware` ловит panic и возвращает безопасный `internal_error`.
- `requireAuth` проверяет `Authorization: Bearer <token>`.
- `requestIDMiddleware` ставит `X-Request-ID` в response и кладёт id в context.
- `securityHeadersMiddleware` добавляет базовые защитные HTTP headers.
- `authRateLimit` ограничивает частые запросы на register/login.

Как работает JWT:

1. `requireAuth` достаёт token из header.
2. Передаёт token в `auth.ParseToken`.
3. `auth.ParseToken` проверяет подпись, срок действия и строго требует `HS256`.
4. Если token валидный, кладёт `userID` в context.
5. Handler берёт `userID` только из context.

Как объяснить: middleware - это общий фильтр перед запросами: request id, security headers, CORS, защита, rate limiting и аварийная обработка ошибок.

### `internal/api/responses.go`

Формирует единый JSON-ответ.

Успешный ответ:

```json
{
  "success": true,
  "data": {}
}
```

Ответ с ошибкой:

```json
{
  "success": false,
  "error": {
    "code": "validation_error",
    "message": "Некорректные входные данные"
  }
}
```

Как объяснить: frontend всегда получает один и тот же формат, поэтому ему проще обрабатывать ответы.

### `internal/api/validation.go`

Небольшие функции проверки входных данных.

Проверяет:

- email;
- длину пароля: минимум 8 байт, максимум 72 байта;
- положительные ID;
- номер ОГЭ от 6 до 19.

Как объяснить: повторяющаяся валидация вынесена отдельно, чтобы handlers были проще.

### `internal/auth/service.go`

Сервис регистрации, входа и JWT.

Что реализовано:

- `Register`:
  - нормализует email;
  - хэширует пароль через bcrypt;
  - создаёт пользователя в БД;
  - возвращает JWT.
- `Login`:
  - ищет пользователя по email;
  - сравнивает пароль с bcrypt hash;
  - возвращает JWT.
- `Me`:
  - возвращает текущего пользователя по `userID`.
- `ParseToken`:
  - использует `github.com/golang-jwt/jwt/v5`;
  - проверяет подпись JWT;
  - принимает только `HS256`;
  - проверяет срок действия;
  - возвращает `userID`.
- при login для несуществующего email выполняет dummy bcrypt compare, чтобы снизить timing leak.

Важно: пароль никогда не сохраняется как `password`. В БД пишется только `password_hash`.
Password policy для MVP: 8-72 байта, без complexity rules. Верхняя граница 72 байта нужна, потому что bcrypt игнорирует хвост пароля после этого лимита.

Как объяснить: auth отвечает за безопасность пользователя: пароль хэшируется, а дальнейшие запросы работают по токену.

### `internal/db/postgres.go`

Слой работы с PostgreSQL.

Что делает:

- создаёт pgx pool для БД;
- проверяет БД для `/health`;
- создаёт и ищет пользователей;
- выбирает weak/all/custom target для задач;
- для `weak` берёт самую слабую запись из `progress` и пытается восстановить `task_type_id` через `task_types`;
- для `custom` пытается найти `task_type_id` по `oge_number + subtype_code`, но не блокирует MVP, если справочник ещё пустой;
- сохраняет generated tasks вместе с `user_id` и `mode`;
- получает задачу по `task_id` только для текущего пользователя;
- сохраняет attempts;
- обновляет progress;
- создаёт и завершает diagnostic sessions;
- проверяет, что diagnostic session принадлежит текущему пользователю;
- сохраняет diagnostic answers;
- отдаёт progress, stats и recommendations.

Как объяснить: этот файл - единственное место, где backend знает SQL. Остальные пакеты вызывают методы репозитория и не пишут SQL напрямую.

### `migrations/002_generated_tasks_ownership.sql`

Миграция для уже созданных локальных БД.

Что добавляет:

- `generated_tasks.user_id` - владелец задачи;
- `generated_tasks.mode` - режим, в котором задача была создана;
- check constraint на режимы `weak`, `all`, `custom`, `diagnostic`;
- индексы по `user_id` и `user_id + mode`.

Как объяснить: эта миграция нужна, чтобы `/tasks/check` не принимал `mode` из body, а брал настоящий mode из сохранённой задачи и не позволял ученику проверять чужую задачу.

### `internal/ai/client.go`

Клиент OpenRouter.

Что реализовано:

- проверка `IsConfigured`, то есть есть ли `OPENROUTER_API_KEY`;
- генерация задачи;
- проверка ответа ученика;
- подсказка;
- объяснение;
- анализ диагностики.

Важное поведение:

- если `OPENROUTER_API_KEY` пустой, AI не вызывается;
- если `OPENROUTER_API_KEY` задан, `OPENROUTER_MODEL` должен быть задан через `.env`;
- hardcoded default model в коде не используется;
- fake AI responses не создаются;
- для generate/check/hint/explain/diagnostic analysis без ключа будет `ai_unavailable`.

Как объяснить: AI-клиент отвечает только за общение с OpenRouter, а не за маршруты или БД.

### `internal/tasks/service.go`

Сервис тренировочных задач.

Что делает:

- `Generate`:
  - определяет target по режиму;
  - если AI настроен, просит AI сгенерировать задачу;
  - если AI не настроен, возвращает `ai_unavailable`;
  - сохраняет задачу в `generated_tasks` с `user_id` и `mode`.
- `StoredMode` используется только внутренними сценариями: диагностика генерирует конкретные номера как `custom`, но сохраняет `mode=diagnostic`.
- `Check`:
  - получает задачу по `task_id` только для текущего пользователя;
  - проверяет ответ через AI;
  - сохраняет попытку с тем `mode`, в котором задача была сгенерирована;
  - обновляет progress.
- `Hint` и `Explain`:
  - получают задачу по `task_id`;
  - вызывают AI.

Как работают режимы:

- `custom`: frontend передаёт `oge_number` и `subtype_code`.
- `weak`: backend берёт самую слабую тему из `progress`.
- `all`: backend выбирает случайный тип задания.
- `diagnostic`: frontend не передаёт этот режим; его ставит backend при создании задач диагностики.

Как объяснить: tasks service - учебная логика тренировки. API передаёт запрос, а сервис решает, откуда взять тему и как сохранить результат.

### `internal/diagnostic/service.go`

Сервис диагностики.

Что делает:

- `Start`:
  - выбирает 14 targets для номеров 6-19;
  - генерирует задачи через `tasks.Service`.
  - создаёт diagnostic session после успешной генерации задач, чтобы не оставлять пустую session при ошибке генерации.
- `Submit`:
  - проверяет, что `session_id` принадлежит текущему `userID`;
  - проверяет ответы через AI;
  - сохраняет diagnostic answers;
  - обновляет progress;
  - делает AI-анализ слабых тем;
  - завершает diagnostic session.

Если AI key отсутствует, submit возвращает `ai_unavailable`, потому что проверка и анализ зависят от AI.

Как объяснить: diagnostic service собирает полный сценарий диагностики, но использует tasks service для генерации задач.

### `internal/progress/service.go`

Сервис личного кабинета.

Методы:

- `GetProgress` - список прогресса по темам.
- `GetStats` - общая статистика.
- `GetRecommendations` - темы, которые нужно потренировать.

Как объяснить: progress service не считает всё в handler, а отдаёт готовые данные из repository.

### Тестовые файлы

Добавлены:

- `internal/auth/service_test.go`
- `internal/api/router_test.go`
- `internal/tasks/service_test.go`

Что проверяют:

- пароль сохраняется как bcrypt hash;
- duplicate register возвращает conflict;
- invalid login возвращает unauthorized;
- auth rate limit возвращает `rate_limited` после частых запросов с одного IP/path;
- protected route требует Bearer token;
- `/health` возвращает `db_unavailable`, если БД недоступна;
- strict JSON запрещает неизвестные поля;
- `/tasks/check` не принимает `mode`;
- `custom` требует `oge_number` и `subtype_code`;
- генерация возвращает `ai_unavailable`, если AI key отсутствует;
- `/tasks/check` сохраняет attempt с mode, который был записан у задачи;
- JWT с неподдерживаемым `alg` отклоняется;
- auth endpoints принимают пустой `Content-Type`, но запрещают неверный `Content-Type`.
- каждый response получает `X-Request-ID`;
- security headers выставляются на responses;
- auth rate limit можно настраивать через `.env`;
- пароль длиннее 72 байт отклоняется, чтобы не получить silent truncation в bcrypt;
- `OPENROUTER_MODEL` обязателен, если задан `OPENROUTER_API_KEY`;
- `diagnostic/submit` сначала проверяет владельца `session_id`.

## 3. Эндпоинты

### `POST /api/v1/auth/register`

Регистрация пользователя.

Request:

```json
{
  "email": "student@example.com",
  "password": "password123"
}
```

Response:

```json
{
  "success": true,
  "data": {
    "token": "jwt-token",
    "user": {
      "id": 1,
      "email": "student@example.com"
    }
  }
}
```

### `POST /api/v1/auth/login`

Вход пользователя.

Request:

```json
{
  "email": "student@example.com",
  "password": "password123"
}
```

Response такой же, как у регистрации: token + user.

### `GET /api/v1/auth/me`

Защищённый маршрут. Нужен header:

```text
Authorization: Bearer <token>
```

Response:

```json
{
  "success": true,
  "data": {
    "id": 1,
    "email": "student@example.com"
  }
}
```

### `GET /health`

Публичный технический маршрут.

Если PostgreSQL доступен:

```json
{
  "success": true,
  "data": {
    "status": "ok",
    "db": "ok"
  }
}
```

Если PostgreSQL недоступен:

```json
{
  "success": false,
  "error": {
    "code": "db_unavailable",
    "message": "База данных недоступна"
  }
}
```

### `POST /api/v1/tasks/generate`

Генерация задачи.

`mode=custom`:

```json
{
  "mode": "custom",
  "oge_number": 9,
  "subtype_code": "linear_equation"
}
```

`mode=weak`:

```json
{
  "mode": "weak"
}
```

Flow для `mode=weak`:

1. Backend берёт текущего пользователя из JWT.
2. Ищет в `progress` запись с минимальным `mastery_score`.
3. Берёт из неё `oge_number` и `subtype_code`.
4. Пытается восстановить `task_type_id` через `task_types`.
5. Генерирует задачу через AI и сохраняет её как `mode=weak`.

`mode=all`:

```json
{
  "mode": "all"
}
```

Response:

```json
{
  "success": true,
  "data": {
    "id": 1,
    "mode": "custom",
    "oge_number": 9,
    "subtype_code": "linear_equation",
    "question": "Решите уравнение x + 2 = 5.",
    "source": "qwen"
  }
}
```

### `POST /api/v1/tasks/check`

Проверка ответа. `mode` здесь не передаётся.

Request:

```json
{
  "task_id": 1,
  "student_answer": "3"
}
```

Response:

```json
{
  "success": true,
  "data": {
    "is_correct": true,
    "short_feedback": "Верно",
    "reason": "Ответ совпадает с правильным"
  }
}
```

### `POST /api/v1/tasks/hint`

Request:

```json
{
  "task_id": 1
}
```

Response:

```json
{
  "success": true,
  "data": {
    "hint": "Начни с переноса числа 2 в правую часть."
  }
}
```

### `POST /api/v1/tasks/explain`

Request:

```json
{
  "task_id": 1
}
```

Response:

```json
{
  "success": true,
  "data": {
    "explanation": "Чтобы решить уравнение, нужно оставить x слева...",
    "steps": ["x + 2 = 5", "x = 5 - 2", "x = 3"]
  }
}
```

### `POST /api/v1/diagnostic/start`

Создаёт диагностику и возвращает 14 задач.

Response:

```json
{
  "success": true,
  "data": {
    "session_id": 1,
    "tasks": []
  }
}
```

### `POST /api/v1/diagnostic/submit`

Request:

```json
{
  "session_id": 1,
  "answers": [
    {
      "task_id": 1,
      "student_answer": "3"
    }
  ]
}
```

Response:

```json
{
  "success": true,
  "data": {
    "session_id": 1,
    "answers": [],
    "analysis": {
      "summary": "Краткий анализ",
      "weak_topics": ["linear_equation"]
    }
  }
}
```

### `GET /api/v1/progress`

Возвращает прогресс по темам.

### `GET /api/v1/progress/stats`

Возвращает общую статистику.

### `GET /api/v1/progress/recommendations`

Возвращает темы, которые стоит потренировать.

## 4. Как запустить

1. Перейти в backend:

```powershell
cd C:\Users\Nikita\Desktop\bd_oge\backend
```

2. Создать `.env` на основе `.env.example` и заполнить реальные значения:

```text
PORT=8080
TOKEN_TTL=24h
AUTH_RATE_LIMIT_REQUESTS=10
AUTH_RATE_LIMIT_WINDOW=1m
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173
DATABASE_URL=postgres://postgres:<password>@localhost:5432/oge_training_db?sslmode=disable
AUTH_SECRET=<long-random-secret-at-least-32-chars>
OPENROUTER_BASE_URL=https://openrouter.ai/api/v1
OPENROUTER_API_KEY=<openrouter-key>
OPENROUTER_MODEL=qwen/qwen-2.5-72b-instruct
```

Если `OPENROUTER_API_KEY` пока нет, можно оставить его пустым. Тогда AI endpoints будут возвращать `ai_unavailable`. Если ключ задан, `OPENROUTER_MODEL` обязателен.

3. Применить миграции к локальной БД:

```powershell
psql "$env:DATABASE_URL" -v ON_ERROR_STOP=1 -f .\migrations\001_init.sql
psql "$env:DATABASE_URL" -v ON_ERROR_STOP=1 -f .\migrations\002_generated_tasks_ownership.sql
```

4. Запустить сервер:

```powershell
go run .\cmd\server
```

## 5. Как тестировать через curl

В PowerShell лучше использовать `curl.exe`, чтобы не попасть в alias PowerShell.

### Health

```powershell
curl.exe http://localhost:8080/health
```

### Register

```powershell
curl.exe -X POST http://localhost:8080/api/v1/auth/register `
  -H "Content-Type: application/json" `
  -d "{\"email\":\"student@example.com\",\"password\":\"password123\"}"
```

### Login

```powershell
curl.exe -X POST http://localhost:8080/api/v1/auth/login `
  -H "Content-Type: application/json" `
  -d "{\"email\":\"student@example.com\",\"password\":\"password123\"}"
```

Из ответа нужно взять `data.token`.

### Me

```powershell
curl.exe http://localhost:8080/api/v1/auth/me `
  -H "Authorization: Bearer <token>"
```

### Generate custom task

```powershell
curl.exe -X POST http://localhost:8080/api/v1/tasks/generate `
  -H "Authorization: Bearer <token>" `
  -H "Content-Type: application/json" `
  -d "{\"mode\":\"custom\",\"oge_number\":9,\"subtype_code\":\"linear_equation\"}"
```

### Generate weak task

```powershell
curl.exe -X POST http://localhost:8080/api/v1/tasks/generate `
  -H "Authorization: Bearer <token>" `
  -H "Content-Type: application/json" `
  -d "{\"mode\":\"weak\"}"
```

### Check answer

```powershell
curl.exe -X POST http://localhost:8080/api/v1/tasks/check `
  -H "Authorization: Bearer <token>" `
  -H "Content-Type: application/json" `
  -d "{\"task_id\":1,\"student_answer\":\"3\"}"
```

### Hint

```powershell
curl.exe -X POST http://localhost:8080/api/v1/tasks/hint `
  -H "Authorization: Bearer <token>" `
  -H "Content-Type: application/json" `
  -d "{\"task_id\":1}"
```

### Explain

```powershell
curl.exe -X POST http://localhost:8080/api/v1/tasks/explain `
  -H "Authorization: Bearer <token>" `
  -H "Content-Type: application/json" `
  -d "{\"task_id\":1}"
```

### Progress

```powershell
curl.exe http://localhost:8080/api/v1/progress `
  -H "Authorization: Bearer <token>"
```

```powershell
curl.exe http://localhost:8080/api/v1/progress/stats `
  -H "Authorization: Bearer <token>"
```

```powershell
curl.exe http://localhost:8080/api/v1/progress/recommendations `
  -H "Authorization: Bearer <token>"
```

## 6. Как проверить API в Postman

### 6.1. Создать Environment

В Postman создай environment, например `oge-ai-trainer local`.

Добавь переменные:

| Variable | Initial value | Current value |
| --- | --- | --- |
| `base_url` | `http://localhost:8080` | `http://localhost:8080` |
| `token` | пусто | пусто |

Дальше во всех запросах используй `{{base_url}}` вместо полного адреса.

### 6.2. Health

Создай запрос:

```text
GET {{base_url}}/health
```

Ожидаемый успешный ответ:

```json
{
  "success": true,
  "data": {
    "status": "ok",
    "db": "ok"
  }
}
```

Если БД недоступна, будет:

```json
{
  "success": false,
  "error": {
    "code": "db_unavailable",
    "message": "База данных недоступна"
  }
}
```

### 6.3. Register

Создай запрос:

```text
POST {{base_url}}/api/v1/auth/register
```

Headers:

```text
Content-Type: application/json
```

Body -> raw -> JSON:

```json
{
  "email": "student@example.com",
  "password": "password123"
}
```

Во вкладке `Tests` добавь скрипт, чтобы Postman сам сохранил token:

```javascript
const body = pm.response.json();
if (body.success && body.data && body.data.token) {
  pm.environment.set("token", body.data.token);
}
```

После запроса переменная `token` заполнится автоматически.

### 6.4. Login

Создай запрос:

```text
POST {{base_url}}/api/v1/auth/login
```

Body такой же:

```json
{
  "email": "student@example.com",
  "password": "password123"
}
```

Во вкладке `Tests` можно использовать тот же скрипт сохранения token.

### 6.5. Проверка protected route

Создай запрос:

```text
GET {{base_url}}/api/v1/auth/me
```

Authorization:

- Type: `Bearer Token`
- Token: `{{token}}`

Ожидаемый ответ:

```json
{
  "success": true,
  "data": {
    "id": 1,
    "email": "student@example.com"
  }
}
```

Если token не передать, будет `unauthorized`.

### 6.6. Generate task

Создай запрос:

```text
POST {{base_url}}/api/v1/tasks/generate
```

Authorization:

- Type: `Bearer Token`
- Token: `{{token}}`

Headers:

```text
Content-Type: application/json
```

Body для custom:

```json
{
  "mode": "custom",
  "oge_number": 9,
  "subtype_code": "linear_equation"
}
```

Body для weak:

```json
{
  "mode": "weak"
}
```

Body для all:

```json
{
  "mode": "all"
}
```

Если `OPENROUTER_API_KEY` не задан, ожидаемый ответ:

```json
{
  "success": false,
  "error": {
    "code": "ai_unavailable",
    "message": "AI-сервис недоступен"
  }
}
```

### 6.7. Check answer

Создай запрос:

```text
POST {{base_url}}/api/v1/tasks/check
```

Authorization:

- Type: `Bearer Token`
- Token: `{{token}}`

Body:

```json
{
  "task_id": 1,
  "student_answer": "3"
}
```

Важно: `mode` сюда не передаётся. Backend берёт режим из сохранённой задачи.

### 6.8. Hint и Explain

Hint:

```text
POST {{base_url}}/api/v1/tasks/hint
```

Explain:

```text
POST {{base_url}}/api/v1/tasks/explain
```

Body одинаковый:

```json
{
  "task_id": 1
}
```

### 6.9. Progress

Создай три GET-запроса с Bearer token:

```text
GET {{base_url}}/api/v1/progress
GET {{base_url}}/api/v1/progress/stats
GET {{base_url}}/api/v1/progress/recommendations
```

### 6.10. Что смотреть в Postman

- Status code: `200`, `201`, `400`, `401`, `404`, `409`, `429`, `503`.
- Header `X-Request-ID`: есть в каждом ответе.
- Security headers: `X-Content-Type-Options`, `X-Frame-Options`, `Referrer-Policy`.
- Единый JSON-формат: всегда `success/data` или `success/error`.
- Для protected routes всегда должен быть Bearer token.

## 7. Проверки, которые выполнены

Выполнено форматирование:

```powershell
gofmt -w <изменённые Go-файлы>
```

Выполнены тесты:

```powershell
go test ./...
```

Проверенные сценарии:

- регистрация хэширует пароль через bcrypt;
- повторная регистрация email возвращает `conflict`;
- неверный пароль возвращает `unauthorized`;
- auth rate limit возвращает `rate_limited` после частых запросов с одного IP/path;
- protected route без token возвращает `unauthorized`;
- `/health` возвращает success при доступной БД и `db_unavailable` при ошибке;
- strict JSON запрещает неизвестные поля;
- `/tasks/check` не принимает `mode`;
- `mode=custom` требует `oge_number` и `subtype_code`;
- генерация задачи возвращает `ai_unavailable`, если AI key отсутствует;
- `/tasks/check` сохраняет attempt с mode, который был записан у задачи;
- JWT с неподдерживаемым `alg` отклоняется;
- auth endpoints принимают пустой `Content-Type`, но запрещают неверный `Content-Type`.
- каждый response получает `X-Request-ID`;
- security headers выставляются на responses;
- auth rate limit можно настраивать через `.env`;
- пароль длиннее 72 байт отклоняется;
- `OPENROUTER_MODEL` обязателен при заданном `OPENROUTER_API_KEY`;
- `diagnostic/submit` проверяет владельца `session_id` до AI-вызовов.

## 8. Что важно сказать при объяснении работы

1. Все пользовательские маршруты находятся под `/api/v1`, кроме технического `/health`.
2. Все ответы имеют одинаковую оболочку `success/data` или `success/error`.
3. `userID` не приходит от frontend, а берётся только из JWT.
4. Пароль не хранится в БД как текст, только bcrypt hash в `users.password_hash`.
5. `AUTH_SECRET` и `OPENROUTER_API_KEY` лежат только в `.env`.
6. Если OpenRouter key отсутствует, fake AI responses не создаются.
7. Для генерации задачи при недоступном AI backend возвращается `ai_unavailable`.
8. JWT проверяется стандартной библиотекой и только с `HS256`.
9. Auth endpoints защищены настраиваемым rate limiting.
10. `X-Request-ID` помогает искать конкретный запрос в логах.
11. Security headers добавляются централизованно middleware-слоем.
12. Внутренние ошибки Go, SQL и OpenRouter не показываются пользователю напрямую.
