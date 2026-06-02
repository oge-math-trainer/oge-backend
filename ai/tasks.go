package ai

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/oge-math-trainer/oge-backend.git/internal/diagnostic"
	"github.com/oge-math-trainer/oge-backend.git/internal/tasks"
)

//go:embed tasks_catalog.json
var embeddedTaskCatalog []byte

type catalogEntry struct {
	Title            string
	Description      string
	Example          string
	Examples         []string
	ForbiddenTopics  []string
	RequiredKeywords []string
}

type sdamgiaTaskSpec struct {
	Format             string
	SubtypeInstruction string
	Forbidden          string
	TemplateMarker     string
}

var (
	catalogOnce sync.Once
	catalogData map[string]map[string]catalogEntry
	catalogErr  error
)

var catalogAliases = map[int]map[string]string{
	6: {
		"numbers_fractions": "fractions_common",
		"numbers_decimal":   "fractions_decimal",
		"numbers_integers":  "fractions_mixed",
	},
	7: {
		"numberline_compare":  "number_line",
		"numberline_plot":     "number_line",
		"numbers_compare":     "compare_numbers",
		"inequalities_simple": "simple_inequalities",
	},
	8: {
		"algebra_powers":   "powers",
		"algebra_roots":    "roots",
		"algebra_formulas": "formulas_short",
	},
	9: {
		"equations_linear":        "eq_linear",
		"equations_quadratic":     "eq_quadratic",
		"equations_rational":      "eq_rational",
		"equations_system_linear": "eq_systems_linear",
		"equations_system_mixed":  "eq_systems_linear",
	},
	10: {
		"probability_classic": "prob_classic",
		"probability_tree":    "prob_theorems",
	},
	11: {
		"graphs_linear":        "graph_kb",
		"graphs_linear_signs":  "graph_kb",
		"graphs_quadratic":     "func_graph_match",
		"graphs_inverse":       "func_graph_match",
		"graphs_transform":     "func_graph_match",
		"graphs_match":         "graph_formula_match",
		"graphs_match_func":    "func_graph_match",
		"graphs_match_formula": "graph_formula_match",
	},
	12: {
		"formulas_subst":     "formula_calc",
		"formulas_units":     "formula_calc",
		"formulas_practical": "linear_eq_practical",
		"formula_compute":    "formula_calc",
		"formula_linear":     "linear_eq_practical",
	},
	14: {
		"progression_arithm": "prog_arith",
		"progression_geom":   "prog_geom",
	},
	15: {
		"triangles_angles":   "tri_angles",
		"triangles_pythagor": "tri_right",
		"triangles_area":     "tri_general",
		"triangles_lines":    "tri_general",
	},
	16: {
		"circle_tangents": "circle_tangent_chord",
		"circle_sector":   "circle_angles",
	},
	17: {
		"quad_properties": "quad_parallelogram",
		"quad_area":       "quad_rectangle",
		"quad_diagonals":  "quad_rhombus",
	},
	18: {
		"grid_distance": "geo_distance",
		"grid_midline":  "geo_midline",
		"grid_pythagor": "geo_pythagor",
		"grid_area":     "geo_area",
		"grid_length":   "geo_side_length",
	},
	19: {
		"logic_angles": "logic_angles_lines",
		"logic_quad":   "logic_quads",
	},
}

var validAnswerRe = regexp.MustCompile(`^-?[0-9]+(,[0-9]+)?$`)

func (e *catalogEntry) UnmarshalJSON(data []byte) error {
	var raw struct {
		Title            string            `json:"title"`
		Description      string            `json:"description"`
		Example          string            `json:"example"`
		Examples         []json.RawMessage `json:"examples"`
		ForbiddenTopics  []string          `json:"forbidden_topics"`
		RequiredKeywords []string          `json:"required_keywords"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	*e = catalogEntry{
		Title:            raw.Title,
		Description:      raw.Description,
		Example:          raw.Example,
		ForbiddenTopics:  raw.ForbiddenTopics,
		RequiredKeywords: raw.RequiredKeywords,
	}
	for _, item := range raw.Examples {
		if example := decodeCatalogExample(item); example != "" {
			e.Examples = append(e.Examples, example)
		}
	}
	return nil
}

func decodeCatalogExample(item json.RawMessage) string {
	var text string
	if err := json.Unmarshal(item, &text); err == nil {
		return strings.TrimSpace(text)
	}

	var objectExample struct {
		Question string `json:"question"`
	}
	if err := json.Unmarshal(item, &objectExample); err == nil {
		return strings.TrimSpace(objectExample.Question)
	}
	return ""
}

func loadCatalog() (map[string]map[string]catalogEntry, error) {
	catalogOnce.Do(func() {
		catalogData = make(map[string]map[string]catalogEntry)
		catalogErr = json.Unmarshal(embeddedTaskCatalog, &catalogData)
	})
	return catalogData, catalogErr
}

func getCatalogEntry(target tasks.Target) catalogEntry {
	catalog, err := loadCatalog()
	if err != nil {
		return catalogEntry{}
	}
	byNumber := catalog[fmt.Sprintf("%d", target.OgeNumber)]
	if byNumber == nil {
		return catalogEntry{}
	}
	if entry, ok := byNumber[target.SubtypeCode]; ok {
		return entry
	}
	if alias := catalogAlias(target); alias != "" {
		if entry, ok := byNumber[alias]; ok {
			return entry
		}
	}
	return catalogEntry{}
}

func catalogAlias(target tasks.Target) string {
	byNumber := catalogAliases[target.OgeNumber]
	if byNumber == nil {
		return ""
	}
	return byNumber[strings.TrimSpace(target.SubtypeCode)]
}

func validateAnswer(answer string) error {
	answer = strings.TrimSpace(answer)
	if answer == "" {
		return fmt.Errorf("correct_answer is empty")
	}
	if !validAnswerRe.MatchString(answer) {
		return fmt.Errorf("correct_answer has unsupported format: %q", answer)
	}
	return nil
}

func specSubtypeCode(target tasks.Target) string {
	subtype := strings.TrimSpace(target.SubtypeCode)
	switch subtype {
	case "fractions_common":
		return "numbers_fractions"
	case "fractions_decimal":
		return "numbers_decimal"
	case "fractions_mixed":
		return "numbers_fractions"
	case "numberline_plot", "numbers_compare":
		return "numberline_compare"
	case "powers":
		return "algebra_powers"
	case "roots":
		return "algebra_roots"
	case "fsu", "formulas_short":
		return "algebra_formulas"
	case "eq_linear":
		return "equations_linear"
	case "eq_quadratic":
		return "equations_quadratic"
	case "eq_rational":
		return "equations_rational"
	case "eq_system_linear", "eq_systems_linear":
		return "equations_system_linear"
	case "prob_classic":
		return "probability_classic"
	case "prob_theorems":
		return "probability_tree"
	case "formula_compute":
		return "formulas_subst"
	case "formula_linear":
		return "formulas_practical"
	case "tri_angles":
		return "triangles_angles"
	case "tri_right":
		return "triangles_pythagor"
	case "tri_general", "tri_isosceles":
		return "triangles_lines"
	case "circle_tangent_chord":
		return "circle_tangents"
	case "quad_parallelogram", "quad_rectangle", "quad_rhombus":
		return "quad_properties"
	case "logic_quad":
		return "logic_quads"
	default:
		return subtype
	}
}

func sdamgiaSpec(target tasks.Target) sdamgiaTaskSpec {
	subtype := specSubtypeCode(target)
	spec := sdamgiaTaskSpec{
		Forbidden:      "Не меняй номер задания и не уходи в соседние типы.",
		TemplateMarker: fmt.Sprintf("oge_%d_%s", target.OgeNumber, target.SubtypeCode),
	}

	switch target.OgeNumber {
	case 6:
		spec.Format = "Формат СДАМ ГИА №6: короткое задание на вычисление значения числового выражения."
		spec.Forbidden = "Не генерируй уравнения, неравенства, вероятность, графики или геометрию."
		switch subtype {
		case "fractions_common":
			spec.SubtypeInstruction = "Используй обыкновенные дроби, смешанные числа или дробь от числа. Ответ должен быть целым числом или десятичной дробью с запятой."
		case "fractions_decimal":
			spec.SubtypeInstruction = "Используй десятичные дроби и порядок действий. Ответ пиши десятичной дробью с запятой или целым числом."
		case "fractions_mixed":
			spec.SubtypeInstruction = "Используй целые, отрицательные и рациональные числа, но без переменных. Допускай смешанные вычисления с дробями и десятичными числами."
		}

	case 7:
		spec.Format = "Формат СДАМ ГИА №7: задача на числовую прямую, сравнение чисел или выбор варианта ответа."
		spec.Forbidden = "Не используй решение уравнений, вероятность или геометрию."
		switch subtype {
		case "simple_inequalities":
			spec.Format = "Формат ОГЭ №7: точки на координатной прямой и утверждения о них."
			spec.SubtypeInstruction = `1. Выбери ОДИН из форматов:
   			- Одна точка a (между двумя целыми числами)
   			- Две точки a и b (a слева от 0, b справа)
   			- Три точки a, b, c (в порядке возрастания)
   			- Три точки p, q, r (для сравнения разностей)
 
			2. Дай ровно 4 варианта ответа в формате:
			   1) утверждение
			   2) утверждение
			   3) утверждение
			   4) утверждение
			3. correct_answer должен быть номером варианта: "1", "2", "3" или "4"
			4. visual_data: type:"number_line" с точками и их метками`
			spec.Forbidden = "НЕ генерируй интервалы [-1; 3). НЕ проси выбрать число из множества. НЕ пиши правильный ответ в условии."
		case "compare_numbers":
			spec.SubtypeInstruction = `Дай задачу на оценку числа БЕЗ визуализации. Форматы:
		    - "Какому из данных промежутков принадлежит число sqrt{53}?" + 4 варианта промежутков
		    - "Какое из данных чисел заключено между 1/6 и 1/3?" + 4 варианта чисел
		    correct_answer должен быть номером варианта: "1", "2", "3" или "4".
		    ВАЖНО: НЕ включай visual_data. НЕ рисуй координатную прямую.`
		case "number_line":
			spec.SubtypeInstruction = "Дай числовую прямую с отмеченными точками A, B, C, D. Спроси, какая точка соответствует заданному числу (например \\sqrt{72}, π, -2.5). Дай 4 варианта ответа: 1)A 2)B 3)C 4)D. correct_answer должен быть номером варианта (1, 2, 3 или 4). visual_data должен содержать type:'number_line' с точками и их метками."
		}

	case 8:
		spec.Format = "Формат СДАМ ГИА №8: упростить или вычислить алгебраическое выражение."
		spec.Forbidden = "Не генерируй уравнение как №9 и неравенство как №13."
		switch subtype {
		case "powers":
			spec.SubtypeInstruction = "Основная операция: свойства степеней с целым показателем."
		case "roots":
			spec.SubtypeInstruction = "Основная операция: квадратные корни, вынесение множителя из-под корня, упрощение корней."
		case "formulas_short":
			spec.SubtypeInstruction = "Основная операция: формулы сокращённого умножения или разложение на множители."
		}

	case 9:
		spec.Format = "Формат СДАМ ГИА №9: решить уравнение или систему уравнений."
		spec.Forbidden = "Не используй неравенства и геометрию."
		switch subtype {
		case "eq_linear":
			spec.SubtypeInstruction = "Только линейное уравнение с одним неизвестным."
		case "eq_quadratic":
			spec.SubtypeInstruction = "Квадратное уравнение с целым корнем; ответом пусть будет один корень по условию или сумма корней."
		case "eq_rational":
			spec.SubtypeInstruction = "Дробно-рациональное уравнение с проверкой ОДЗ; ответ числом."
		case "eq_systems_linear":
			spec.SubtypeInstruction = "Система двух линейных уравнений; попроси найти x, y или их сумму. Пиши системы уравнений через \\begin{cases} <пример> \\\\ <пример> \\end{cases}"
		}

	case 10:
		spec.Format = "Формат СДАМ ГИА №10: вероятность случайного события."
		spec.Forbidden = "Не используй уравнения, неравенства, графики или геометрию."
		switch subtype {
		case "prob_classic":
			spec.SubtypeInstruction = "Используй классическую вероятность: билеты, шары, кубик, монета, карточки. Подбирай числа так, чтобы ответ был конечной десятичной дробью с запятой, например 0,25 или 0,4."
		case "prob_theorems":
			spec.SubtypeInstruction = "Используй задачи на противоположные события или сумму вероятностей. Формат: 'Вероятность, что ручка пишет плохо = 0,09. Найдите вероятность, что она пишет хорошо.' Ответ: 1 - 0,09 = 0,91. correct_answer должен быть десятичной дробью с запятой, например '0,91'."
		}

	case 11:
		spec.Format = "Формат СДАМ ГИА №11: по графикам функций установить соответствие или определить параметры."
		spec.SubtypeInstruction = "correct_answer должен быть только последовательностью цифр, например 231. Не используй буквы, дефисы и формат 1-A."
		spec.Forbidden = "Не используй физические графики движения, таблицы или текстовые задачи без графика функции."
		switch subtype {
		case "graph_kb":
			spec.SubtypeInstruction = "Дай график линейной функции y = kx + b. Спроси о знаках коэффициентов k и b. Варианты: '1) k>0, b>0', '2) k<0, b>0' и т.д. Дай 4 варианта. correct_answer - номер правильного варианта (1, 2, 3 или 4)."
		case "func_graph_match":
			spec.SubtypeInstruction = "Дай 3 функции (разных типов: линейная, парабола, гипербола) и 3 графика под метками А, Б, В. Попроси установить соответствие. correct_answer должен быть строкой из 3 цифр, например '231' где первая цифра - номер графика для первой функции. НЕ показывай ответ в условии."
		case "graph_formula_match":
			spec.SubtypeInstruction = "Дай 3 графика функций и 3 формулы. Попроси установить соответствие. correct_answer должен быть строкой из 3 цифр, например '231'. НЕ показывай ответ в условии задачи. Используй разные типы функций: линейные, параболы, гиперболы."
		}

	case 12:
		spec.Format = "Формат СДАМ ГИА №12: вычислить значение по готовой формуле из практического контекста."
		spec.Forbidden = "Не проси решить уравнение как самостоятельную цель. Не давай тривиальные подстановки без контекста."
		switch subtype {
		case "formula_calc":
			spec.SubtypeInstruction = `Дай практическую формулу из жизни с контекстом. Примеры: 'Стоимость такси: C = 150 + 11(t-5), где t — минуты (t>5). Найдите стоимость 15-минутной поездки.' или 'Площадь треугольника: S = ½ah. Найдите сторону a, если S=28, h=14.' correct_answer должен быть числом.`
		case "linear_eq_practical":
			spec.SubtypeInstruction = `Дай линейное уравнение в практическом контексте. correct_answer должен быть числом. Примеры: 
			'Площадь ромба S (в м^2) можно вычислить по формуле S = \frac{1}{2}d_1d_2 , где d_1, d_2 — диагонали ромба (в метрах). Пользуясь этой формулой, найдите диагональ d_1 если диагональ d_2 равна 30 м, а площадь ромба 120 м^2.';
			'Объем пирамиды вычисляют по формуле V = \frac{1}{3}Sh где S — площадь основания пирамиды, h — ее высота. Объем пирамиды равен 40, площадь основания 15. Чему равна высота пирамиды?'`
		}

	case 13:
		spec.Format = "Формат СДАМ ГИА №13: решить неравенство или систему неравенств и выбрать подходящий вариант ответа."
		switch subtype {
		case "ineq_linear":
			spec.SubtypeInstruction = `Решение неравенства первой степени и выбор верного варианта. Дай 4 варианта ответа с промежутками. correct_answer - номер правильного варианта одной цифрой. Пример:
			'Решите неравенство 5 - 4(x - 2) < 22 - x\n1)(-3; +∞)\n2)(-∞; -\frac{1}{3})\n3)(-\frac{1}{3}; +∞)\n4)(-∞; -3)'; 
			'Решите неравенство 9x + 8 > 8x - 8.\n1)(-∞; -16)\n2)(-16; +∞)\n3)(-∞; 0)\n4)(0; +∞)'`
		case "ineq_quadratic":
			spec.SubtypeInstruction = `Метод интервалов, анализ знаков квадратного трёхчлена. Дай 4 варианта ответа с промежутками. correct_answer - номер правильного варианта одной цифрой.`
		case "ineq_system":
			spec.SubtypeInstruction = `Решение систем линейных и квадратных неравенств. системы уравнений записывай через \begin{cases} <пример> \\ <пример> \end{cases}. Дай 4 варианта ответа с промежутками. correct_answer - номер правильного варианта одной цифрой.
			НИКАКИХ РИСУНКОВ, варианты ответов записывай текстом`
		}

	case 14:
		spec.Format = "Формат СДАМ ГИА №14: арифметическая или геометрическая прогрессия, часто практическая текстовая задача."
		spec.Forbidden = "Не генерируй задачи на вероятность, уравнения или геометрию."
		switch subtype {
		case "prog_geom":
			spec.SubtypeInstruction = "Используй геометрическую прогрессию: знаменатель, последовательное изменение, рост или уменьшение."
		case "prog_arith":
			spec.SubtypeInstruction = "Используй арифметическую прогрессию: разность, номер члена или сумма первых членов."
		}

	case 15:
		spec.Format = "Формат СДАМ ГИА №15: планиметрия по треугольнику, обычно с чертежом."
		spec.Forbidden = "Не используй окружность как главную тему №16 и четырёхугольники как №17."
		switch subtype {
		case "tri_angles":
			spec.SubtypeInstruction = "Работай с углами треугольника, внешним углом или равнобедренным треугольником."
		case "tri_right":
			spec.SubtypeInstruction = "Работай с прямоугольным треугольником и теоремой Пифагора. Подбирай пифагоровы тройки, чтобы correct_answer был целым числом без приближений."
		case "tri_isosceles":
			spec.SubtypeInstruction = "Работай с равнобедренным треугольником: свойства углов при основании, медиана/биссектриса/высота к основанию."
		case "tri_general":
			spec.SubtypeInstruction = "Работай с медианой, биссектрисой, высотой или средней линией треугольника."
		}

	case 16:
		spec.Format = "Формат СДАМ ГИА №16: окружность и её элементы, обязательно с чертежом."
		spec.Forbidden = "Не генерируй задачу только про треугольник, трапецию или прямоугольник без окружности."
		switch subtype {
		case "circle_angles":
			spec.SubtypeInstruction = "Строго задача про ВПИСАННЫЕ и ЦЕНТРАЛЬНЫЕ углы. В условии и visual_data должна быть ОКРУЖНОСТЬ с центром O, хорды/дуги/углы. НЕ рисуй треугольники без окружности."
		case "circle_inscribed":
			spec.SubtypeInstruction = "ОКРУЖНОСТЬ ВПИСАНА в многоугольник (касается всех сторон). Найди радиус, сторону или периметр. Подбирай числа так, чтобы ответ был целым или простой десятичной дробью."
		case "circle_circumscribed":
			spec.SubtypeInstruction = "МНОГОУГОЛЬНИК ВПИСАН в окружность (все вершины лежат на окружности). Найди радиус, сторону или угол. НЕ путай с вписанной окружностью."
		case "circle_tangent_chord":
			spec.TemplateMarker = "circle_tangent_chord tangent_from_external_point"
			spec.SubtypeInstruction = "Из внешней точки A к окружности проведены касательные AB и AC или одна касательная AT; радиус к точке касания перпендикулярен касательной; найти радиус, расстояние до центра или длину касательной."
		}

	case 17:
		spec.Format = "Формат СДАМ ГИА №17: свойства, диагонали или площадь четырёхугольника, обычно с чертежом."
		spec.Forbidden = "Не генерируй задачу про одну окружность как №16 и клетчатую бумагу как №18."
		switch subtype {
		case "quad_trapezoid":
			spec.SubtypeInstruction = "Работай с трапецией, основаниями, средней линией или высотой."
		case "quad_parallelogram":
			spec.SubtypeInstruction = "Строго задача про ПАРАЛЛЕЛОГРАММ. Используй свойства: противоположные стороны равны, диагонали делятся пополам."
		case "quad_rectangle":
			spec.SubtypeInstruction = "Строго задача про ПРЯМОУГОЛЬНИК. Используй свойства: диагонали равны, все углы 90°, площадь = a·b. НЕ генерируй треугольники."
		case "quad_rhombus":
			spec.SubtypeInstruction = "Строго задача про РОМБ. Все стороны равны, диагонали перпендикулярны и делят углы пополам. visual_data должен отражать равные стороны и перпендикулярные диагонали."
		}

	case 18:
		spec.Format = "Формат СДАМ ГИА №18: геометрия на клетчатой бумаге."
		spec.Forbidden = "Не используй обычный чертёж без клетчатой решётки."
		switch subtype {
		case "geo_distance":
			spec.SubtypeInstruction = "На клетчатой бумаге отмечены две точки. Найди длину отрезка между ними."
		case "geo_midline":
			spec.SubtypeInstruction = "На клетчатой бумаге изображён ТРЕУГОЛЬНИК или ТРАПЕЦИЯ. Нужно найти длину СРЕДНЕЙ ЛИНИИ. НЕ расстояние между двумя точками!"
		case "geo_pythagor":
			spec.SubtypeInstruction = "На клетчатой бумаге дан прямоугольный треугольник или фигура, где нужно применить теорему Пифагора для нахождения длины отрезка."
		case "geo_area":
			spec.SubtypeInstruction = "На клетчатой бумаге изображена фигура. Найди её площадь (например, методом Пика или разбиением)."
		case "geo_side_length":
			spec.SubtypeInstruction = "На клетчатой бумаге даны координаты вершин. Найди длину стороны многоугольника."
		}

	case 19:
		spec.Format = "Формат СДАМ ГИА №19: выбрать номера верных геометрических утверждений."
		spec.SubtypeInstruction = "НЕ начинай условие с фраз 'На рисунке изображены...'. Дай ровно 3 утверждения. correct_answer - только цифры верных утверждений в возрастающем порядке, например '13'. НЕ усложняй конструкциями с пересекающимися хордами в точке О."
		spec.Forbidden = "Не проси вычислить длину, площадь или угол; это проверка истинности утверждений."
	}

	return spec
}

func buildGeneratePrompt(target tasks.Target, previousError string) string {
	return buildGeneratePromptWithSeed(target, previousError, "")
}

func buildGeneratePromptWithSeed(target tasks.Target, previousError, variationSeed string) string {
	entry := getCatalogEntry(target)
	spec := sdamgiaSpec(target)
	example := selectCatalogExample(entry, variationSeed)

	var forbidden []string
	forbidden = append(forbidden, spec.Forbidden)
	forbidden = append(forbidden, entry.ForbiddenTopics...)

	var required []string
	required = append(required, entry.RequiredKeywords...)
	required = append(required, spec.SubtypeInstruction)

	visualRule := visualPromptRule(target)
	retryRule := ""
	if strings.TrimSpace(previousError) != "" {
		retryRule = "\nPrevious attempt was rejected. Fix this problem: " + truncateFeedback(previousError)
	}
	seedRule := ""
	if strings.TrimSpace(variationSeed) != "" {
		seedRule = "\nVariation seed: " + strings.TrimSpace(variationSeed)
	}

	return fmt.Sprintf(`Generate exactly one new Russian OGE math task in SDAM GIA / math-oge.sdamgia.ru style.

Target:
- oge_number: %d
- subtype_code: %s
- subtype title: %s
- subtype description: %s
- template marker: %s

MANDATORY: Your task MUST be structurally identical to this example. Same topic, same operation type, different numbers only:
%s

If you generate a task about marinades, cooking, food, or any non-mathematical real-world topic unrelated to the subtype above — it will be REJECTED.

Novelty requirements:
- Create a genuinely new task for this generation.
- Do not reuse the reference example's wording, story, numbers, graph coordinates, answer, or visual_data.
- If a previous attempt was rejected as a duplicate/conflict, change the scenario and all key numeric values, not just punctuation.
- Prefer simple integer arithmetic and one unambiguous checked answer.%s

Arithmetic consistency rules:
- Solve the task completely before returning JSON.
- correct_answer must be exactly the final numeric result from solution_steps.
- The last solution step must explicitly end with "Ответ: <correct_answer>." using the same value as correct_answer.
- Do not use approximate answers. If the natural answer is a fraction, irrational value, or rounded number, change the task numbers so the answer fits the allowed correct_answer regex.
- For probability tasks, return a decimal probability between 0 and 1 with comma as decimal separator, not "m,n" for a fraction m/n.
- A separate worker reviewer will reject the task if arithmetic, options, visual_data, or correct_answer are inconsistent.

Required SDAM GIA format:
%s

Subtype requirements:
%s

Forbidden:
%s

Return ONLY valid JSON with this shape:
{
  "question": "problem text in Russian",
  "correct_answer": "12",
  "solution_steps": ["step 1", "step 2", "step 3"],
  "self_check": "how to verify the answer",
  "is_valid": true,
  "validation_notes": ""
}
If visual_data is required below, include it as an additional JSON field named "visual_data"; otherwise omit it.

Strict answer rules:
- correct_answer must match ^-?[0-9]+(,[0-9]+)?$
- If the answer is a set/order of options or statements, write only digits without spaces, for example "231" or "13".
- Never return "1-A", "1) A", fractions like "2/3", words, units, or explanations in correct_answer.

GLOBAL RULE: The task must be about pure mathematics matching the subtype. Never invent real-world scenarios (marinades, cooking, speed of cars, biology) unless the subtype explicitly requires practical tasks like formulas_practical or progression_geom.

Text rules:
- All user-facing text must be in Russian.
- Use LaTeX for formulas and escape JSON backslashes correctly.
- Do not put TikZ, PGFPlots, \draw, \node, \foreach, \begin{axis}, or other drawing code in question, solution_steps, or self_check. Diagrams must be described only in visual_data.
- Create a new task in the same format; do not copy SDAM GIA examples verbatim.

%s%s`,
		target.OgeNumber,
		target.SubtypeCode,
		nonEmpty(entry.Title, "selected subtype"),
		nonEmpty(entry.Description, "no catalog description"),
		spec.TemplateMarker,
		nonEmpty(example, "no catalog example"),
		seedRule,
		spec.Format,
		strings.Join(nonEmptySlice(required), "\n"),
		strings.Join(nonEmptySlice(forbidden), "\n"),
		visualRule,
		retryRule,
	)
}

func selectCatalogExample(entry catalogEntry, variationSeed string) string {
	examples := catalogExamples(entry)
	if len(examples) == 0 {
		return ""
	}
	if strings.TrimSpace(variationSeed) == "" {
		return examples[0]
	}
	h := fnv.New32a()
	_, _ = h.Write([]byte(variationSeed))
	return examples[int(h.Sum32())%len(examples)]
}

func catalogExamples(entry catalogEntry) []string {
	seen := make(map[string]struct{})
	examples := make([]string, 0, len(entry.Examples)+1)
	for _, value := range append([]string{entry.Example}, entry.Examples...) {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		key := strings.ToLower(value)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		examples = append(examples, value)
	}
	return examples
}

func visualPromptRule(target tasks.Target) string {
	switch tasks.VisualKindForTarget(target) {
	case tasks.VisualKindGraph:
		return `Required visual_data:
Add "visual_data" using this schema. For smooth curves provide at least 11 evenly spaced points across the x range.
Example for parabola y=x^2 with 11 points:
{"type":"graph","x_axis":{"min":-5,"max":5},"y_axis":{"min":-2,"max":10},"plots":[{"id":"A","label":"Graph A","points":[{"x":-5,"y":25},{"x":-4,"y":16},{"x":-3,"y":9},{"x":-2,"y":4},{"x":-1,"y":1},{"x":0,"y":0},{"x":1,"y":1},{"x":2,"y":4},{"x":3,"y":9},{"x":4,"y":16},{"x":5,"y":25}]}]}
For OGE 11 matching tasks create exactly 3 separate plots as separate graph pictures: A, B, C. Each plot minimum 11 points. Do not combine curves.`
	case tasks.VisualKindNumberLine:
		return `Required visual_data:
Add "visual_data" using this schema:
{"type":"number_line","axis":{"min":-5,"max":5},"interval":{"start":-1,"end":3,"start_closed":true,"end_closed":false},"points":[{"value":-1,"closed":true,"label":"-1"},{"value":3,"closed":false,"label":"3"}]}
The interval and points must match the condition and the correct option.`
	case tasks.VisualKindGeometry:
		return `Required visual_data:
Add "visual_data" using this schema:
{"type":"geometry","shape":"triangle","vertices":[{"label":"A","x":0,"y":0},{"label":"B","x":6,"y":0},{"label":"C","x":0,"y":8}],"labels":{"A":"A","B":"B","C":"C"},"segments":[{"from":"A","to":"B","label":"6"},{"from":"A","to":"C","label":"8"},{"from":"B","to":"C"}]}
IMPORTANT: For circle tasks (subtype contains 'circle'), set shape to "circle", add "center":{"label":"O","x":0,"y":0}, "radius":5, and relevant points/chords/angles. For quadrilaterals, set shape to "quadrilateral" with 4 vertices. visual_data.type must always be "geometry".`
	case tasks.VisualKindGrid:
		return `Required visual_data:
Add "visual_data" using this schema:
{"type":"grid","width":8,"height":6,"points":[{"label":"A","x":1,"y":1},{"label":"B","x":6,"y":1},{"label":"C","x":6,"y":4}],"segments":[{"from":"A","to":"B"},{"from":"B","to":"C"},{"from":"C","to":"A"}],"fill":["A","B","C"]}
All point coordinates must be grid nodes.`
	default:
		return `Do not include visual_data and do not include graphs for this task type.`
	}
}

func nonEmpty(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func nonEmptySlice(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			out = append(out, "- "+value)
		}
	}
	if len(out) == 0 {
		return []string{"- no additional requirements"}
	}
	return out
}

// func sanitizeAIResponse(raw string) string {
// 	raw = strings.TrimSpace(raw)
// 	raw = strings.TrimPrefix(raw, "```json")
// 	raw = strings.TrimPrefix(raw, "```")
// 	raw = strings.TrimSuffix(raw, "```")

// 	start := strings.Index(raw, "{")
// 	end := strings.LastIndex(raw, "}")
// 	if start >= 0 && end > start {
// 		raw = raw[start : end+1]
// 	}

// 	raw = strings.ReplaceAll(raw, `\ \ `, "")
// 	raw = strings.ReplaceAll(raw, `\ \`, "")
// 	raw = strings.ReplaceAll(raw, `\(`, "")
// 	raw = strings.ReplaceAll(raw, `\)`, "")
// 	raw = strings.ReplaceAll(raw, `\[`, "")
// 	raw = strings.ReplaceAll(raw, `\]`, "")
// 	raw = strings.ReplaceAll(raw, `\\(`, "")
// 	raw = strings.ReplaceAll(raw, `\\)`, "")
// 	raw = strings.ReplaceAll(raw, `\\[`, "")
// 	raw = strings.ReplaceAll(raw, `\\]`, "")
// 	raw = strings.ReplaceAll(raw, `\begin{cases}`, "")
// 	raw = strings.ReplaceAll(raw, `\end{cases}`, "")
// 	raw = strings.ReplaceAll(raw, `\\begin{cases}`, "")
// 	raw = strings.ReplaceAll(raw, `\\end{cases}`, "")
// 	raw = strings.ReplaceAll(raw, `\begin`, "")
// 	raw = strings.ReplaceAll(raw, `\end`, "")

// 	reFrac := regexp.MustCompile(`\\frac\{([^}]+)\}\{([^}]+)\}`)
// 	raw = reFrac.ReplaceAllString(raw, "$1/$2")
// 	raw = regexp.MustCompile(`\\cdot`).ReplaceAllString(raw, "*")
// 	reSqrt := regexp.MustCompile(`\\sqrt\{([^}]+)\}`)
// 	raw = reSqrt.ReplaceAllString(raw, "sqrt($1)")
// 	raw = regexp.MustCompile(`\\{1,2}([a-zA-Z0-9{}()^/_\-+*=])`).ReplaceAllString(raw, "$1")
// 	raw = strings.ReplaceAll(raw, `\\`, "")
// 	raw = strings.ReplaceAll(raw, `\`, "")
// 	raw = strings.ReplaceAll(raw, "$", "")
// 	raw = regexp.MustCompile(`\s+`).ReplaceAllString(raw, " ")

// 	return strings.TrimSpace(raw)
// }

func (c *Client) IsConfigured() bool {
	return c != nil && strings.TrimSpace(c.apiKey) != "" && strings.TrimSpace(c.baseURL) != "" && strings.TrimSpace(c.model) != ""
}

func (c *Client) GenerateTask(ctx context.Context, target tasks.Target, feedback string) (tasks.GeneratedContent, error) {
	if target.OgeNumber < 6 || target.OgeNumber > 19 {
		return tasks.GeneratedContent{}, fmt.Errorf("GenerateTask: oge_number must be between 6 and 19")
	}

	systemPrompt := "Return only valid JSON. No markdown. No text outside JSON."
	var lastErr error

	for attempt := 1; attempt <= 3; attempt++ {
		select {
		case <-ctx.Done():
			return tasks.GeneratedContent{}, ctx.Err()
		default:
		}

		prompt := buildGeneratePromptWithSeed(target, generationFeedback(feedback, lastErr), generationVariationSeed(target, attempt))
		raw, err := c.Chat(systemPrompt, prompt, generationMaxTokens(target), generationTemperature(target))
		if err != nil {
			lastErr = err
			continue
		}

		var result tasks.GeneratedContent
		if err := decodeAIJSON(raw, &result); err != nil {
			lastErr = fmt.Errorf("could not parse AI JSON: %w", err)
			continue
		}

		normalizeGeneratedContent(&result)
		if err := validateGeneratedTask(target, &result); err != nil {
			lastErr = err
			continue
		}

		return result, nil
	}

	if lastErr != nil {
		return tasks.GeneratedContent{}, fmt.Errorf("GenerateTask: %w", lastErr)
	}
	return tasks.GeneratedContent{}, fmt.Errorf("GenerateTask: AI returned no valid task")
}

func generationFeedback(feedback string, err error) string {
	feedback = strings.TrimSpace(feedback)
	errText := errorText(err)
	switch {
	case feedback != "" && errText != "":
		return feedback + "; " + errText
	case feedback != "":
		return feedback
	default:
		return errText
	}
}

func errorText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func generationVariationSeed(target tasks.Target, attempt int) string {
	return fmt.Sprintf("%d:%s:%d:%d", target.OgeNumber, target.SubtypeCode, attempt, time.Now().UnixNano())
}

func generationMaxTokens(target tasks.Target) int {
	if tasks.RequiresVisualData(target) {
		return 3200
	}
	return 2200
}

func generationTemperature(target tasks.Target) float64 {
	if tasks.RequiresVisualData(target) {
		return 0.25
	}
	return 0.15
}

func truncateFeedback(value string) string {
	value = strings.TrimSpace(value)
	if len(value) <= 700 {
		return value
	}
	return value[:700] + "..."
}

func (c *Client) ReviewGeneratedTask(ctx context.Context, target tasks.Target, content tasks.GeneratedContent) (tasks.GenerationReview, error) {
	select {
	case <-ctx.Done():
		return tasks.GenerationReview{}, ctx.Err()
	default:
	}

	systemPrompt := `You are a strict verifier for Russian OGE math tasks.
Return only valid JSON. No markdown. No text outside JSON.
Check mathematical correctness, subtype fit, answer format, options, and visual_data consistency.`

	type reviewInput struct {
		Target  tasks.Target           `json:"target"`
		Content tasks.GeneratedContent `json:"content"`
	}
	inputBytes, err := json.Marshal(reviewInput{
		Target:  target,
		Content: content,
	})
	if err != nil {
		return tasks.GenerationReview{}, fmt.Errorf("ReviewGeneratedTask: %w", err)
	}

	prompt := fmt.Sprintf(`Review this generated OGE math task before it is saved by the worker.

Input JSON:
%s

Accept only if all checks pass:
- The problem has exactly one correct answer.
- content.correct_answer exactly matches the mathematically correct final answer.
- solution_steps are mathematically valid and end with the same final answer.
- correct_answer matches ^-?[0-9]+(,[0-9]+)?$ and uses comma for decimals.
- For probability tasks, correct_answer is a decimal from 0 to 1, not a numerator/denominator pair.
- For OGE 13, the question must contain answer options and correct_answer must be the correct option number.
- For OGE 19, the question must contain exactly 3 statements and correct_answer must contain the numbers of all true statements in increasing order.
- If visual_data is present or required, it must match the question and use the expected type.

Return exactly this JSON:
{"is_valid":true,"reason":"","correct_answer":"12"}

If invalid, return:
{"is_valid":false,"reason":"short concrete reason in Russian","correct_answer":"correct value if known, otherwise empty string"}`, string(inputBytes))

	raw, err := c.Chat(systemPrompt, prompt, 700, 0)
	if err != nil {
		return tasks.GenerationReview{}, fmt.Errorf("ReviewGeneratedTask: %w", err)
	}

	var review tasks.GenerationReview
	if err := decodeAIJSON(raw, &review); err != nil {
		return tasks.GenerationReview{}, fmt.Errorf("ReviewGeneratedTask: could not parse AI JSON: %w", err)
	}
	review.Reason = strings.TrimSpace(review.Reason)
	review.CorrectAnswer = strings.TrimSpace(review.CorrectAnswer)
	return review, nil
}

func decodeAIJSON(raw string, dest any) error {
	cleaned := extractJSONObject(raw)
	if err := json.Unmarshal([]byte(cleaned), dest); err == nil {
		return nil
	}
	return json.Unmarshal([]byte(repairJSONStringEscapes(cleaned)), dest)
}

func extractJSONObject(raw string) string {
	raw = strings.TrimSpace(strings.TrimPrefix(raw, "\ufeff"))
	if strings.HasPrefix(raw, "```") {
		lines := strings.Split(raw, "\n")
		if len(lines) > 0 && strings.HasPrefix(strings.TrimSpace(lines[0]), "```") {
			lines = lines[1:]
		}
		if len(lines) > 0 && strings.HasPrefix(strings.TrimSpace(lines[len(lines)-1]), "```") {
			lines = lines[:len(lines)-1]
		}
		raw = strings.TrimSpace(strings.Join(lines, "\n"))
	}

	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start >= 0 && end > start {
		return strings.TrimSpace(raw[start : end+1])
	}
	return raw
}

func repairJSONStringEscapes(raw string) string {
	var b strings.Builder
	b.Grow(len(raw) + 16)

	inString := false
	for i := 0; i < len(raw); i++ {
		ch := raw[i]
		if !inString {
			b.WriteByte(ch)
			if ch == '"' {
				inString = true
			}
			continue
		}

		switch ch {
		case '"':
			b.WriteByte(ch)
			inString = false
		case '\\':
			if i+1 >= len(raw) {
				b.WriteString(`\\`)
				continue
			}
			next := raw[i+1]
			if next == '"' || next == '\\' || next == '/' || validUnicodeEscape(raw, i+1) {
				b.WriteByte(ch)
				b.WriteByte(next)
				i++
				continue
			}
			b.WriteString(`\\`)
		case '\n', '\r':
			b.WriteByte(' ')
		case '\t':
			b.WriteByte(' ')
		default:
			b.WriteByte(ch)
		}
	}

	return b.String()
}

func validUnicodeEscape(raw string, escapeIndex int) bool {
	if escapeIndex >= len(raw) || raw[escapeIndex] != 'u' || escapeIndex+4 >= len(raw) {
		return false
	}
	for _, ch := range raw[escapeIndex+1 : escapeIndex+5] {
		if (ch < '0' || ch > '9') && (ch < 'a' || ch > 'f') && (ch < 'A' || ch > 'F') {
			return false
		}
	}
	return true
}

func normalizeGeneratedContent(result *tasks.GeneratedContent) {
	result.Question = cleanText(result.Question)
	result.CorrectAnswer = strings.TrimSpace(result.CorrectAnswer)
	result.SelfCheck = cleanText(result.SelfCheck)
	result.ValidationNotes = strings.TrimSpace(result.ValidationNotes)
	if len(result.VisualData) == 0 {
		result.VisualData = nil
	}
	for i, step := range result.SolutionSteps {
		result.SolutionSteps[i] = cleanText(step)
	}
}

func cleanText(s string) string {
	s = strings.TrimSpace(s)
	// убираем \left и \right
	s = strings.ReplaceAll(s, `\left`, "")
	s = strings.ReplaceAll(s, `\right`, "")
	// // убираем \cdot → *
	// s = strings.ReplaceAll(s, `\cdot`, "*")
	// // убираем \ldots → ...
	// s = strings.ReplaceAll(s, `\ldots`, "...")
	// // убираем \infty → бесконечность
	// s = strings.ReplaceAll(s, `\infty`, "бесконечность")
	// // убираем \pm → ±
	// s = strings.ReplaceAll(s, `\pm`, "±")
	// // убираем \neq → ≠
	// s = strings.ReplaceAll(s, `\neq`, "≠")
	// // убираем \leq → <=
	// s = strings.ReplaceAll(s, `\leq`, "<=")
	// // убираем \geq → >=
	// s = strings.ReplaceAll(s, `\geq`, ">=")
	return s
}

func validateGeneratedTask(target tasks.Target, result *tasks.GeneratedContent) error {
	if result == nil {
		return fmt.Errorf("generated content is empty")
	}
	if strings.TrimSpace(result.Question) == "" {
		return fmt.Errorf("question is empty")
	}
	if err := validateAnswer(result.CorrectAnswer); err != nil {
		return err
	}
	if len(result.SolutionSteps) < 3 {
		return fmt.Errorf("solution_steps must contain at least 3 steps")
	}
	if !result.IsValid {
		return fmt.Errorf("is_valid=false")
	}
	if err := tasks.ValidateUserFacingText(*result); err != nil {
		return err
	}
	// if containsForbiddenMathSyntax(result.Question) || containsForbiddenMathSyntax(strings.Join(result.SolutionSteps, " ")) {
	// 	return fmt.Errorf("text contains LaTeX or forbidden math syntax")
	// }
	if tasks.RequiresVisualData(target) {
		visualData, err := tasks.NormalizeAndValidateVisualData(target, result)
		if err != nil {
			return err
		}
		result.VisualData = visualData
	} else {
		result.VisualData = nil
		result.Graphs = nil
	}
	if err := validateTaskTheme(target, result.Question); err != nil {
		return err
	}
	return nil
}

// func containsForbiddenMathSyntax(text string) bool {
// 	return strings.Contains(text, `\`) || strings.Contains(text, "$")
// }

func validateTaskTheme(target tasks.Target, question string) error {
	text := strings.ToLower(question)
	switch target.OgeNumber {
	case 9:
		if strings.Contains(text, "неравен") {
			return fmt.Errorf("№9 must be an equation task, not an inequality")
		}
	case 10:
		if strings.Contains(text, "решите уравнение") || strings.Contains(text, "неравенство") || strings.Contains(text, "график") {
			return fmt.Errorf("№10 must be a probability task")
		}
	case 13:
		if strings.Contains(text, "решите уравнение") && !strings.Contains(text, "неравен") {
			return fmt.Errorf("№13 must be an inequality task")
		}
	case 16:
		if target.SubtypeCode == "circle_tangents" && !strings.Contains(text, "касател") {
			return fmt.Errorf("№16 circle_tangents must be about tangents")
		}
		if !strings.Contains(text, "окруж") && !strings.Contains(text, "касател") && !strings.Contains(text, "хорд") && !strings.Contains(text, "дуг") && !strings.Contains(text, "сектор") {
			return fmt.Errorf("№16 must be about a circle")
		}
	case 19:
		if !strings.Contains(text, "утвержден") && !strings.Contains(text, "верн") {
			return fmt.Errorf("№19 must ask for true statements")
		}
	}
	return nil
}

func (c *Client) CheckAnswer(ctx context.Context, task tasks.Task, studentAnswer string) (tasks.CheckResult, error) {
	systemPrompt := `You are a mathematics teacher checking a student's answer to an OGE exam problem.
Respond ONLY with valid JSON. No markdown, no code blocks. Start with { and end with }.
Return exactly this structure:
{
  "is_correct": true or false,
  "short_feedback": "one sentence in Russian, what the student did right or wrong",
  "reason": "brief explanation in Russian why the answer is correct or incorrect"
}
Rules:
- All text in Russian
- is_correct must be boolean
- Be strict: only accept mathematically correct answers
- Normalize answers: treat comma and dot as decimal separator, ignore leading zeros.`

	type checkInput struct {
		Question      string `json:"question"`
		CorrectAnswer string `json:"correct_answer"`
		StudentAnswer string `json:"student_answer"`
	}
	inputBytes, err := json.Marshal(checkInput{
		Question:      task.Question,
		CorrectAnswer: task.CorrectAnswer,
		StudentAnswer: studentAnswer,
	})
	if err != nil {
		return tasks.CheckResult{}, fmt.Errorf("CheckAnswer: %w", err)
	}

	raw, err := c.Chat(systemPrompt, string(inputBytes), 300, 0.2)
	if err != nil {
		return tasks.CheckResult{}, fmt.Errorf("CheckAnswer: %w", err)
	}
	var result tasks.CheckResult
	if err := decodeAIJSON(raw, &result); err != nil {
		return tasks.CheckResult{}, fmt.Errorf("CheckAnswer: could not parse AI JSON: %w", err)
	}

	return result, nil
}

func (c *Client) Hint(ctx context.Context, task tasks.Task) (tasks.HintResult, error) {
	systemPrompt := `You are a supportive mathematics tutor helping a 9th-grade Russian student.
The student gave a wrong answer and needs a gentle nudge.
Respond ONLY with valid JSON. No markdown, no code blocks. Start with { and end with }.
Return exactly this structure:
{
  "hint": "a guiding question in Russian that points toward the right approach without revealing the answer"
}
Rules:
- All text in Russian
- The hint must be a question, not a statement
- Do NOT reveal the answer or the full solution method
- Be specific to this problem, not generic advice
- Maximum 2 sentences.`

	type hintInput struct {
		Question string `json:"question"`
	}
	inputBytes, err := json.Marshal(hintInput{Question: task.Question})
	if err != nil {
		return tasks.HintResult{}, fmt.Errorf("Hint: %w", err)
	}

	raw, err := c.Chat(systemPrompt, string(inputBytes), 300, 0.6)
	if err != nil {
		return tasks.HintResult{}, fmt.Errorf("Hint: %w", err)
	}
	var parsed struct {
		Hint string `json:"hint"`
	}
	if err := decodeAIJSON(raw, &parsed); err != nil {
		return tasks.HintResult{}, fmt.Errorf("Hint: could not parse AI JSON: %w", err)
	}

	return tasks.HintResult{Hint: strings.TrimSpace(parsed.Hint)}, nil
}

func (c *Client) Explain(ctx context.Context, task tasks.Task) (tasks.ExplainResult, error) {
	systemPrompt := `You are an expert mathematics teacher explaining the full solution of an OGE exam problem.
The student failed to solve it. Give a complete step-by-step explanation.
Respond ONLY with valid JSON. No markdown, no code blocks. Start with { and end with }.
Return exactly this structure:
{
  "explanation": "full explanation in Russian: what is given, what is required, the method used",
  "steps": ["step 1 with calculation", "step 2 with calculation", "step 3 with calculation"]
}
Rules:
- All text in Russian
- steps minimum 3, maximum 7
- Each step must show the calculation explicitly
- Use REALLY simple language for a 9th-grade student, like he barely understands maths
- Say hello at the start`

	type explainInput struct {
		Question      string `json:"question"`
		CorrectAnswer string `json:"correct_answer"`
	}
	inputBytes, err := json.Marshal(explainInput{
		Question:      task.Question,
		CorrectAnswer: task.CorrectAnswer,
	})
	if err != nil {
		return tasks.ExplainResult{}, fmt.Errorf("Explain: %w", err)
	}

	raw, err := c.Chat(systemPrompt, string(inputBytes), 600, 0.4)
	if err != nil {
		return tasks.ExplainResult{}, fmt.Errorf("Explain: %w", err)
	}
	var result tasks.ExplainResult
	if err := decodeAIJSON(raw, &result); err != nil {
		return tasks.ExplainResult{}, fmt.Errorf("Explain: could not parse AI JSON: %w", err)
	}

	return result, nil
}

func (c *Client) AnalyzeDiagnostic(ctx context.Context, answers []diagnostic.AnswerResult) (diagnostic.Analysis, error) {
	select {
	case <-ctx.Done():
		return diagnostic.Analysis{}, ctx.Err()
	default:
	}

	systemPrompt := `Ты анализируешь диагностику ученика по ОГЭ. Верни только валидный JSON без markdown.`
	payload, err := json.Marshal(answers)
	if err != nil {
		return diagnostic.Analysis{}, fmt.Errorf("AnalyzeDiagnostic: %w", err)
	}

	prompt := fmt.Sprintf(`Проанализируй результаты диагностики по заданиям ОГЭ №6-19.
Ответы: %s
Верни строго JSON:
{"summary":"краткий вывод на русском","weak_topics":["тема 1","тема 2"]}`, string(payload))

	raw, err := c.Chat(systemPrompt, prompt, 600, 0.2)
	if err != nil {
		return diagnostic.Analysis{}, fmt.Errorf("AnalyzeDiagnostic: %w", err)
	}
	var result diagnostic.Analysis
	if err := decodeAIJSON(raw, &result); err != nil {
		return diagnostic.Analysis{}, fmt.Errorf("AnalyzeDiagnostic: could not parse AI JSON: %w", err)
	}
	result.Summary = strings.TrimSpace(result.Summary)
	for i := range result.WeakTopics {
		result.WeakTopics[i] = strings.TrimSpace(result.WeakTopics[i])
	}
	return result, nil
}
