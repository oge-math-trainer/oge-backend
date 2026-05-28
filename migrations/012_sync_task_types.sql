BEGIN;

-- ============================================
-- Задание 6: переименовываем подтипы
-- numbers_fractions → fractions_common
-- numbers_decimal → fractions_decimal
-- numbers_integers → fractions_mixed
-- ============================================
UPDATE task_types SET subtype_code = 'fractions_common', title = 'Действия с обыкновенными дробями'
WHERE oge_number = 6 AND subtype_code = 'numbers_fractions';

UPDATE task_types SET subtype_code = 'fractions_decimal', title = 'Действия с десятичными дробями'
WHERE oge_number = 6 AND subtype_code = 'numbers_decimal';

UPDATE task_types SET subtype_code = 'fractions_mixed', title = 'Действия с обыкновенными и десятичными дробями'
WHERE oge_number = 6 AND subtype_code = 'numbers_integers';

-- ============================================
-- Задание 7: переименовываем подтипы
-- numberline_compare → compare_numbers
-- inequalities_simple → simple_inequalities
-- добавляем number_line
-- ============================================
UPDATE task_types SET subtype_code = 'compare_numbers', title = 'Сравнение чисел (без прямой)'
WHERE oge_number = 7 AND subtype_code = 'numberline_compare';

UPDATE task_types SET subtype_code = 'simple_inequalities', title = 'Простейшие неравенства'
WHERE oge_number = 7 AND subtype_code = 'inequalities_simple';

INSERT INTO subtopics (topic_id, code, title, description)
SELECT t.id, 'number_line', 'Числа на координатной прямой', 'Определение положения чисел на координатной прямой'
FROM topics t WHERE t.code = 'algebra'
ON CONFLICT (topic_id, code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 7, t.id, s.id, 'number_line', 'Числа на координатной прямой', 'Определение положения чисел на координатной прямой'
FROM topics t, subtopics s
WHERE t.code = 'algebra' AND s.code = 'number_line'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

-- ============================================
-- Задание 8: переименовываем подтипы
-- algebra_powers → powers
-- algebra_roots → roots
-- algebra_formulas → formulas_short
-- удаляем algebra_fractions (нет в новом JSON)
-- ============================================
UPDATE task_types SET subtype_code = 'powers', title = 'Степени с целым показателем'
WHERE oge_number = 8 AND subtype_code = 'algebra_powers';

UPDATE task_types SET subtype_code = 'roots', title = 'Арифметические корни'
WHERE oge_number = 8 AND subtype_code = 'algebra_roots';

UPDATE task_types SET subtype_code = 'formulas_short', title = 'Формулы сокращённого умножения'
WHERE oge_number = 8 AND subtype_code = 'algebra_formulas';

DELETE FROM task_types WHERE oge_number = 8 AND subtype_code = 'algebra_fractions';

-- ============================================
-- Задание 9: переименовываем подтипы
-- equations_linear → eq_linear
-- equations_quadratic → eq_quadratic
-- equations_rational → eq_rational
-- equations_system_linear → eq_systems_linear
-- удаляем equations_system_mixed (нет в новом JSON)
-- ============================================
UPDATE task_types SET subtype_code = 'eq_linear', title = 'Линейные уравнения'
WHERE oge_number = 9 AND subtype_code = 'equations_linear';

UPDATE task_types SET subtype_code = 'eq_quadratic', title = 'Квадратные уравнения'
WHERE oge_number = 9 AND subtype_code = 'equations_quadratic';

UPDATE task_types SET subtype_code = 'eq_rational', title = 'Рациональные уравнения'
WHERE oge_number = 9 AND subtype_code = 'equations_rational';

UPDATE task_types SET subtype_code = 'eq_systems_linear', title = 'Системы линейных уравнений'
WHERE oge_number = 9 AND subtype_code = 'equations_system_linear';

DELETE FROM task_types WHERE oge_number = 9 AND subtype_code = 'equations_system_mixed';

-- ============================================
-- Задание 10: переименовываем подтипы
-- probability_classic → prob_classic
-- probability_tree → prob_theorems
-- ============================================
UPDATE task_types SET subtype_code = 'prob_classic', title = 'Классические вероятности'
WHERE oge_number = 10 AND subtype_code = 'probability_classic';

UPDATE task_types SET subtype_code = 'prob_theorems', title = 'Теоремы о вероятностях событий'
WHERE oge_number = 10 AND subtype_code = 'probability_tree';

-- ============================================
-- Задание 11: переименовываем подтипы
-- graphs_linear → graph_kb
-- graphs_match → func_graph_match
-- удаляем graphs_quadratic, graphs_inverse, graphs_transform
-- добавляем graph_formula_match
-- ============================================
UPDATE task_types SET subtype_code = 'graph_kb', title = 'Соответствие графиков и знаков коэффициентов k и b'
WHERE oge_number = 11 AND subtype_code = 'graphs_linear';

UPDATE task_types SET subtype_code = 'func_graph_match', title = 'Соответствие функций и их графиков'
WHERE oge_number = 11 AND subtype_code = 'graphs_match';

DELETE FROM task_types WHERE oge_number = 11 AND subtype_code IN ('graphs_quadratic', 'graphs_inverse', 'graphs_transform');

INSERT INTO subtopics (topic_id, code, title, description)
SELECT t.id, 'graph_formula_match', 'Соответствие графиков и формул', 'Определение формулы по графику функции'
FROM topics t WHERE t.code = 'algebra'
ON CONFLICT (topic_id, code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 11, t.id, s.id, 'graph_formula_match', 'Соответствие графиков и формул', 'Определение аналитической формулы по заданному графику'
FROM topics t, subtopics s
WHERE t.code = 'algebra' AND s.code = 'graph_formula_match'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

-- ============================================
-- Задание 12: переименовываем и добавляем подтипы
-- formulas_subst → formula_calc
-- удаляем formulas_units, formulas_practical
-- добавляем linear_eq_practical
-- ============================================
UPDATE task_types SET subtype_code = 'formula_calc', title = 'Вычисление по формуле (практика)'
WHERE oge_number = 12 AND subtype_code = 'formulas_subst';

DELETE FROM task_types WHERE oge_number = 12 AND subtype_code IN ('formulas_units', 'formulas_practical');

INSERT INTO subtopics (topic_id, code, title, description)
SELECT t.id, 'linear_eq_practical', 'Линейные уравнения в практических задачах', 'Текстовые задачи на движение, работу, проценты'
FROM topics t WHERE t.code = 'algebra'
ON CONFLICT (topic_id, code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 12, t.id, s.id, 'linear_eq_practical', 'Линейные уравнения в практических задачах', 'Текстовые задачи на движение, работу, проценты'
FROM topics t, subtopics s
WHERE t.code = 'algebra' AND s.code = 'linear_eq_practical'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

-- ============================================
-- Задание 13: подтипы совпадают, оставляем как есть
-- ============================================

-- ============================================
-- Задание 14: переименовываем подтипы
-- progression_arithm → prog_arith
-- progression_geom → prog_geom
-- ============================================
UPDATE task_types SET subtype_code = 'prog_arith', title = 'Арифметическая прогрессия'
WHERE oge_number = 14 AND subtype_code = 'progression_arithm';

UPDATE task_types SET subtype_code = 'prog_geom', title = 'Геометрическая прогрессия'
WHERE oge_number = 14 AND subtype_code = 'progression_geom';

-- ============================================
-- Задание 15: переименовываем и добавляем подтипы
-- triangles_angles → tri_angles
-- удаляем triangles_pythagor, triangles_area, triangles_lines
-- добавляем tri_general, tri_isosceles, tri_right
-- ============================================
UPDATE task_types SET subtype_code = 'tri_angles', title = 'Углы треугольника'
WHERE oge_number = 15 AND subtype_code = 'triangles_angles';

DELETE FROM task_types WHERE oge_number = 15 AND subtype_code IN ('triangles_pythagor', 'triangles_area', 'triangles_lines');

INSERT INTO subtopics (topic_id, code, title, description)
SELECT t.id, 'tri_general', 'Треугольники общего вида', 'Соотношения сторон и углов, медианы, биссектрисы, высоты'
FROM topics t WHERE t.code = 'geometry'
ON CONFLICT (topic_id, code) DO NOTHING;

INSERT INTO subtopics (topic_id, code, title, description)
SELECT t.id, 'tri_isosceles', 'Равнобедренные треугольники', 'Свойства углов при основании, медиана/биссектриса/высота к основанию'
FROM topics t WHERE t.code = 'geometry'
ON CONFLICT (topic_id, code) DO NOTHING;

INSERT INTO subtopics (topic_id, code, title, description)
SELECT t.id, 'tri_right', 'Прямоугольный треугольник', 'Теорема Пифагора, тригонометрические отношения'
FROM topics t WHERE t.code = 'geometry'
ON CONFLICT (topic_id, code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 15, t.id, s.id, 'tri_general', 'Треугольники общего вида', 'Соотношения сторон и углов'
FROM topics t, subtopics s WHERE t.code = 'geometry' AND s.code = 'tri_general'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 15, t.id, s.id, 'tri_isosceles', 'Равнобедренные треугольники', 'Свойства равнобедренного треугольника'
FROM topics t, subtopics s WHERE t.code = 'geometry' AND s.code = 'tri_isosceles'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 15, t.id, s.id, 'tri_right', 'Прямоугольный треугольник', 'Теорема Пифагора и свойства прямоугольного треугольника'
FROM topics t, subtopics s WHERE t.code = 'geometry' AND s.code = 'tri_right'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

-- ============================================
-- Задание 16: переименовываем и добавляем подтипы
-- circle_angles → circle_angles (без изменений)
-- circle_inscribed → circle_inscribed (без изменений)
-- удаляем circle_elements, circle_tangents, circle_sector
-- добавляем circle_circumscribed, circle_tangent_chord
-- ============================================
DELETE FROM task_types WHERE oge_number = 16 AND subtype_code IN ('circle_elements', 'circle_tangents', 'circle_sector');

INSERT INTO subtopics (topic_id, code, title, description)
SELECT t.id, 'circle_circumscribed', 'Окружность описанная вокруг многоугольника', 'Радиус описанной окружности'
FROM topics t WHERE t.code = 'geometry'
ON CONFLICT (topic_id, code) DO NOTHING;

INSERT INTO subtopics (topic_id, code, title, description)
SELECT t.id, 'circle_tangent_chord', 'Касательная хорда секущая радиус', 'Свойства касательных, равенство отрезков'
FROM topics t WHERE t.code = 'geometry'
ON CONFLICT (topic_id, code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 16, t.id, s.id, 'circle_circumscribed', 'Окружность описанная вокруг многоугольника', 'Радиус описанной окружности'
FROM topics t, subtopics s WHERE t.code = 'geometry' AND s.code = 'circle_circumscribed'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 16, t.id, s.id, 'circle_tangent_chord', 'Касательная хорда секущая радиус', 'Свойства касательных'
FROM topics t, subtopics s WHERE t.code = 'geometry' AND s.code = 'circle_tangent_chord'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

-- ============================================
-- Задание 17: переименовываем и добавляем подтипы
-- quad_trapezoid → quad_trapezoid (без изменений)
-- удаляем quad_properties, quad_area, quad_diagonals
-- добавляем quad_parallelogram, quad_rectangle, quad_rhombus
-- ============================================
DELETE FROM task_types WHERE oge_number = 17 AND subtype_code IN ('quad_properties', 'quad_area', 'quad_diagonals');

INSERT INTO subtopics (topic_id, code, title, description)
SELECT t.id, 'quad_parallelogram', 'Параллелограмм', 'Противоположные стороны и углы, диагонали, площадь'
FROM topics t WHERE t.code = 'geometry'
ON CONFLICT (topic_id, code) DO NOTHING;

INSERT INTO subtopics (topic_id, code, title, description)
SELECT t.id, 'quad_rectangle', 'Прямоугольник', 'Свойства углов и диагоналей, площадь'
FROM topics t WHERE t.code = 'geometry'
ON CONFLICT (topic_id, code) DO NOTHING;

INSERT INTO subtopics (topic_id, code, title, description)
SELECT t.id, 'quad_rhombus', 'Ромб', 'Равенство сторон, перпендикулярность диагоналей, площадь'
FROM topics t WHERE t.code = 'geometry'
ON CONFLICT (topic_id, code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 17, t.id, s.id, 'quad_parallelogram', 'Параллелограмм', 'Свойства параллелограмма'
FROM topics t, subtopics s WHERE t.code = 'geometry' AND s.code = 'quad_parallelogram'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 17, t.id, s.id, 'quad_rectangle', 'Прямоугольник', 'Свойства прямоугольника'
FROM topics t, subtopics s WHERE t.code = 'geometry' AND s.code = 'quad_rectangle'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 17, t.id, s.id, 'quad_rhombus', 'Ромб', 'Свойства ромба'
FROM topics t, subtopics s WHERE t.code = 'geometry' AND s.code = 'quad_rhombus'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

-- ============================================
-- Задание 18: переименовываем и добавляем подтипы
-- grid_distance → geo_distance
-- grid_midline → geo_midline
-- grid_pythagor → geo_pythagor
-- grid_area → geo_area
-- добавляем geo_side_length
-- ============================================
UPDATE task_types SET subtype_code = 'geo_distance', title = 'Расстояние между точками и от точки до прямой'
WHERE oge_number = 18 AND subtype_code = 'grid_distance';

UPDATE task_types SET subtype_code = 'geo_midline', title = 'Средняя линия'
WHERE oge_number = 18 AND subtype_code = 'grid_midline';

UPDATE task_types SET subtype_code = 'geo_pythagor', title = 'Теорема Пифагора'
WHERE oge_number = 18 AND subtype_code = 'grid_pythagor';

UPDATE task_types SET subtype_code = 'geo_area', title = 'Площади'
WHERE oge_number = 18 AND subtype_code = 'grid_area';

INSERT INTO subtopics (topic_id, code, title, description)
SELECT t.id, 'geo_side_length', 'Длины сторон', 'Нахождение длин сторон многоугольников по координатам'
FROM topics t WHERE t.code = 'geometry'
ON CONFLICT (topic_id, code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 18, t.id, s.id, 'geo_side_length', 'Длины сторон', 'Нахождение длин сторон по координатам вершин'
FROM topics t, subtopics s WHERE t.code = 'geometry' AND s.code = 'geo_side_length'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

-- ============================================
-- Задание 19: переименовываем подтипы
-- logic_angles → logic_angles_lines
-- logic_triangles → logic_triangles (без изменений)
-- logic_quad → logic_quads
-- logic_circle → logic_circle (без изменений)
-- ============================================
UPDATE task_types SET subtype_code = 'logic_angles_lines', title = 'Утверждения об углах и прямых'
WHERE oge_number = 19 AND subtype_code = 'logic_angles';

UPDATE task_types SET subtype_code = 'logic_quads', title = 'Утверждения о четырёхугольниках'
WHERE oge_number = 19 AND subtype_code = 'logic_quad';

COMMIT;