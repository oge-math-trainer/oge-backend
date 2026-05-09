BEGIN;

-- Очищаем старые данные (если есть)
TRUNCATE TABLE task_types CASCADE;
TRUNCATE TABLE subtopics CASCADE;
TRUNCATE TABLE topics CASCADE;

-- Добавляем темы
INSERT INTO topics (code, title, description) VALUES
('algebra', 'Алгебра', 'Числа, вычисления, алгебраические выражения'),
('geometry', 'Геометрия', 'Планиметрия');

-- Задание 6: Числа
INSERT INTO subtopics (topic_id, code, title, description) VALUES
(1, 'numbers_fractions', 'Обыкновенные дроби', 'Действия с дробями'),
(1, 'numbers_decimal', 'Десятичные дроби', 'Десятичные дроби'),
(1, 'numbers_integers', 'Целые числа', 'Целые и рациональные');

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description) VALUES
(6, 1, 1, 'numbers_fractions', 'Обыкновенные дроби', 'Действия с обыкновенными дробями'),
(6, 1, 2, 'numbers_decimal', 'Десятичные дроби', 'Действия с десятичными дробями'),
(6, 1, 3, 'numbers_integers', 'Целые числа', 'Целые и рациональные числа');

-- Задание 7: Числовая прямая
INSERT INTO subtopics (topic_id, code, title, description) VALUES
(1, 'numberline_compare', 'Сравнение чисел', 'Расположение на прямой'),
(1, 'inequalities_simple', 'Простейшие неравенства', 'Сравнение выражений');

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description) VALUES
(7, 1, 4, 'numberline_compare', 'Числовая прямая', 'Сравнение чисел'),
(7, 1, 5, 'inequalities_simple', 'Простейшие неравенства', 'Сравнение выражений');

-- Задание 8: Алгебраические выражения
INSERT INTO subtopics (topic_id, code, title, description) VALUES
(1, 'algebra_powers', 'Степени', 'Степени с целым показателем'),
(1, 'algebra_roots', 'Корни', 'Арифметические корни'),
(1, 'algebra_formulas', 'Формулы', 'Формулы сокращённого умножения'),
(1, 'algebra_fractions', 'Алгебраические дроби', 'Преобразование дробей');

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description) VALUES
(8, 1, 6, 'algebra_powers', 'Степени', 'Степени'),
(8, 1, 7, 'algebra_roots', 'Корни', 'Корни'),
(8, 1, 8, 'algebra_formulas', 'Формулы', 'Формулы'),
(8, 1, 9, 'algebra_fractions', 'Алгебраические дроби', 'Дроби');

-- Задание 9: Уравнения
INSERT INTO subtopics (topic_id, code, title, description) VALUES
(1, 'equations_linear', 'Линейные уравнения', 'ax + b = 0'),
(1, 'equations_quadratic', 'Квадратные уравнения', 'Дискриминант'),
(1, 'equations_rational', 'Дробно-рациональные', 'Рациональные'),
(1, 'equations_system_linear', 'Системы линейных', 'Метод подстановки'),
(1, 'equations_system_mixed', 'Смешанные системы', 'Линейное + квадратное');

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description) VALUES
(9, 1, 10, 'equations_linear', 'Линейные уравнения', 'Линейные'),
(9, 1, 11, 'equations_quadratic', 'Квадратные уравнения', 'Квадратные'),
(9, 1, 12, 'equations_rational', 'Дробно-рациональные', 'Рациональные'),
(9, 1, 13, 'equations_system_linear', 'Системы линейных', 'Системы'),
(9, 1, 14, 'equations_system_mixed', 'Смешанные системы', 'Смешанные');

-- Задание 10: Вероятность
INSERT INTO subtopics (topic_id, code, title, description) VALUES
(1, 'probability_classic', 'Классическая вероятность', 'Монеты, кубики'),
(1, 'probability_tree', 'Дерево вероятностей', 'Многошаговые');

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description) VALUES
(10, 1, 15, 'probability_classic', 'Классическая вероятность', 'Вероятность'),
(10, 1, 16, 'probability_tree', 'Дерево вероятностей', 'Дерево');

-- Задание 11: Графики
INSERT INTO subtopics (topic_id, code, title, description) VALUES
(1, 'graphs_linear', 'Линейная функция', 'y = kx + b'),
(1, 'graphs_quadratic', 'Квадратичная', 'Парабола'),
(1, 'graphs_inverse', 'Обратная пропорциональность', 'y = k/x'),
(1, 'graphs_transform', 'Преобразования', 'Сдвиги'),
(1, 'graphs_match', 'Соответствие', 'График и формула');

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description) VALUES
(11, 1, 17, 'graphs_linear', 'Линейная функция', 'Линейная'),
(11, 1, 18, 'graphs_quadratic', 'Квадратичная', 'Парабола'),
(11, 1, 19, 'graphs_inverse', 'Обратная пропорциональность', 'Обратная'),
(11, 1, 20, 'graphs_transform', 'Преобразования', 'Сдвиги'),
(11, 1, 21, 'graphs_match', 'Соответствие', 'Соответствие');

-- Задание 12: Формулы
INSERT INTO subtopics (topic_id, code, title, description) VALUES
(1, 'formulas_subst', 'Подстановка', 'Вычисление'),
(1, 'formulas_units', 'Единицы измерения', 'Перевод'),
(1, 'formulas_practical', 'Практическая', 'Текстовая');

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description) VALUES
(12, 1, 22, 'formulas_subst', 'Подстановка', 'Подстановка'),
(12, 1, 23, 'formulas_units', 'Единицы измерения', 'Перевод'),
(12, 1, 24, 'formulas_practical', 'Практическая', 'Практическая');

-- Задание 13: Неравенства
INSERT INTO subtopics (topic_id, code, title, description) VALUES
(1, 'ineq_linear', 'Линейные', 'ax + b > 0'),
(1, 'ineq_quadratic', 'Квадратные', 'Метод интервалов'),
(1, 'ineq_system', 'Системы', 'Пересечение'),
(1, 'ineq_rational', 'Дробные', 'Дробные');

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description) VALUES
(13, 1, 25, 'ineq_linear', 'Линейные', 'Линейные'),
(13, 1, 26, 'ineq_quadratic', 'Квадратные', 'Квадратные'),
(13, 1, 27, 'ineq_system', 'Системы', 'Системы'),
(13, 1, 28, 'ineq_rational', 'Дробные', 'Дробные');

-- Задание 14: Прогрессии
INSERT INTO subtopics (topic_id, code, title, description) VALUES
(1, 'progression_arithm', 'Арифметическая', 'n-й член'),
(1, 'progression_geom', 'Геометрическая', 'Знаменатель');

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description) VALUES
(14, 1, 29, 'progression_arithm', 'Арифметическая', 'Арифметическая'),
(14, 1, 30, 'progression_geom', 'Геометрическая', 'Геометрическая');

-- Задание 15: Треугольники (Геометрия)
INSERT INTO subtopics (topic_id, code, title, description) VALUES
(2, 'triangles_angles', 'Углы', 'Сумма углов'),
(2, 'triangles_pythagor', 'Пифагор', 'Теорема'),
(2, 'triangles_area', 'Площадь', 'Формулы'),
(2, 'triangles_lines', 'Замечательные линии', 'Медианы');

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description) VALUES
(15, 2, 31, 'triangles_angles', 'Углы', 'Углы'),
(15, 2, 32, 'triangles_pythagor', 'Пифагор', 'Пифагор'),
(15, 2, 33, 'triangles_area', 'Площадь', 'Площадь'),
(15, 2, 34, 'triangles_lines', 'Замечательные линии', 'Линии');

-- Задание 16: Окружность
INSERT INTO subtopics (topic_id, code, title, description) VALUES
(2, 'circle_elements', 'Элементы', 'Радиус, диаметр'),
(2, 'circle_angles', 'Углы', 'Вписанные'),
(2, 'circle_tangents', 'Касательные', 'Свойства'),
(2, 'circle_inscribed', 'Вписанные окружности', 'Около фигур'),
(2, 'circle_sector', 'Сектор', 'Площадь');

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description) VALUES
(16, 2, 35, 'circle_elements', 'Элементы', 'Элементы'),
(16, 2, 36, 'circle_angles', 'Углы', 'Углы'),
(16, 2, 37, 'circle_tangents', 'Касательные', 'Касательные'),
(16, 2, 38, 'circle_inscribed', 'Вписанные', 'Вписанные'),
(16, 2, 39, 'circle_sector', 'Сектор', 'Сектор');

-- Задание 17: Четырёхугольники
INSERT INTO subtopics (topic_id, code, title, description) VALUES
(2, 'quad_properties', 'Свойства', 'Параллелограмм'),
(2, 'quad_trapezoid', 'Трапеция', 'Равнобедренная'),
(2, 'quad_area', 'Площади', 'Формулы'),
(2, 'quad_diagonals', 'Диагонали', 'Свойства');

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description) VALUES
(17, 2, 40, 'quad_properties', 'Свойства', 'Свойства'),
(17, 2, 41, 'quad_trapezoid', 'Трапеция', 'Трапеция'),
(17, 2, 42, 'quad_area', 'Площади', 'Площади'),
(17, 2, 43, 'quad_diagonals', 'Диагонали', 'Диагонали');

-- Задание 18: Решётка
INSERT INTO subtopics (topic_id, code, title, description) VALUES
(2, 'grid_distance', 'Расстояние', 'Координаты'),
(2, 'grid_pythagor', 'Пифагор', 'По клеткам'),
(2, 'grid_area', 'Площадь', 'Метод Пика'),
(2, 'grid_midline', 'Средняя линия', 'Свойства');

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description) VALUES
(18, 2, 44, 'grid_distance', 'Расстояние', 'Расстояние'),
(18, 2, 45, 'grid_pythagor', 'Пифагор', 'Пифагор'),
(18, 2, 46, 'grid_area', 'Площадь', 'Площадь'),
(18, 2, 47, 'grid_midline', 'Средняя линия', 'Линия');

-- Задание 19: Утверждения
INSERT INTO subtopics (topic_id, code, title, description) VALUES
(2, 'logic_angles', 'Углы', 'Вертикальные'),
(2, 'logic_triangles', 'Треугольники', 'Равенство'),
(2, 'logic_quad', 'Четырёхугольники', 'Свойства'),
(2, 'logic_circle', 'Окружность', 'Касательные');

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description) VALUES
(19, 2, 48, 'logic_angles', 'Углы', 'Углы'),
(19, 2, 49, 'logic_triangles', 'Треугольники', 'Треугольники'),
(19, 2, 50, 'logic_quad', 'Четырёхугольники', 'Четырёхугольники'),
(19, 2, 51, 'logic_circle', 'Окружность', 'Окружность');

COMMIT;
