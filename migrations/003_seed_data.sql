BEGIN;

-- Добавляем одну тему (Алгебра)
INSERT INTO topics (code, title, description) VALUES
('algebra', 'Алгебра', 'Алгебраические выражения, уравнения, неравенства');

-- Добавляем подтему (Линейные уравнения)
INSERT INTO subtopics (topic_id, code, title, description) VALUES
(1, 'linear_eq', 'Линейные уравнения', 'Уравнения вида ax + b = c');

-- Добавляем тип задания (№7 ОГЭ - линейные уравнения)
INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description) VALUES
(7, 1, 1, 'linear_eq', 'Линейные уравнения', 'Решите линейное уравнение');

COMMIT;
