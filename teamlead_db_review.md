## Что сделано

Создана и применена PostgreSQL-схема по документу `🔑 Суть MVP.docx`.

В базе созданы таблицы:

- `users`
- `topics`
- `subtopics`
- `task_types`
- `generated_tasks`
- `fallback_tasks`
- `diagnostic_sessions`
- `diagnostic_answers`
- `attempts`
- `progress`
- `ai_logs`

Итоговая схема лежит в одном файле:

```text
migrations/001_init.sql
```

## Логика тем и тренировок

После диагностики ученик выбирает режим тренировки:

- `weak` - слабые темы. Backend берёт записи из `progress` с низким `mastery_score` и через `task_type_id` выходит на `task_types`, `subtopics`, `topics`.
- `all` - по всем заданиям. Backend выбирает типы заданий из `task_types` по номерам ОГЭ 6-19.
- `custom` - ручной выбор. Frontend строит выпадающий список: `topics` -> `subtopics` -> `task_types`.

Связь справочников:

```text
topics
  -> subtopics
      -> task_types
          -> generated_tasks / fallback_tasks / progress
```

Важно: `task_types` - это не сами задания, а классификатор типов заданий: номер ОГЭ + тема + подтема + код подтипа. Сами задачи лежат в `generated_tasks` и `fallback_tasks`.

## Ключевые решения

- Регистрация пользователя рассчитана на email + пароль.
- В таблице `users` хранится не пароль, а `password_hash`.
- Хэширование пароля должно выполняться на backend-стороне, например через Argon2id или bcrypt.
- `topics`, `subtopics`, `task_types` созданы как справочники, но пока не заполнены, потому что нужен утверждённый список тем, подтем и типов заданий.
- `fallback_tasks` создана, но не заполнена, потому что нужен набор заранее проверенных ручных задач.
- Для AI-ответов и служебных данных используются поля `JSONB`.

## Основные ограничения

- `oge_number` разрешён только от `6` до `19`.
- `attempts.mode` разрешён только: `weak`, `all`, `custom`, `diagnostic`.
- `progress.mastery_score` разрешён только от `0` до `1`.
- `progress.correct_count` не может быть больше `progress.attempts_count`.
- Email пользователя уникален через индекс по `lower(email)`.
- `password_hash` обязателен и не может быть пустым.
- `subtopics` всегда привязана к `topics`.
- `task_types` всегда привязана к конкретной теме и подтеме.

## Как проверить локально

PowerShell:

```powershell
cd C:\Users\Nikita\Desktop\bd_oge

$env:PGPASSWORD = '<пароль PostgreSQL>'

& 'C:\Program Files\PostgreSQL\18\bin\psql.exe' -h localhost -p 5432 -U postgres -d oge_training_db -c "SELECT current_database(), current_user, current_schema();"

& 'C:\Program Files\PostgreSQL\18\bin\psql.exe' -h localhost -p 5432 -U postgres -d oge_training_db -c "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' ORDER BY table_name;"

& 'C:\Program Files\PostgreSQL\18\bin\psql.exe' -h localhost -p 5432 -U postgres -d oge_training_db -c "SELECT column_name, data_type, is_nullable FROM information_schema.columns WHERE table_schema = 'public' AND table_name IN ('topics', 'subtopics', 'task_types', 'users') ORDER BY table_name, ordinal_position;"

Remove-Item Env:PGPASSWORD
```

Ожидаемые таблицы:

```text
ai_logs
attempts
diagnostic_answers
diagnostic_sessions
fallback_tasks
generated_tasks
progress
subtopics
task_types
topics
users
```

Ожидаемые поля `users`:

```text
id
email
password_hash
created_at
```

## Проверка в pgAdmin

Важно открыть Query Tool именно у базы `oge_training_db`.

Проверить текущую базу:

```sql
SELECT current_database(), current_user, current_schema();
```

Должно быть:

```text
oge_training_db | postgres | public
```

Проверить таблицы:

```sql
SELECT table_name
FROM information_schema.tables
WHERE table_schema = 'public'
ORDER BY table_name;
```

Проверить справочники:

```sql
SELECT * FROM public.topics;
SELECT * FROM public.subtopics;
SELECT * FROM public.task_types;
```

Проверить ручной пул задач:

```sql
SELECT * FROM public.fallback_tasks;
```

Если pgAdmin пишет, что таблица не существует, почти всегда Query Tool открыт не в базе `oge_training_db`.

## Что намеренно не заполнено

- `topics` пустая, потому что нужен утверждённый список тем.
- `subtopics` пустая, потому что нужен утверждённый список подтем.
- `task_types` пустая, потому что нужен список связок: номер ОГЭ + тема + подтема + код подтипа.
- `fallback_tasks` пустая, потому что нужен набор заранее проверенных ручных задач.
- `users` пустая, потому что регистрация будет выполняться через backend.
- `generated_tasks`, `attempts`, `progress`, `ai_logs` будут заполняться приложением во время работы.

## Что можно улучшить дальше

- Создать отдельного PostgreSQL-пользователя для приложения вместо использования `postgres`.
- Добавить `.env.example` с `DATABASE_URL`, без реального пароля.
- Подготовить seed-файл для `topics`, `subtopics`, `task_types`.
- Подготовить seed-файл для `fallback_tasks`, когда появятся ручные задачи.
- Добавить миграционный инструмент на backend-стороне, например `golang-migrate`.
- Добавить триггер для автоматического обновления `progress.updated_at`.

## Статус

Для задачи "сделать базу данных" схема готова и применена локально.

Нужна проверка тимлида:

- хватает ли текущего набора таблиц для диагностики и трёх режимов тренировки;
- устраивает ли связь `topics` -> `subtopics` -> `task_types`;
- нужен ли seed справочников уже сейчас;
- устраивают ли поля `JSONB` для AI-ответов;
- нужен ли отдельный пользователь БД для backend.
