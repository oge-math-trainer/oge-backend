BEGIN;

-- ============================================
-- ТЕМА 1: АЛГЕБРА (Задания 6-14)
-- ============================================

INSERT INTO topics (code, title, description) VALUES
('algebra', 'Алгебра', 'Числа, вычисления, алгебраические выражения, уравнения, неравенства, функции'),
('geometry', 'Геометрия', 'Планиметрия: треугольники, окружности, четырёхугольники, многоугольники')
ON CONFLICT (code) DO NOTHING;

-- ============================================
-- ЗАДАНИЕ 6: Числа и вычисления
-- ============================================

INSERT INTO subtopics (topic_id, code, title, description) VALUES
((SELECT id FROM topics WHERE code='algebra'), 'numbers_fractions', 'Действия с обыкновенными дробями', 'Сложение, вычитание, умножение, деление обыкновенных дробей'),
((SELECT id FROM topics WHERE code='algebra'), 'numbers_decimal', 'Действия с десятичными дробями', 'Арифметические операции с десятичными дробями'),
((SELECT id FROM topics WHERE code='algebra'), 'numbers_integers', 'Целые и рациональные числа', 'Базовые арифметические операции с целыми и рациональными числами')
ON CONFLICT (topic_id, code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 6, t.id, s.id, 'numbers_fractions', 'Обыкновенные дроби', 'Действия с обыкновенными дробями'
FROM topics t, subtopics s 
WHERE t.code='algebra' AND s.code='numbers_fractions'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 6, t.id, s.id, 'numbers_decimal', 'Десятичные дроби', 'Действия с десятичными дробями'
FROM topics t, subtopics s 
WHERE t.code='algebra' AND s.code='numbers_decimal'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 6, t.id, s.id, 'numbers_integers', 'Целые числа', 'Действия с целыми и рациональными числами'
FROM topics t, subtopics s 
WHERE t.code='algebra' AND s.code='numbers_integers'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

-- ============================================
-- ЗАДАНИЕ 7: Числовая прямая и неравенства
-- ============================================

INSERT INTO subtopics (topic_id, code, title, description) VALUES
((SELECT id FROM topics WHERE code='algebra'), 'numberline_compare', 'Сравнение чисел на прямой', 'Расположение точек на числовой прямой, порядок чисел'),
((SELECT id FROM topics WHERE code='algebra'), 'inequalities_simple', 'Простейшие неравенства', 'Сравнение выражений, оценка значений')
ON CONFLICT (topic_id, code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 7, t.id, s.id, 'numberline_compare', 'Числовая прямая', 'Сравнение чисел на числовой прямой'
FROM topics t, subtopics s 
WHERE t.code='algebra' AND s.code='numberline_compare'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 7, t.id, s.id, 'inequalities_simple', 'Простейшие неравенства', 'Сравнение выражений и оценка значений'
FROM topics t, subtopics s 
WHERE t.code='algebra' AND s.code='inequalities_simple'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

-- ============================================
-- ЗАДАНИЕ 8: Алгебраические выражения
-- ============================================

INSERT INTO subtopics (topic_id, code, title, description) VALUES
((SELECT id FROM topics WHERE code='algebra'), 'algebra_powers', 'Степени с целым показателем', 'Свойства степеней, отрицательные показатели'),
((SELECT id FROM topics WHERE code='algebra'), 'algebra_roots', 'Арифметические корни', 'Квадратные корни, вынесение и внесение множителя'),
((SELECT id FROM topics WHERE code='algebra'), 'algebra_formulas', 'Формулы сокращённого умножения', 'Формулы (a±b)², a²-b², разложение на множители'),
((SELECT id FROM topics WHERE code='algebra'), 'algebra_fractions', 'Алгебраические дроби', 'Сокращение, сложение, умножение алгебраических дробей')
ON CONFLICT (topic_id, code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 8, t.id, s.id, 'algebra_powers', 'Степени', 'Степени с целым показателем'
FROM topics t, subtopics s 
WHERE t.code='algebra' AND s.code='algebra_powers'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 8, t.id, s.id, 'algebra_roots', 'Корни', 'Арифметические квадратные корни'
FROM topics t, subtopics s 
WHERE t.code='algebra' AND s.code='algebra_roots'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 8, t.id, s.id, 'algebra_formulas', 'Формулы сокращённого умножения', 'Применение формул сокращённого умножения'
FROM topics t, subtopics s 
WHERE t.code='algebra' AND s.code='algebra_formulas'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 8, t.id, s.id, 'algebra_fractions', 'Алгебраические дроби', 'Преобразование алгебраических дробей'
FROM topics t, subtopics s 
WHERE t.code='algebra' AND s.code='algebra_fractions'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

-- ============================================
-- ЗАДАНИЕ 9: Уравнения и системы
-- ============================================

INSERT INTO subtopics (topic_id, code, title, description) VALUES
((SELECT id FROM topics WHERE code='algebra'), 'equations_linear', 'Линейное уравнение', 'Уравнения вида ax + b = 0'),
((SELECT id FROM topics WHERE code='algebra'), 'equations_quadratic', 'Квадратное уравнение', 'Дискриминант, теорема Виета, неполные квадратные уравнения'),
((SELECT id FROM topics WHERE code='algebra'), 'equations_rational', 'Дробно-рациональное уравнение', 'Уравнения, сводящиеся к линейным или квадратным'),
((SELECT id FROM topics WHERE code='algebra'), 'equations_system_linear', 'Система линейных уравнений', 'Метод подстановки или сложения'),
((SELECT id FROM topics WHERE code='algebra'), 'equations_system_mixed', 'Смешанная система', 'Система с одним линейным и одним квадратным уравнением')
ON CONFLICT (topic_id, code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 9, t.id, s.id, 'equations_linear', 'Линейные уравнения', 'Решение линейных уравнений'
FROM topics t, subtopics s 
WHERE t.code='algebra' AND s.code='equations_linear'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 9, t.id, s.id, 'equations_quadratic', 'Квадратные уравнения', 'Решение квадратных уравнений'
FROM topics t, subtopics s 
WHERE t.code='algebra' AND s.code='equations_quadratic'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 9, t.id, s.id, 'equations_rational', 'Дробно-рациональные уравнения', 'Решение дробно-рациональных уравнений'
FROM topics t, subtopics s 
WHERE t.code='algebra' AND s.code='equations_rational'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 9, t.id, s.id, 'equations_system_linear', 'Системы линейных уравнений', 'Решение систем линейных уравнений'
FROM topics t, subtopics s 
WHERE t.code='algebra' AND s.code='equations_system_linear'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 9, t.id, s.id, 'equations_system_mixed', 'Смешанные системы', 'Решение смешанных систем уравнений'
FROM topics t, subtopics s 
WHERE t.code='algebra' AND s.code='equations_system_mixed'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

-- ============================================
-- ЗАДАНИЕ 10: Вероятность и статистика
-- ============================================

INSERT INTO subtopics (topic_id, code, title, description) VALUES
((SELECT id FROM topics WHERE code='algebra'), 'probability_classic', 'Классическая вероятность', 'Вероятность событий: монеты, кубики, шары, билеты'),
((SELECT id FROM topics WHERE code='algebra'), 'probability_tree', 'Дерево вероятностей', 'Многошаговые события, дерево вероятностей')
ON CONFLICT (topic_id, code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 10, t.id, s.id, 'probability_classic', 'Классическая вероятность', 'Вычисление классической вероятности'
FROM topics t, subtopics s 
WHERE t.code='algebra' AND s.code='probability_classic'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 10, t.id, s.id, 'probability_tree', 'Дерево вероятностей', 'Решение задач с использованием дерева вероятностей'
FROM topics t, subtopics s 
WHERE t.code='algebra' AND s.code='probability_tree'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

-- ============================================
-- ЗАДАНИЕ 11: Графики функций
-- ============================================

INSERT INTO subtopics (topic_id, code, title, description) VALUES
((SELECT id FROM topics WHERE code='algebra'), 'graphs_linear', 'Линейная функция', 'Функция y = kx + b, угловой коэффициент, параллельность'),
((SELECT id FROM topics WHERE code='algebra'), 'graphs_quadratic', 'Квадратичная функция', 'Парабола, вершина, сдвиги, направление ветвей'),
((SELECT id FROM topics WHERE code='algebra'), 'graphs_inverse', 'Обратная пропорциональность', 'Функция y = k/x, асимптоты, расположение в четвертях'),
((SELECT id FROM topics WHERE code='algebra'), 'graphs_transform', 'Сдвиги и растяжения', 'Преобразования графиков: параметры k, b, a'),
((SELECT id FROM topics WHERE code='algebra'), 'graphs_match', 'Соответствие графика и формулы', 'Выбор формулы по графику и наоборот')
ON CONFLICT (topic_id, code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 11, t.id, s.id, 'graphs_linear', 'Линейная функция', 'График линейной функции y = kx + b'
FROM topics t, subtopics s 
WHERE t.code='algebra' AND s.code='graphs_linear'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 11, t.id, s.id, 'graphs_quadratic', 'Квадратичная функция', 'График квадратичной функции (парабола)'
FROM topics t, subtopics s 
WHERE t.code='algebra' AND s.code='graphs_quadratic'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 11, t.id, s.id, 'graphs_inverse', 'Обратная пропорциональность', 'График функции y = k/x'
FROM topics t, subtopics s 
WHERE t.code='algebra' AND s.code='graphs_inverse'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 11, t.id, s.id, 'graphs_transform', 'Преобразования графиков', 'Сдвиги и растяжения графиков функций'
FROM topics t, subtopics s 
WHERE t.code='algebra' AND s.code='graphs_transform'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 11, t.id, s.id, 'graphs_match', 'Соответствие графика и формулы', 'Установление соответствия между графиком и формулой'
FROM topics t, subtopics s 
WHERE t.code='algebra' AND s.code='graphs_match'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

-- ============================================
-- ЗАДАНИЕ 12: Расчёты по формулам
-- ============================================

INSERT INTO subtopics (topic_id, code, title, description) VALUES
((SELECT id FROM topics WHERE code='algebra'), 'formulas_subst', 'Подстановка в формулу', 'Вычисление значения по готовой формуле'),
((SELECT id FROM topics WHERE code='algebra'), 'formulas_units', 'Перевод единиц измерения', 'Перевод между метрами, километрами, часами, минутами, процентами'),
((SELECT id FROM topics WHERE code='algebra'), 'formulas_practical', 'Практическая задача', 'Текстовая задача на составление линейного уравнения')
ON CONFLICT (topic_id, code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 12, t.id, s.id, 'formulas_subst', 'Подстановка в формулу', 'Вычисление по готовой формуле'
FROM topics t, subtopics s 
WHERE t.code='algebra' AND s.code='formulas_subst'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 12, t.id, s.id, 'formulas_units', 'Перевод единиц', 'Перевод единиц измерения'
FROM topics t, subtopics s 
WHERE t.code='algebra' AND s.code='formulas_units'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 12, t.id, s.id, 'formulas_practical', 'Практическая задача', 'Решение практических задач по формулам'
FROM topics t, subtopics s 
WHERE t.code='algebra' AND s.code='formulas_practical'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

-- ============================================
-- ЗАДАНИЕ 13: Неравенства и системы неравенств
-- ============================================

INSERT INTO subtopics (topic_id, code, title, description) VALUES
((SELECT id FROM topics WHERE code='algebra'), 'ineq_linear', 'Линейное неравенство', 'Решение неравенств вида ax + b > 0, ответ промежутком'),
((SELECT id FROM topics WHERE code='algebra'), 'ineq_quadratic', 'Квадратное неравенство', 'Метод интервалов, ветви параболы'),
((SELECT id FROM topics WHERE code='algebra'), 'ineq_system', 'Система неравенств', 'Пересечение числовых промежутков'),
((SELECT id FROM topics WHERE code='algebra'), 'ineq_rational', 'Дробно-рациональное неравенство', 'Метод интервалов для дробных неравенств')
ON CONFLICT (topic_id, code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 13, t.id, s.id, 'ineq_linear', 'Линейные неравенства', 'Решение линейных неравенств'
FROM topics t, subtopics s 
WHERE t.code='algebra' AND s.code='ineq_linear'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 13, t.id, s.id, 'ineq_quadratic', 'Квадратные неравенства', 'Решение квадратных неравенств'
FROM topics t, subtopics s 
WHERE t.code='algebra' AND s.code='ineq_quadratic'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 13, t.id, s.id, 'ineq_system', 'Системы неравенств', 'Решение систем неравенств'
FROM topics t, subtopics s 
WHERE t.code='algebra' AND s.code='ineq_system'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 13, t.id, s.id, 'ineq_rational', 'Дробно-рациональные неравенства', 'Решение дробно-рациональных неравенств'
FROM topics t, subtopics s 
WHERE t.code='algebra' AND s.code='ineq_rational'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

-- ============================================
-- ЗАДАНИЕ 14: Прогрессии
-- ============================================

INSERT INTO subtopics (topic_id, code, title, description) VALUES
((SELECT id FROM topics WHERE code='algebra'), 'progression_arithm', 'Арифметическая прогрессия', 'n-й член, разность, сумма членов'),
((SELECT id FROM topics WHERE code='algebra'), 'progression_geom', 'Геометрическая прогрессия', 'n-й член, знаменатель, сумма членов')
ON CONFLICT (topic_id, code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 14, t.id, s.id, 'progression_arithm', 'Арифметическая прогрессия', 'Задачи на арифметическую прогрессию'
FROM topics t, subtopics s 
WHERE t.code='algebra' AND s.code='progression_arithm'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 14, t.id, s.id, 'progression_geom', 'Геометрическая прогрессия', 'Задачи на геометрическую прогрессию'
FROM topics t, subtopics s 
WHERE t.code='algebra' AND s.code='progression_geom'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

-- ============================================
-- ЗАДАНИЕ 15: Треугольники (ГЕОМЕТРИЯ)
-- ============================================

INSERT INTO subtopics (topic_id, code, title, description) VALUES
((SELECT id FROM topics WHERE code='geometry'), 'triangles_angles', 'Углы треугольника', 'Сумма углов, внешний угол, биссектрисы'),
((SELECT id FROM topics WHERE code='geometry'), 'triangles_pythagor', 'Теорема Пифагора', 'Прямоугольный треугольник, катеты, гипотенуза'),
((SELECT id FROM topics WHERE code='geometry'), 'triangles_area', 'Площадь треугольника', 'Формулы площади через основание и высоту, стороны, синус'),
((SELECT id FROM topics WHERE code='geometry'), 'triangles_lines', 'Замечательные линии', 'Медианы, высоты, биссектрисы и их свойства')
ON CONFLICT (topic_id, code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 15, t.id, s.id, 'triangles_angles', 'Углы треугольника', 'Задачи на углы в треугольнике'
FROM topics t, subtopics s 
WHERE t.code='geometry' AND s.code='triangles_angles'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 15, t.id, s.id, 'triangles_pythagor', 'Теорема Пифагора', 'Применение теоремы Пифагора'
FROM topics t, subtopics s 
WHERE t.code='geometry' AND s.code='triangles_pythagor'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 15, t.id, s.id, 'triangles_area', 'Площадь треугольника', 'Вычисление площади треугольника'
FROM topics t, subtopics s 
WHERE t.code='geometry' AND s.code='triangles_area'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 15, t.id, s.id, 'triangles_lines', 'Замечательные линии', 'Свойства медиан, высот, биссектрис'
FROM topics t, subtopics s 
WHERE t.code='geometry' AND s.code='triangles_lines'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

-- ============================================
-- ЗАДАНИЕ 16: Окружность и круг
-- ============================================

INSERT INTO subtopics (topic_id, code, title, description) VALUES
((SELECT id FROM topics WHERE code='geometry'), 'circle_elements', 'Элементы окружности', 'Радиус, диаметр, хорда, дуга, секущая'),
((SELECT id FROM topics WHERE code='geometry'), 'circle_angles', 'Вписанные и центральные углы', 'Градусные меры углов, опирающиеся на дугу'),
((SELECT id FROM topics WHERE code='geometry'), 'circle_tangents', 'Касательные', 'Свойства касательной, отрезки касательных из одной точки'),
((SELECT id FROM topics WHERE code='geometry'), 'circle_inscribed', 'Вписанные и описанные окружности', 'Окружности около треугольника и многоугольника'),
((SELECT id FROM topics WHERE code='geometry'), 'circle_sector', 'Площадь сектора и сегмента', 'Вычисление площади части круга по углу или дуге')
ON CONFLICT (topic_id, code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 16, t.id, s.id, 'circle_elements', 'Элементы окружности', 'Задачи на элементы окружности'
FROM topics t, subtopics s 
WHERE t.code='geometry' AND s.code='circle_elements'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 16, t.id, s.id, 'circle_angles', 'Вписанные и центральные углы', 'Углы в окружности'
FROM topics t, subtopics s 
WHERE t.code='geometry' AND s.code='circle_angles'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 16, t.id, s.id, 'circle_tangents', 'Касательные', 'Свойства касательных к окружности'
FROM topics t, subtopics s 
WHERE t.code='geometry' AND s.code='circle_tangents'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 16, t.id, s.id, 'circle_inscribed', 'Вписанные и описанные окружности', 'Окружности, вписанные и описанные около фигур'
FROM topics t, subtopics s 
WHERE t.code='geometry' AND s.code='circle_inscribed'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 16, t.id, s.id, 'circle_sector', 'Площадь сектора и сегмента', 'Вычисление площади частей круга'
FROM topics t, subtopics s 
WHERE t.code='geometry' AND s.code='circle_sector'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

-- ============================================
-- ЗАДАНИЕ 17: Четырёхугольники и многоугольники
-- ============================================

INSERT INTO subtopics (topic_id, code, title, description) VALUES
((SELECT id FROM topics WHERE code='geometry'), 'quad_properties', 'Свойства четырёхугольников', 'Параллелограмм, прямоугольник, ромб, квадрат'),
((SELECT id FROM topics WHERE code='geometry'), 'quad_trapezoid', 'Трапеция', 'Равнобедренная трапеция, средняя линия, диагонали'),
((SELECT id FROM topics WHERE code='geometry'), 'quad_area', 'Площади четырёхугольников', 'Формулы площадей через стороны, диагонали, углы'),
((SELECT id FROM topics WHERE code='geometry'), 'quad_diagonals', 'Диагонали и их свойства', 'Перпендикулярность, биссектрисы, точка пересечения')
ON CONFLICT (topic_id, code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 17, t.id, s.id, 'quad_properties', 'Свойства четырёхугольников', 'Задачи на свойства параллелограмма, прямоугольника, ромба, квадрата'
FROM topics t, subtopics s 
WHERE t.code='geometry' AND s.code='quad_properties'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 17, t.id, s.id, 'quad_trapezoid', 'Трапеция', 'Задачи на трапецию'
FROM topics t, subtopics s 
WHERE t.code='geometry' AND s.code='quad_trapezoid'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 17, t.id, s.id, 'quad_area', 'Площади четырёхугольников', 'Вычисление площадей четырёхугольников'
FROM topics t, subtopics s 
WHERE t.code='geometry' AND s.code='quad_area'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 17, t.id, s.id, 'quad_diagonals', 'Диагонали четырёхугольников', 'Свойства диагоналей'
FROM topics t, subtopics s 
WHERE t.code='geometry' AND s.code='quad_diagonals'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

-- ============================================
-- ЗАДАНИЕ 18: Фигуры на квадратной решётке
-- ============================================

INSERT INTO subtopics (topic_id, code, title, description) VALUES
((SELECT id FROM topics WHERE code='geometry'), 'grid_distance', 'Расстояние между точками', 'Координатная плоскость, расстояние'),
((SELECT id FROM topics WHERE code='geometry'), 'grid_pythagor', 'Пифагор на решётке', 'Стороны и высоты по клеткам'),
((SELECT id FROM topics WHERE code='geometry'), 'grid_area', 'Площадь на клетчатой бумаге', 'Метод Пика, подсчёт клеток'),
((SELECT id FROM topics WHERE code='geometry'), 'grid_midline', 'Средняя линия на решётке', 'Свойства средней линии треугольника и трапеции')
ON CONFLICT (topic_id, code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 18, t.id, s.id, 'grid_distance', 'Расстояние на решётке', 'Вычисление расстояния между точками на координатной плоскости'
FROM topics t, subtopics s 
WHERE t.code='geometry' AND s.code='grid_distance'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 18, t.id, s.id, 'grid_pythagor', 'Пифагор на решётке', 'Применение теоремы Пифагора на клетчатой бумаге'
FROM topics t, subtopics s 
WHERE t.code='geometry' AND s.code='grid_pythagor'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 18, t.id, s.id, 'grid_area', 'Площадь на клетчатой бумаге', 'Вычисление площади фигур на клетчатой бумаге'
FROM topics t, subtopics s 
WHERE t.code='geometry' AND s.code='grid_area'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 18, t.id, s.id, 'grid_midline', 'Средняя линия на решётке', 'Свойства средней линии на клетчатой бумаге'
FROM topics t, subtopics s 
WHERE t.code='geometry' AND s.code='grid_midline'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

-- ============================================
-- ЗАДАНИЕ 19: Верные/неверные утверждения
-- ============================================

INSERT INTO subtopics (topic_id, code, title, description) VALUES
((SELECT id FROM topics WHERE code='geometry'), 'logic_angles', 'Утверждения об углах и прямых', 'Вертикальные, смежные, параллельные, перпендикулярные прямые'),
((SELECT id FROM topics WHERE code='geometry'), 'logic_triangles', 'Утверждения о треугольниках', 'Равенство, подобие, площади, теоремы'),
((SELECT id FROM topics WHERE code='geometry'), 'logic_quad', 'Утверждения о четырёхугольниках', 'Свойства параллелограмма, трапеции, ромба, квадрата'),
((SELECT id FROM topics WHERE code='geometry'), 'logic_circle', 'Утверждения об окружности', 'Касательные, хорды, вписанные и центральные углы')
ON CONFLICT (topic_id, code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 19, t.id, s.id, 'logic_angles', 'Утверждения об углах и прямых', 'Анализ утверждений об углах и прямых'
FROM topics t, subtopics s 
WHERE t.code='geometry' AND s.code='logic_angles'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 19, t.id, s.id, 'logic_triangles', 'Утверждения о треугольниках', 'Анализ утверждений о треугольниках'
FROM topics t, subtopics s 
WHERE t.code='geometry' AND s.code='logic_triangles'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 19, t.id, s.id, 'logic_quad', 'Утверждения о четырёхугольниках', 'Анализ утверждений о четырёхугольниках'
FROM topics t, subtopics s 
WHERE t.code='geometry' AND s.code='logic_quad'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

INSERT INTO task_types (oge_number, topic_id, subtopic_id, subtype_code, title, description)
SELECT 19, t.id, s.id, 'logic_circle', 'Утверждения об окружности', 'Анализ утверждений об окружности'
FROM topics t, subtopics s 
WHERE t.code='geometry' AND s.code='logic_circle'
ON CONFLICT (oge_number, subtype_code) DO NOTHING;

COMMIT;
