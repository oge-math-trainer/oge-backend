BEGIN;

-- Добавляем темы (если нет)
INSERT INTO topics (code, title, description) VALUES
('algebra', 'Алгебра', 'Числа, вычисления, алгебраические выражения, уравнения, неравенства, функции'),
('geometry', 'Геометрия', 'Планиметрия: треугольники, окружности, четырёхугольники, многоугольники')
ON CONFLICT (code) DO NOTHING;

-- Задание 6: Числа
INSERT INTO subtopics (topic_id, code, title, description)
SELECT t.id, 'numbers_fractions', 'Обыкновенные дроби', 'Действия с обыкновенными дробями'
FROM topics t WHERE t.code = 'algebra'
ON CONFLICT (topic_id, code) DO NOTHING;

INSERT INTO subtopics (topic_id, code, title, description)
SELECT t.id, 'numbers_decimal', 'Десятичные дроби', 'Действия с десятичными дробями'
FROM topics t WHERE t.code = 'algebra'
ON CONFLICT (topic_id, code) DO NOTHING;

INSERT INTO subtopics (topic_id, code, title, description)
SELECT t.id, 'numbers_integers', 'Целые числа', 'Целые и рациональные числа'
FROM topics t WHERE t.code = 'algebra'
ON CONFLICT (topic_id, code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 6, t.id, s.id, 'numbers_fractions', 'Обыкновенные дроби', 'Действия с обыкновенными дробями'
FROM topics t
JOIN subtopics s ON s.topic_id = t.id AND s.code = 'numbers_fractions'
WHERE t.code = 'algebra'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 6, t.id, s.id, 'numbers_decimal', 'Десятичные дроби', 'Действия с десятичными дробями'
FROM topics t
JOIN subtopics s ON s.topic_id = t.id AND s.code = 'numbers_decimal'
WHERE t.code = 'algebra'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 6, t.id, s.id, 'numbers_integers', 'Целые числа', 'Целые и рациональные числа'
FROM topics t
JOIN subtopics s ON s.topic_id = t.id AND s.code = 'numbers_integers'
WHERE t.code = 'algebra'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

-- Задание 7: Числовая прямая
INSERT INTO subtopics (topic_id, code, title, description)
SELECT t.id, 'numberline_compare', 'Сравнение чисел', 'Расположение на числовой прямой'
FROM topics t WHERE t.code = 'algebra'
ON CONFLICT (topic_id, code) DO NOTHING;

INSERT INTO subtopics (topic_id, code, title, description)
SELECT t.id, 'inequalities_simple', 'Простейшие неравенства', 'Сравнение выражений'
FROM topics t WHERE t.code = 'algebra'
ON CONFLICT (topic_id, code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 7, t.id, s.id, 'numberline_compare', 'Числовая прямая', 'Сравнение чисел'
FROM topics t
JOIN subtopics s ON s.topic_id = t.id AND s.code = 'numberline_compare'
WHERE t.code = 'algebra'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 7, t.id, s.id, 'inequalities_simple', 'Простейшие неравенства', 'Сравнение выражений'
FROM topics t
JOIN subtopics s ON s.topic_id = t.id AND s.code = 'inequalities_simple'
WHERE t.code = 'algebra'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

-- Задание 8: Алгебраические выражения
INSERT INTO subtopics (topic_id, code, title, description)
SELECT t.id, 'algebra_powers', 'Степени', 'Степени с целым показателем'
FROM topics t WHERE t.code = 'algebra'
ON CONFLICT (topic_id, code) DO NOTHING;

INSERT INTO subtopics (topic_id, code, title, description)
SELECT t.id, 'algebra_roots', 'Корни', 'Арифметические корни'
FROM topics t WHERE t.code = 'algebra'
ON CONFLICT (topic_id, code) DO NOTHING;

INSERT INTO subtopics (topic_id, code, title, description)
SELECT t.id, 'algebra_formulas', 'Формулы', 'Формулы сокращённого умножения'
FROM topics t WHERE t.code = 'algebra'
ON CONFLICT (topic_id, code) DO NOTHING;

INSERT INTO subtopics (topic_id, code, title, description)
SELECT t.id, 'algebra_fractions', 'Алгебраические дроби', 'Преобразование дробей'
FROM topics t WHERE t.code = 'algebra'
ON CONFLICT (topic_id, code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 8, t.id, s.id, 'algebra_powers', 'Степени', 'Степени'
FROM topics t
JOIN subtopics s ON s.topic_id = t.id AND s.code = 'algebra_powers'
WHERE t.code = 'algebra'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 8, t.id, s.id, 'algebra_roots', 'Корни', 'Корни'
FROM topics t
JOIN subtopics s ON s.topic_id = t.id AND s.code = 'algebra_roots'
WHERE t.code = 'algebra'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 8, t.id, s.id, 'algebra_formulas', 'Формулы', 'Формулы'
FROM topics t
JOIN subtopics s ON s.topic_id = t.id AND s.code = 'algebra_formulas'
WHERE t.code = 'algebra'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 8, t.id, s.id, 'algebra_fractions', 'Алгебраические дроби', 'Дроби'
FROM topics t
JOIN subtopics s ON s.topic_id = t.id AND s.code = 'algebra_fractions'
WHERE t.code = 'algebra'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

-- Задание 9: Уравнения
INSERT INTO subtopics (topic_id, code, title, description)
SELECT t.id, 'equations_linear', 'Линейные уравнения', 'ax + b = 0'
FROM topics t WHERE t.code = 'algebra'
ON CONFLICT (topic_id, code) DO NOTHING;

INSERT INTO subtopics (topic_id, code, title, description)
SELECT t.id, 'equations_quadratic', 'Квадратные уравнения', 'Дискриминант'
FROM topics t WHERE t.code = 'algebra'
ON CONFLICT (topic_id, code) DO NOTHING;

INSERT INTO subtopics (topic_id, code, title, description)
SELECT t.id, 'equations_rational', 'Дробно-рациональные', 'Рациональные уравнения'
FROM topics t WHERE t.code = 'algebra'
ON CONFLICT (topic_id, code) DO NOTHING;

INSERT INTO subtopics (topic_id, code, title, description)
SELECT t.id, 'equations_system_linear', 'Системы линейных', 'Метод подстановки'
FROM topics t WHERE t.code = 'algebra'
ON CONFLICT (topic_id, code) DO NOTHING;

INSERT INTO subtopics (topic_id, code, title, description)
SELECT t.id, 'equations_system_mixed', 'Смешанные системы', 'Линейное + квадратное'
FROM topics t WHERE t.code = 'algebra'
ON CONFLICT (topic_id, code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 9, t.id, s.id, 'equations_linear', 'Линейные уравнения', 'Линейные'
FROM topics t
JOIN subtopics s ON s.topic_id = t.id AND s.code = 'equations_linear'
WHERE t.code = 'algebra'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 9, t.id, s.id, 'equations_quadratic', 'Квадратные уравнения', 'Квадратные'
FROM topics t
JOIN subtopics s ON s.topic_id = t.id AND s.code = 'equations_quadratic'
WHERE t.code = 'algebra'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 9, t.id, s.id, 'equations_rational', 'Дробно-рациональные', 'Рациональные'
FROM topics t
JOIN subtopics s ON s.topic_id = t.id AND s.code = 'equations_rational'
WHERE t.code = 'algebra'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 9, t.id, s.id, 'equations_system_linear', 'Системы линейных', 'Системы'
FROM topics t
JOIN subtopics s ON s.topic_id = t.id AND s.code = 'equations_system_linear'
WHERE t.code = 'algebra'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 9, t.id, s.id, 'equations_system_mixed', 'Смешанные системы', 'Смешанные'
FROM topics t
JOIN subtopics s ON s.topic_id = t.id AND s.code = 'equations_system_mixed'
WHERE t.code = 'algebra'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

COMMIT;
