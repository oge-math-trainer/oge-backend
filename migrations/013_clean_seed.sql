BEGIN;

-- ============================================
-- TOPICS
-- ============================================
INSERT INTO topics (code, title, description) VALUES
('algebra', 'Алгебра', 'Числа, вычисления, алгебраические выражения, уравнения, неравенства, функции'),
('geometry', 'Геометрия', 'Планиметрия: треугольники, окружности, четырёхугольники, многоугольники')
ON CONFLICT (code) DO NOTHING;

-- ============================================
-- SUBTOPICS — Алгебра
-- ============================================
INSERT INTO subtopics (topic_id, code, title, description) VALUES
((SELECT id FROM topics WHERE code='algebra'), 'fractions_common', 'Действия с обыкновенными дробями', 'Сложение, вычитание, умножение, деление обыкновенных и смешанных дробей'),
((SELECT id FROM topics WHERE code='algebra'), 'fractions_decimal', 'Действия с десятичными дробями', 'Арифметические операции с десятичными дробями'),
((SELECT id FROM topics WHERE code='algebra'), 'fractions_mixed', 'Действия с обыкновенными и десятичными дробями', 'Совместные вычисления, порядок действий'),
((SELECT id FROM topics WHERE code='algebra'), 'compare_numbers', 'Сравнение чисел', 'Какое число заключено между двумя числами, или какому промежутку принадлежит'),
((SELECT id FROM topics WHERE code='algebra'), 'number_line', 'Числа на координатной прямой', 'Определение положения чисел на координатной прямой'),
((SELECT id FROM topics WHERE code='algebra'), 'simple_inequalities', 'Простейшие неравенства', 'Дано множество чисел, найти принадлежащее ему'),
((SELECT id FROM topics WHERE code='algebra'), 'powers', 'Степени с целым показателем', 'Свойства степеней, действия со степенями'),
((SELECT id FROM topics WHERE code='algebra'), 'roots', 'Арифметические корни', 'Квадратные корни, вынесение/внесение множителя'),
((SELECT id FROM topics WHERE code='algebra'), 'formulas_short', 'Формулы сокращённого умножения', 'Квадрат суммы/разности, разность квадратов'),
((SELECT id FROM topics WHERE code='algebra'), 'eq_linear', 'Линейные уравнения', 'Уравнения первой степени'),
((SELECT id FROM topics WHERE code='algebra'), 'eq_quadratic', 'Квадратные уравнения', 'Полные и неполные квадратные уравнения'),
((SELECT id FROM topics WHERE code='algebra'), 'eq_rational', 'Рациональные уравнения', 'Уравнения с переменной в знаменателе'),
((SELECT id FROM topics WHERE code='algebra'), 'eq_systems_linear', 'Системы линейных уравнений', 'Метод подстановки, метод сложения'),
((SELECT id FROM topics WHERE code='algebra'), 'prob_classic', 'Классические вероятности', 'Формула P = m/n'),
((SELECT id FROM topics WHERE code='algebra'), 'prob_theorems', 'Теоремы о вероятностях', 'Вероятность противоположного события'),
((SELECT id FROM topics WHERE code='algebra'), 'graph_kb', 'Знаки коэффициентов k и b', 'Определение знаков по графику линейной функции'),
((SELECT id FROM topics WHERE code='algebra'), 'func_graph_match', 'Соответствие функций и графиков', 'Установление соответствия между функцией и графиком'),
((SELECT id FROM topics WHERE code='algebra'), 'graph_formula_match', 'Соответствие графиков и формул', 'Определение формулы по графику'),
((SELECT id FROM topics WHERE code='algebra'), 'formula_calc', 'Вычисление по формуле', 'Подстановка значений в формулу'),
((SELECT id FROM topics WHERE code='algebra'), 'linear_eq_practical', 'Линейные уравнения в практических задачах', 'Текстовые задачи на движение, работу, проценты'),
((SELECT id FROM topics WHERE code='algebra'), 'ineq_linear', 'Линейные неравенства', 'Неравенства первой степени с выбором ответа'),
((SELECT id FROM topics WHERE code='algebra'), 'ineq_quadratic', 'Квадратные неравенства', 'Метод интервалов'),
((SELECT id FROM topics WHERE code='algebra'), 'ineq_system', 'Системы неравенств', 'Решение систем неравенств'),
((SELECT id FROM topics WHERE code='algebra'), 'prog_arith', 'Арифметическая прогрессия', 'Формула n-го члена, сумма первых n членов'),
((SELECT id FROM topics WHERE code='algebra'), 'prog_geom', 'Геометрическая прогрессия', 'Формула n-го члена, знаменатель, сумма')
ON CONFLICT (topic_id, code) DO NOTHING;

-- ============================================
-- SUBTOPICS — Геометрия
-- ============================================
INSERT INTO subtopics (topic_id, code, title, description) VALUES
((SELECT id FROM topics WHERE code='geometry'), 'tri_angles', 'Углы треугольника', 'Сумма углов, внешний угол, равнобедренный треугольник'),
((SELECT id FROM topics WHERE code='geometry'), 'tri_general', 'Треугольники общего вида', 'Соотношения сторон и углов, медианы, биссектрисы'),
((SELECT id FROM topics WHERE code='geometry'), 'tri_isosceles', 'Равнобедренные треугольники', 'Свойства углов при основании'),
((SELECT id FROM topics WHERE code='geometry'), 'tri_right', 'Прямоугольный треугольник', 'Теорема Пифагора, тригонометрические отношения'),
((SELECT id FROM topics WHERE code='geometry'), 'circle_angles', 'Центральные и вписанные углы', 'Градусные меры дуг, связь центрального и вписанного углов'),
((SELECT id FROM topics WHERE code='geometry'), 'circle_inscribed', 'Окружность вписанная в многоугольник', 'Радиус вписанной окружности'),
((SELECT id FROM topics WHERE code='geometry'), 'circle_circumscribed', 'Окружность описанная вокруг многоугольника', 'Радиус описанной окружности'),
((SELECT id FROM topics WHERE code='geometry'), 'circle_tangent_chord', 'Касательная хорда секущая радиус', 'Свойства касательных, равенство отрезков'),
((SELECT id FROM topics WHERE code='geometry'), 'quad_trapezoid', 'Трапеция', 'Средняя линия, углы при основании, площадь'),
((SELECT id FROM topics WHERE code='geometry'), 'quad_parallelogram', 'Параллелограмм', 'Противоположные стороны и углы, диагонали'),
((SELECT id FROM topics WHERE code='geometry'), 'quad_rectangle', 'Прямоугольник', 'Свойства углов и диагоналей, площадь'),
((SELECT id FROM topics WHERE code='geometry'), 'quad_rhombus', 'Ромб', 'Равенство сторон, перпендикулярность диагоналей'),
((SELECT id FROM topics WHERE code='geometry'), 'geo_distance', 'Расстояние между точками', 'Формула расстояния, координатный метод'),
((SELECT id FROM topics WHERE code='geometry'), 'geo_midline', 'Средняя линия', 'Средняя линия треугольника и трапеции'),
((SELECT id FROM topics WHERE code='geometry'), 'geo_pythagor', 'Теорема Пифагора на клетчатой бумаге', 'Длины наклонных отрезков'),
((SELECT id FROM topics WHERE code='geometry'), 'geo_area', 'Площади на клетчатой бумаге', 'Метод Пика, разбиение на фигуры'),
((SELECT id FROM topics WHERE code='geometry'), 'geo_side_length', 'Длины сторон', 'Нахождение длин сторон по координатам'),
((SELECT id FROM topics WHERE code='geometry'), 'logic_angles_lines', 'Утверждения об углах и прямых', 'Вертикальные, смежные, накрест лежащие углы'),
((SELECT id FROM topics WHERE code='geometry'), 'logic_triangles', 'Утверждения о треугольниках', 'Признаки равенства/подобия, теоремы'),
((SELECT id FROM topics WHERE code='geometry'), 'logic_quads', 'Утверждения о четырёхугольниках', 'Свойства параллелограмма, ромба, прямоугольника'),
((SELECT id FROM topics WHERE code='geometry'), 'logic_circle', 'Утверждения об окружности', 'Вписанные/центральные углы, касательные')
ON CONFLICT (topic_id, code) DO NOTHING;

-- ============================================
-- TASK TYPES
-- ============================================

-- Задание 6
INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 6, t.id, s.id, 'fractions_common', 'Действия с обыкновенными дробями', 'Сложение, вычитание, умножение, деление обыкновенных и смешанных дробей'
FROM topics t, subtopics s WHERE t.code='algebra' AND s.code='fractions_common'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 6, t.id, s.id, 'fractions_decimal', 'Действия с десятичными дробями', 'Арифметические операции с десятичными дробями'
FROM topics t, subtopics s WHERE t.code='algebra' AND s.code='fractions_decimal'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 6, t.id, s.id, 'fractions_mixed', 'Действия с обыкновенными и десятичными дробями', 'Совместные вычисления, порядок действий'
FROM topics t, subtopics s WHERE t.code='algebra' AND s.code='fractions_mixed'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

-- Задание 7
INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 7, t.id, s.id, 'compare_numbers', 'Сравнение чисел', 'Какое число заключено между двумя числами'
FROM topics t, subtopics s WHERE t.code='algebra' AND s.code='compare_numbers'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 7, t.id, s.id, 'number_line', 'Числа на координатной прямой', 'Определение положения чисел на координатной прямой'
FROM topics t, subtopics s WHERE t.code='algebra' AND s.code='number_line'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 7, t.id, s.id, 'simple_inequalities', 'Простейшие неравенства', 'Дано множество чисел, найти принадлежащее ему'
FROM topics t, subtopics s WHERE t.code='algebra' AND s.code='simple_inequalities'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

-- Задание 8
INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 8, t.id, s.id, 'powers', 'Степени с целым показателем', 'Свойства степеней, действия со степенями'
FROM topics t, subtopics s WHERE t.code='algebra' AND s.code='powers'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 8, t.id, s.id, 'roots', 'Арифметические корни', 'Квадратные корни, вынесение/внесение множителя'
FROM topics t, subtopics s WHERE t.code='algebra' AND s.code='roots'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 8, t.id, s.id, 'formulas_short', 'Формулы сокращённого умножения', 'Квадрат суммы/разности, разность квадратов'
FROM topics t, subtopics s WHERE t.code='algebra' AND s.code='formulas_short'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

-- Задание 9
INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 9, t.id, s.id, 'eq_linear', 'Линейные уравнения', 'Уравнения первой степени'
FROM topics t, subtopics s WHERE t.code='algebra' AND s.code='eq_linear'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 9, t.id, s.id, 'eq_quadratic', 'Квадратные уравнения', 'Полные и неполные квадратные уравнения'
FROM topics t, subtopics s WHERE t.code='algebra' AND s.code='eq_quadratic'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 9, t.id, s.id, 'eq_rational', 'Рациональные уравнения', 'Уравнения с переменной в знаменателе'
FROM topics t, subtopics s WHERE t.code='algebra' AND s.code='eq_rational'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 9, t.id, s.id, 'eq_systems_linear', 'Системы линейных уравнений', 'Метод подстановки, метод сложения'
FROM topics t, subtopics s WHERE t.code='algebra' AND s.code='eq_systems_linear'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

-- Задание 10
INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 10, t.id, s.id, 'prob_classic', 'Классические вероятности', 'Формула P = m/n'
FROM topics t, subtopics s WHERE t.code='algebra' AND s.code='prob_classic'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 10, t.id, s.id, 'prob_theorems', 'Теоремы о вероятностях', 'Вероятность противоположного события'
FROM topics t, subtopics s WHERE t.code='algebra' AND s.code='prob_theorems'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

-- Задание 11
INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 11, t.id, s.id, 'graph_kb', 'Знаки коэффициентов k и b', 'Определение знаков по графику линейной функции'
FROM topics t, subtopics s WHERE t.code='algebra' AND s.code='graph_kb'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 11, t.id, s.id, 'func_graph_match', 'Соответствие функций и графиков', 'Установление соответствия между функцией и графиком'
FROM topics t, subtopics s WHERE t.code='algebra' AND s.code='func_graph_match'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 11, t.id, s.id, 'graph_formula_match', 'Соответствие графиков и формул', 'Определение формулы по графику'
FROM topics t, subtopics s WHERE t.code='algebra' AND s.code='graph_formula_match'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

-- Задание 12
INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 12, t.id, s.id, 'formula_calc', 'Вычисление по формуле', 'Подстановка значений в формулу'
FROM topics t, subtopics s WHERE t.code='algebra' AND s.code='formula_calc'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 12, t.id, s.id, 'linear_eq_practical', 'Линейные уравнения в практических задачах', 'Текстовые задачи на движение, работу, проценты'
FROM topics t, subtopics s WHERE t.code='algebra' AND s.code='linear_eq_practical'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

-- Задание 13
INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 13, t.id, s.id, 'ineq_linear', 'Линейные неравенства', 'Неравенства первой степени с выбором ответа'
FROM topics t, subtopics s WHERE t.code='algebra' AND s.code='ineq_linear'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 13, t.id, s.id, 'ineq_quadratic', 'Квадратные неравенства', 'Метод интервалов'
FROM topics t, subtopics s WHERE t.code='algebra' AND s.code='ineq_quadratic'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 13, t.id, s.id, 'ineq_system', 'Системы неравенств', 'Решение систем неравенств'
FROM topics t, subtopics s WHERE t.code='algebra' AND s.code='ineq_system'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

-- Задание 14
INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 14, t.id, s.id, 'prog_arith', 'Арифметическая прогрессия', 'Формула n-го члена, сумма первых n членов'
FROM topics t, subtopics s WHERE t.code='algebra' AND s.code='prog_arith'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 14, t.id, s.id, 'prog_geom', 'Геометрическая прогрессия', 'Формула n-го члена, знаменатель, сумма'
FROM topics t, subtopics s WHERE t.code='algebra' AND s.code='prog_geom'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

-- Задание 15
INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 15, t.id, s.id, 'tri_angles', 'Углы треугольника', 'Сумма углов, внешний угол'
FROM topics t, subtopics s WHERE t.code='geometry' AND s.code='tri_angles'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 15, t.id, s.id, 'tri_general', 'Треугольники общего вида', 'Соотношения сторон и углов'
FROM topics t, subtopics s WHERE t.code='geometry' AND s.code='tri_general'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 15, t.id, s.id, 'tri_isosceles', 'Равнобедренные треугольники', 'Свойства равнобедренного треугольника'
FROM topics t, subtopics s WHERE t.code='geometry' AND s.code='tri_isosceles'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 15, t.id, s.id, 'tri_right', 'Прямоугольный треугольник', 'Теорема Пифагора'
FROM topics t, subtopics s WHERE t.code='geometry' AND s.code='tri_right'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

-- Задание 16
INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 16, t.id, s.id, 'circle_angles', 'Центральные и вписанные углы', 'Градусные меры дуг'
FROM topics t, subtopics s WHERE t.code='geometry' AND s.code='circle_angles'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 16, t.id, s.id, 'circle_inscribed', 'Окружность вписанная в многоугольник', 'Радиус вписанной окружности'
FROM topics t, subtopics s WHERE t.code='geometry' AND s.code='circle_inscribed'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 16, t.id, s.id, 'circle_circumscribed', 'Окружность описанная вокруг многоугольника', 'Радиус описанной окружности'
FROM topics t, subtopics s WHERE t.code='geometry' AND s.code='circle_circumscribed'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 16, t.id, s.id, 'circle_tangent_chord', 'Касательная хорда секущая радиус', 'Свойства касательных'
FROM topics t, subtopics s WHERE t.code='geometry' AND s.code='circle_tangent_chord'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

-- Задание 17
INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 17, t.id, s.id, 'quad_trapezoid', 'Трапеция', 'Средняя линия, углы при основании, площадь'
FROM topics t, subtopics s WHERE t.code='geometry' AND s.code='quad_trapezoid'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 17, t.id, s.id, 'quad_parallelogram', 'Параллелограмм', 'Противоположные стороны и углы, диагонали'
FROM topics t, subtopics s WHERE t.code='geometry' AND s.code='quad_parallelogram'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 17, t.id, s.id, 'quad_rectangle', 'Прямоугольник', 'Свойства углов и диагоналей'
FROM topics t, subtopics s WHERE t.code='geometry' AND s.code='quad_rectangle'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 17, t.id, s.id, 'quad_rhombus', 'Ромб', 'Равенство сторон, перпендикулярность диагоналей'
FROM topics t, subtopics s WHERE t.code='geometry' AND s.code='quad_rhombus'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

-- Задание 18
INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 18, t.id, s.id, 'geo_distance', 'Расстояние между точками', 'Формула расстояния, координатный метод'
FROM topics t, subtopics s WHERE t.code='geometry' AND s.code='geo_distance'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 18, t.id, s.id, 'geo_midline', 'Средняя линия', 'Средняя линия треугольника и трапеции'
FROM topics t, subtopics s WHERE t.code='geometry' AND s.code='geo_midline'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 18, t.id, s.id, 'geo_pythagor', 'Теорема Пифагора на клетчатой бумаге', 'Длины наклонных отрезков'
FROM topics t, subtopics s WHERE t.code='geometry' AND s.code='geo_pythagor'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 18, t.id, s.id, 'geo_area', 'Площади на клетчатой бумаге', 'Метод Пика, разбиение на фигуры'
FROM topics t, subtopics s WHERE t.code='geometry' AND s.code='geo_area'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 18, t.id, s.id, 'geo_side_length', 'Длины сторон', 'Нахождение длин сторон по координатам'
FROM topics t, subtopics s WHERE t.code='geometry' AND s.code='geo_side_length'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

-- Задание 19
INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 19, t.id, s.id, 'logic_angles_lines', 'Утверждения об углах и прямых', 'Вертикальные, смежные, накрест лежащие углы'
FROM topics t, subtopics s WHERE t.code='geometry' AND s.code='logic_angles_lines'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 19, t.id, s.id, 'logic_triangles', 'Утверждения о треугольниках', 'Признаки равенства/подобия, теоремы'
FROM topics t, subtopics s WHERE t.code='geometry' AND s.code='logic_triangles'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 19, t.id, s.id, 'logic_quads', 'Утверждения о четырёхугольниках', 'Свойства параллелограмма, ромба, прямоугольника'
FROM topics t, subtopics s WHERE t.code='geometry' AND s.code='logic_quads'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 19, t.id, s.id, 'logic_circle', 'Утверждения об окружности', 'Вписанные/центральные углы, касательные'
FROM topics t, subtopics s WHERE t.code='geometry' AND s.code='logic_circle'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

COMMIT;