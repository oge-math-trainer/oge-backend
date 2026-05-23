BEGIN;

-- ============================================
-- ОБНОВЛЕНИЕ ПОДТИПОВ ПО СПИСКУ "подтипы (2).txt"
-- ============================================

-- 6. Числа и вычисления
INSERT INTO subtopics (topic_id, code, title, description) VALUES
((SELECT id FROM topics WHERE code='algebra'), 'fractions_common', 'Действия с обыкновенными дробями', 'Сложение, вычитание, умножение, деление обыкновенных дробей'),
((SELECT id FROM topics WHERE code='algebra'), 'fractions_decimal', 'Действия с десятичными дробями', 'Арифметические операции с десятичными дробями'),
((SELECT id FROM topics WHERE code='algebra'), 'fractions_mixed', 'Действия с обыкновенными и десятичными дробями', 'Смешанные операции с дробями разных видов')
ON CONFLICT (topic_id, code) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 6, t.id, s.id, 'fractions_common', 'Обыкновенные дроби', 'Действия с обыкновенными дробями' FROM topics t, subtopics s WHERE t.code='algebra' AND s.code='fractions_common'
ON CONFLICT (oge_number, subtype_code) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 6, t.id, s.id, 'fractions_decimal', 'Десятичные дроби', 'Действия с десятичными дробями' FROM topics t, subtopics s WHERE t.code='algebra' AND s.code='fractions_decimal'
ON CONFLICT (oge_number, subtype_code) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 6, t.id, s.id, 'fractions_mixed', 'Смешанные дроби', 'Действия с обыкновенными и десятичными дробями' FROM topics t, subtopics s WHERE t.code='algebra' AND s.code='fractions_mixed'
ON CONFLICT (oge_number, subtype_code) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description;

-- 7. Числовая прямая и неравенства
INSERT INTO subtopics (topic_id, code, title, description) VALUES
((SELECT id FROM topics WHERE code='algebra'), 'inequalities_simple', 'Неравенства', 'Сравнение выражений, оценка значений'),
((SELECT id FROM topics WHERE code='algebra'), 'numbers_compare', 'Сравнение чисел', 'Сравнение целых, дробей, корней, степеней'),
((SELECT id FROM topics WHERE code='algebra'), 'numberline_plot', 'Числа на прямой', 'Расположение точек на числовой прямой')
ON CONFLICT (topic_id, code) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 7, t.id, s.id, 'inequalities_simple', 'Неравенства', 'Простейшие неравенства' FROM topics t, subtopics s WHERE t.code='algebra' AND s.code='inequalities_simple'
ON CONFLICT (oge_number, subtype_code) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 7, t.id, s.id, 'numbers_compare', 'Сравнение чисел', 'Сравнение чисел и выражений' FROM topics t, subtopics s WHERE t.code='algebra' AND s.code='numbers_compare'
ON CONFLICT (oge_number, subtype_code) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 7, t.id, s.id, 'numberline_plot', 'Числа на прямой', 'Координатная прямая' FROM topics t, subtopics s WHERE t.code='algebra' AND s.code='numberline_plot'
ON CONFLICT (oge_number, subtype_code) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description;

-- 8. Алгебраические выражения
INSERT INTO subtopics (topic_id, code, title, description) VALUES
((SELECT id FROM topics WHERE code='algebra'), 'powers', 'Степени', 'Свойства степеней с целым показателем'),
((SELECT id FROM topics WHERE code='algebra'), 'roots', 'Корни', 'Арифметические квадратные корни'),
((SELECT id FROM topics WHERE code='algebra'), 'fsu', 'Формулы сокращённого умножения', 'Формулы (a±b)², a²-b², разложение на множители')
ON CONFLICT (topic_id, code) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 8, t.id, s.id, 'powers', 'Степени', 'Степени с целым показателем' FROM topics t, subtopics s WHERE t.code='algebra' AND s.code='powers'
ON CONFLICT (oge_number, subtype_code) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 8, t.id, s.id, 'roots', 'Корни', 'Квадратные корни' FROM topics t, subtopics s WHERE t.code='algebra' AND s.code='roots'
ON CONFLICT (oge_number, subtype_code) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 8, t.id, s.id, 'fsu', 'ФСУ', 'Формулы сокращённого умножения' FROM topics t, subtopics s WHERE t.code='algebra' AND s.code='fsu'
ON CONFLICT (oge_number, subtype_code) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description;

-- 9. Уравнения и системы
INSERT INTO subtopics (topic_id, code, title, description) VALUES
((SELECT id FROM topics WHERE code='algebra'), 'eq_linear', 'Линейные уравнения', 'Уравнения вида ax + b = 0'),
((SELECT id FROM topics WHERE code='algebra'), 'eq_quadratic', 'Квадратные уравнения', 'Дискриминант, теорема Виета'),
((SELECT id FROM topics WHERE code='algebra'), 'eq_rational', 'Рациональные уравнения', 'Дробно-рациональные уравнения'),
((SELECT id FROM topics WHERE code='algebra'), 'eq_system_linear', 'Системы линейных уравнений', 'Метод подстановки или сложения')
ON CONFLICT (topic_id, code) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 9, t.id, s.id, 'eq_linear', 'Линейные уравнения', 'Решение линейных уравнений' FROM topics t, subtopics s WHERE t.code='algebra' AND s.code='eq_linear'
ON CONFLICT (oge_number, subtype_code) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 9, t.id, s.id, 'eq_quadratic', 'Квадратные уравнения', 'Решение квадратных уравнений' FROM topics t, subtopics s WHERE t.code='algebra' AND s.code='eq_quadratic'
ON CONFLICT (oge_number, subtype_code) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 9, t.id, s.id, 'eq_rational', 'Рациональные уравнения', 'Дробно-рациональные уравнения' FROM topics t, subtopics s WHERE t.code='algebra' AND s.code='eq_rational'
ON CONFLICT (oge_number, subtype_code) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 9, t.id, s.id, 'eq_system_linear', 'Системы линейных уравнений', 'Системы двух линейных уравнений' FROM topics t, subtopics s WHERE t.code='algebra' AND s.code='eq_system_linear'
ON CONFLICT (oge_number, subtype_code) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description;

-- 10. Вероятность
INSERT INTO subtopics (topic_id, code, title, description) VALUES
((SELECT id FROM topics WHERE code='algebra'), 'prob_classic', 'Классические вероятности', 'P = m/n, монеты, кубики, шары'),
((SELECT id FROM topics WHERE code='algebra'), 'prob_theorems', 'Теоремы о вероятностях', 'Сложение и умножение вероятностей, условия')
ON CONFLICT (topic_id, code) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 10, t.id, s.id, 'prob_classic', 'Классическая вероятность', 'Вычисление вероятности события' FROM topics t, subtopics s WHERE t.code='algebra' AND s.code='prob_classic'
ON CONFLICT (oge_number, subtype_code) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 10, t.id, s.id, 'prob_theorems', 'Теоремы о вероятностях', 'Теоремы сложения и умножения' FROM topics t, subtopics s WHERE t.code='algebra' AND s.code='prob_theorems'
ON CONFLICT (oge_number, subtype_code) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description;

-- 11. Графики функций (соответствия)
INSERT INTO subtopics (topic_id, code, title, description) VALUES
((SELECT id FROM topics WHERE code='algebra'), 'graphs_linear_signs', 'Графики и знаки k, b', 'Соответствие между графиками y=kx+b и знаками коэффициентов'),
((SELECT id FROM topics WHERE code='algebra'), 'graphs_match_func', 'Функции и графики', 'Соответствие между аналитическим видом и графиком'),
((SELECT id FROM topics WHERE code='algebra'), 'graphs_match_formula', 'Графики и формулы', 'Выбор формулы по графику и наоборот')
ON CONFLICT (topic_id, code) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 11, t.id, s.id, 'graphs_linear_signs', 'Знаки коэффициентов', 'Соответствие графиков и знаков k, b' FROM topics t, subtopics s WHERE t.code='algebra' AND s.code='graphs_linear_signs'
ON CONFLICT (oge_number, subtype_code) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 11, t.id, s.id, 'graphs_match_func', 'Функции и графики', 'Соответствие функций и графиков' FROM topics t, subtopics s WHERE t.code='algebra' AND s.code='graphs_match_func'
ON CONFLICT (oge_number, subtype_code) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 11, t.id, s.id, 'graphs_match_formula', 'Графики и формулы', 'Соответствие графиков и формул' FROM topics t, subtopics s WHERE t.code='algebra' AND s.code='graphs_match_formula'
ON CONFLICT (oge_number, subtype_code) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description;

-- 12. Расчёты по формулам
INSERT INTO subtopics (topic_id, code, title, description) VALUES
((SELECT id FROM topics WHERE code='algebra'), 'formula_compute', 'Вычисление по формуле', 'Подстановка значений в готовые формулы'),
((SELECT id FROM topics WHERE code='algebra'), 'formula_linear', 'Линейные уравнения', 'Текстовые задачи на составление линейного уравнения')
ON CONFLICT (topic_id, code) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 12, t.id, s.id, 'formula_compute', 'Вычисление по формуле', 'Подстановка в формулу' FROM topics t, subtopics s WHERE t.code='algebra' AND s.code='formula_compute'
ON CONFLICT (oge_number, subtype_code) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 12, t.id, s.id, 'formula_linear', 'Линейные уравнения', 'Практические задачи на линейные уравнения' FROM topics t, subtopics s WHERE t.code='algebra' AND s.code='formula_linear'
ON CONFLICT (oge_number, subtype_code) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description;

-- 13. Неравенства
INSERT INTO subtopics (topic_id, code, title, description) VALUES
((SELECT id FROM topics WHERE code='algebra'), 'ineq_linear', 'Линейные неравенства', 'Решение неравенств вида ax + b > 0'),
((SELECT id FROM topics WHERE code='algebra'), 'ineq_quadratic', 'Квадратные неравенства', 'Метод интервалов, ветви параболы'),
((SELECT id FROM topics WHERE code='algebra'), 'ineq_system', 'Системы неравенств', 'Пересечение числовых промежутков')
ON CONFLICT (topic_id, code) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 13, t.id, s.id, 'ineq_linear', 'Линейные неравенства', 'Решение линейных неравенств' FROM topics t, subtopics s WHERE t.code='algebra' AND s.code='ineq_linear'
ON CONFLICT (oge_number, subtype_code) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 13, t.id, s.id, 'ineq_quadratic', 'Квадратные неравенства', 'Решение квадратных неравенств' FROM topics t, subtopics s WHERE t.code='algebra' AND s.code='ineq_quadratic'
ON CONFLICT (oge_number, subtype_code) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 13, t.id, s.id, 'ineq_system', 'Системы неравенств', 'Решение систем неравенств' FROM topics t, subtopics s WHERE t.code='algebra' AND s.code='ineq_system'
ON CONFLICT (oge_number, subtype_code) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description;

-- 14. Прогрессии
INSERT INTO subtopics (topic_id, code, title, description) VALUES
((SELECT id FROM topics WHERE code='algebra'), 'progression_arithm', 'Арифметическая прогрессия', 'n-й член, разность, сумма'),
((SELECT id FROM topics WHERE code='algebra'), 'progression_geom', 'Геометрическая прогрессия', 'n-й член, знаменатель, сумма')
ON CONFLICT (topic_id, code) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 14, t.id, s.id, 'progression_arithm', 'Арифметическая прогрессия', 'Задачи на арифметическую прогрессию' FROM topics t, subtopics s WHERE t.code='algebra' AND s.code='progression_arithm'
ON CONFLICT (oge_number, subtype_code) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 14, t.id, s.id, 'progression_geom', 'Геометрическая прогрессия', 'Задачи на геометрическую прогрессию' FROM topics t, subtopics s WHERE t.code='algebra' AND s.code='progression_geom'
ON CONFLICT (oge_number, subtype_code) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description;

-- 15. Треугольники
INSERT INTO subtopics (topic_id, code, title, description) VALUES
((SELECT id FROM topics WHERE code='geometry'), 'tri_angles', 'Углы треугольника', 'Сумма углов, внешний угол'),
((SELECT id FROM topics WHERE code='geometry'), 'tri_general', 'Треугольники общего вида', 'Неравенство треугольника, произвольные треугольники'),
((SELECT id FROM topics WHERE code='geometry'), 'tri_isosceles', 'Равнобедренные треугольники', 'Свойства равнобедренного и равностороннего треугольников'),
((SELECT id FROM topics WHERE code='geometry'), 'tri_right', 'Прямоугольный треугольник', 'Теорема Пифагора, катеты, гипотенуза, тригонометрия')
ON CONFLICT (topic_id, code) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 15, t.id, s.id, 'tri_angles', 'Углы треугольника', 'Задачи на углы' FROM topics t, subtopics s WHERE t.code='geometry' AND s.code='tri_angles'
ON CONFLICT (oge_number, subtype_code) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 15, t.id, s.id, 'tri_general', 'Треугольники общего вида', 'Свойства произвольных треугольников' FROM topics t, subtopics s WHERE t.code='geometry' AND s.code='tri_general'
ON CONFLICT (oge_number, subtype_code) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 15, t.id, s.id, 'tri_isosceles', 'Равнобедренные треугольники', 'Свойства равнобедренных треугольников' FROM topics t, subtopics s WHERE t.code='geometry' AND s.code='tri_isosceles'
ON CONFLICT (oge_number, subtype_code) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 15, t.id, s.id, 'tri_right', 'Прямоугольный треугольник', 'Теорема Пифагора и прямоугольные треугольники' FROM topics t, subtopics s WHERE t.code='geometry' AND s.code='tri_right'
ON CONFLICT (oge_number, subtype_code) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description;

-- 16. Окружность
INSERT INTO subtopics (topic_id, code, title, description) VALUES
((SELECT id FROM topics WHERE code='geometry'), 'circle_angles', 'Центральные и вписанные углы', 'Градусные меры углов, опирающиеся на дугу'),
((SELECT id FROM topics WHERE code='geometry'), 'circle_inscribed', 'Окружность вписанная', 'Окружность, вписанная в многоугольник'),
((SELECT id FROM topics WHERE code='geometry'), 'circle_circumscribed', 'Окружность описанная', 'Окружность, описанная вокруг многоугольника'),
((SELECT id FROM topics WHERE code='geometry'), 'circle_elements', 'Касательная, хорда, секущая, радиус', 'Элементы окружности и их свойства')
ON CONFLICT (topic_id, code) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 16, t.id, s.id, 'circle_angles', 'Центральные и вписанные углы', 'Углы в окружности' FROM topics t, subtopics s WHERE t.code='geometry' AND s.code='circle_angles'
ON CONFLICT (oge_number, subtype_code) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 16, t.id, s.id, 'circle_inscribed', 'Вписанная окружность', 'Окружность вписанная в многоугольник' FROM topics t, subtopics s WHERE t.code='geometry' AND s.code='circle_inscribed'
ON CONFLICT (oge_number, subtype_code) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 16, t.id, s.id, 'circle_circumscribed', 'Описанная окружность', 'Окружность описанная вокруг многоугольника' FROM topics t, subtopics s WHERE t.code='geometry' AND s.code='circle_circumscribed'
ON CONFLICT (oge_number, subtype_code) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 16, t.id, s.id, 'circle_elements', 'Элементы окружности', 'Касательная, хорда, секущая, радиус' FROM topics t, subtopics s WHERE t.code='geometry' AND s.code='circle_elements'
ON CONFLICT (oge_number, subtype_code) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description;

-- 17. Четырёхугольники
INSERT INTO subtopics (topic_id, code, title, description) VALUES
((SELECT id FROM topics WHERE code='geometry'), 'quad_trapezoid', 'Трапеция', 'Свойства трапеции, средняя линия, диагонали'),
((SELECT id FROM topics WHERE code='geometry'), 'quad_parallelogram', 'Параллелограмм', 'Свойства параллелограмма'),
((SELECT id FROM topics WHERE code='geometry'), 'quad_rectangle', 'Прямоугольник', 'Свойства прямоугольника'),
((SELECT id FROM topics WHERE code='geometry'), 'quad_rhombus', 'Ромб', 'Свойства ромба')
ON CONFLICT (topic_id, code) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 17, t.id, s.id, 'quad_trapezoid', 'Трапеция', 'Задачи на трапецию' FROM topics t, subtopics s WHERE t.code='geometry' AND s.code='quad_trapezoid'
ON CONFLICT (oge_number, subtype_code) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 17, t.id, s.id, 'quad_parallelogram', 'Параллелограмм', 'Задачи на параллелограмм' FROM topics t, subtopics s WHERE t.code='geometry' AND s.code='quad_parallelogram'
ON CONFLICT (oge_number, subtype_code) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 17, t.id, s.id, 'quad_rectangle', 'Прямоугольник', 'Задачи на прямоугольник' FROM topics t, subtopics s WHERE t.code='geometry' AND s.code='quad_rectangle'
ON CONFLICT (oge_number, subtype_code) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 17, t.id, s.id, 'quad_rhombus', 'Ромб', 'Задачи на ромб' FROM topics t, subtopics s WHERE t.code='geometry' AND s.code='quad_rhombus'
ON CONFLICT (oge_number, subtype_code) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description;

-- 18. Фигуры на решётке
INSERT INTO subtopics (topic_id, code, title, description) VALUES
((SELECT id FROM topics WHERE code='geometry'), 'grid_distance', 'Расстояние между точками', 'Координатная плоскость, расстояние'),
((SELECT id FROM topics WHERE code='geometry'), 'grid_midline', 'Средняя линия', 'Средняя линия треугольника и трапеции'),
((SELECT id FROM topics WHERE code='geometry'), 'grid_pythagor', 'Теорема Пифагора', 'Применение теоремы Пифагора на клетках'),
((SELECT id FROM topics WHERE code='geometry'), 'grid_area', 'Площади', 'Метод Пика, подсчёт клеток'),
((SELECT id FROM topics WHERE code='geometry'), 'grid_length', 'Длины сторон', 'Вычисление длин отрезков на решётке')
ON CONFLICT (topic_id, code) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 18, t.id, s.id, 'grid_distance', 'Расстояние между точками', 'Расстояние на координатной плоскости' FROM topics t, subtopics s WHERE t.code='geometry' AND s.code='grid_distance'
ON CONFLICT (oge_number, subtype_code) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 18, t.id, s.id, 'grid_midline', 'Средняя линия', 'Свойства средней линии' FROM topics t, subtopics s WHERE t.code='geometry' AND s.code='grid_midline'
ON CONFLICT (oge_number, subtype_code) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 18, t.id, s.id, 'grid_pythagor', 'Теорема Пифагора', 'Пифагор на клетчатой бумаге' FROM topics t, subtopics s WHERE t.code='geometry' AND s.code='grid_pythagor'
ON CONFLICT (oge_number, subtype_code) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 18, t.id, s.id, 'grid_area', 'Площади', 'Площади фигур на решётке' FROM topics t, subtopics s WHERE t.code='geometry' AND s.code='grid_area'
ON CONFLICT (oge_number, subtype_code) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 18, t.id, s.id, 'grid_length', 'Длины сторон', 'Длины отрезков на решётке' FROM topics t, subtopics s WHERE t.code='geometry' AND s.code='grid_length'
ON CONFLICT (oge_number, subtype_code) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description;

-- 19. Утверждения
INSERT INTO subtopics (topic_id, code, title, description) VALUES
((SELECT id FROM topics WHERE code='geometry'), 'logic_angles_lines', 'Утверждения об углах и прямых', 'Вертикальные, смежные, параллельные прямые'),
((SELECT id FROM topics WHERE code='geometry'), 'logic_triangles', 'Утверждения о треугольниках', 'Равенство, подобие, площади, теоремы'),
((SELECT id FROM topics WHERE code='geometry'), 'logic_quad', 'Утверждения о четырёхугольниках', 'Свойства параллелограмма, трапеции, ромба, квадрата'),
((SELECT id FROM topics WHERE code='geometry'), 'logic_circle', 'Утверждения об окружности', 'Касательные, хорды, вписанные и центральные углы')
ON CONFLICT (topic_id, code) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 19, t.id, s.id, 'logic_angles_lines', 'Утверждения об углах и прямых', 'Анализ утверждений' FROM topics t, subtopics s WHERE t.code='geometry' AND s.code='logic_angles_lines'
ON CONFLICT (oge_number, subtype_code) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 19, t.id, s.id, 'logic_triangles', 'Утверждения о треугольниках', 'Анализ утверждений' FROM topics t, subtopics s WHERE t.code='geometry' AND s.code='logic_triangles'
ON CONFLICT (oge_number, subtype_code) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 19, t.id, s.id, 'logic_quad', 'Утверждения о четырёхугольниках', 'Анализ утверждений' FROM topics t, subtopics s WHERE t.code='geometry' AND s.code='logic_quad'
ON CONFLICT (oge_number, subtype_code) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 19, t.id, s.id, 'logic_circle', 'Утверждения об окружности', 'Анализ утверждений' FROM topics t, subtopics s WHERE t.code='geometry' AND s.code='logic_circle'
ON CONFLICT (oge_number, subtype_code) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description;

COMMIT;
