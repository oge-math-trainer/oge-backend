package ai

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/oge-math-trainer/oge-backend.git/internal/diagnostic"
	"github.com/oge-math-trainer/oge-backend.git/internal/tasks"
)

//go:embed tasks_catalog.json
var embeddedTaskCatalog []byte

type exampleTask struct {
	Question string `json:"question"`
}

type catalogEntry struct {
	Title            string        `json:"title"`
	Description      string        `json:"description"`
	Example          string        `json:"example"`
	Examples         []exampleTask `json:"examples"`
	ForbiddenTopics  []string      `json:"forbidden_topics"`
	RequiredKeywords []string      `json:"required_keywords"`
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

var validAnswerRe = regexp.MustCompile(`^-?[0-9]+(,[0-9]+)?$`)

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
	return byNumber[target.SubtypeCode]
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

func sdamgiaSpec(target tasks.Target) sdamgiaTaskSpec {
	spec := sdamgiaTaskSpec{
		Forbidden:      "Не меняй номер задания и не уходи в соседние типы.",
		TemplateMarker: fmt.Sprintf("oge_%d_%s", target.OgeNumber, target.SubtypeCode),
	}

	switch target.OgeNumber {
	case 6:
		spec.Format = "Формат СДАМ ГИА №6: короткое задание на вычисление значения числового выражения."
		spec.Forbidden = "Не генерируй уравнения, неравенства, вероятность, графики или геометрию."
		switch target.SubtypeCode {
		case "numbers_fractions":
			spec.SubtypeInstruction = "Используй обыкновенные дроби, смешанные числа или дробь от числа. Ответ должен быть целым числом или десятичной дробью с запятой."
		case "numbers_decimal":
			spec.SubtypeInstruction = "Используй десятичные дроби и порядок действий. Ответ пиши десятичной дробью с запятой или целым числом."
		case "numbers_integers":
			spec.SubtypeInstruction = "Используй целые, отрицательные и рациональные числа, но без переменных."
		}
	case 7:
		spec.Format = "Формат СДАМ ГИА №7: числовая прямая, сравнение чисел или выбор подходящего варианта."
		spec.SubtypeInstruction = "Дай 3-4 варианта ответа или точки на числовой прямой. correct_answer должен быть одним числом или номером варианта."
		spec.Forbidden = "Не используй решение уравнений, вероятность или геометрию."
	case 8:
		spec.Format = "Формат СДАМ ГИА №8: упростить или вычислить алгебраическое выражение."
		spec.Forbidden = "Не генерируй уравнение как №9 и неравенство как №13."
		switch target.SubtypeCode {
		case "algebra_powers":
			spec.SubtypeInstruction = "Основная операция: свойства степеней с целым показателем."
		case "algebra_roots":
			spec.SubtypeInstruction = "Основная операция: квадратные корни, вынесение множителя из-под корня, упрощение корней."
		case "algebra_formulas":
			spec.SubtypeInstruction = "Основная операция: формулы сокращённого умножения или разложение на множители."
		case "algebra_fractions":
			spec.SubtypeInstruction = "Основная операция: алгебраические дроби, сокращение или вычисление при заданном значении переменной."
		}
	case 9:
		spec.Format = "Формат СДАМ ГИА №9: решить уравнение или систему уравнений."
		spec.Forbidden = "Не используй неравенства и геометрию."
		switch target.SubtypeCode {
		case "equations_linear":
			spec.SubtypeInstruction = "Только линейное уравнение с одним неизвестным."
		case "equations_quadratic":
			spec.SubtypeInstruction = "Квадратное уравнение с целым корнем; ответом пусть будет один корень по условию или сумма корней."
		case "equations_rational":
			spec.SubtypeInstruction = "Дробно-рациональное уравнение с проверкой ОДЗ; ответ числом."
		case "equations_system_linear":
			spec.SubtypeInstruction = "Система двух линейных уравнений; попроси найти x, y или их сумму."
		case "equations_system_mixed":
			spec.SubtypeInstruction = "Система с линейным и квадратным уравнением; попроси найти одно числовое значение."
		}
	case 10:
		spec.Format = "Формат СДАМ ГИА №10: вероятность случайного события."
		spec.Forbidden = "Не используй уравнения, неравенства, графики или геометрию."
		switch target.SubtypeCode {
		case "probability_classic":
			spec.SubtypeInstruction = "Используй классическую вероятность: билеты, шары, кубик, монета, карточки."
		case "probability_tree":
			spec.SubtypeInstruction = "Используй два независимых или последовательных случайных события, но без рисунка дерева."
		}
	case 11:
		spec.Format = "Формат СДАМ ГИА №11: по графикам функций установить соответствие или определить параметры."
		spec.SubtypeInstruction = "correct_answer должен быть только последовательностью цифр, например 231. Не используй буквы, дефисы и формат 1-A."
		spec.Forbidden = "Не используй физические графики движения, таблицы или текстовые задачи без графика функции."
	case 12:
		spec.Format = "Формат СДАМ ГИА №12: вычислить значение по готовой формуле."
		spec.Forbidden = "Не проси решить уравнение как самостоятельную цель."
		switch target.SubtypeCode {
		case "formulas_subst":
			spec.SubtypeInstruction = "Дай формулу и значения переменных; нужно вычислить результат."
		case "formulas_units":
			spec.SubtypeInstruction = "Дай формулу с единицами измерения или перевод величин."
		case "formulas_practical":
			spec.SubtypeInstruction = "Дай прикладную формулу из жизни: стоимость, работа, скорость, проценты."
		}
	case 13:
		spec.Format = "Формат СДАМ ГИА №13: решить неравенство или систему неравенств и выбрать подходящий вариант ответа."
		spec.SubtypeInstruction = "Дай 4 варианта ответа с промежутками. correct_answer - номер правильного варианта одной цифрой."
		spec.Forbidden = "Не используй уравнения без знаков неравенства."
	case 14:
		spec.Format = "Формат СДАМ ГИА №14: арифметическая или геометрическая прогрессия, часто практическая текстовая задача."
		if target.SubtypeCode == "progression_geom" {
			spec.SubtypeInstruction = "Используй геометрическую прогрессию: знаменатель, последовательное изменение, рост или уменьшение."
		} else {
			spec.SubtypeInstruction = "Используй арифметическую прогрессию: разность, номер члена или сумма первых членов."
		}
		spec.Forbidden = "Не генерируй задачи на вероятность, уравнения или геометрию."
	case 15:
		spec.Format = "Формат СДАМ ГИА №15: планиметрия по треугольнику, обычно с чертежом."
		spec.Forbidden = "Не используй окружность как главную тему №16 и четырёхугольники как №17."
		switch target.SubtypeCode {
		case "triangles_angles":
			spec.SubtypeInstruction = "Работай с углами треугольника, внешним углом или равнобедренным треугольником."
		case "triangles_pythagor":
			spec.SubtypeInstruction = "Работай с прямоугольным треугольником и теоремой Пифагора."
		case "triangles_area":
			spec.SubtypeInstruction = "Работай с площадью треугольника, основанием и высотой."
		case "triangles_lines":
			spec.SubtypeInstruction = "Работай с медианой, биссектрисой, высотой или средней линией треугольника."
		}
	case 16:
		spec.Format = "Формат СДАМ ГИА №16: окружность и её элементы, обычно с чертежом."
		spec.Forbidden = "Не генерируй задачу только про треугольник, трапецию или прямоугольник."
		switch target.SubtypeCode {
		case "circle_elements":
			spec.SubtypeInstruction = "Работай с радиусом, диаметром, хордой, расстоянием от центра до хорды, длиной окружности или площадью круга."
		case "circle_angles":
			spec.SubtypeInstruction = "Работай с вписанным или центральным углом и дугой."
		case "circle_tangents":
			spec.TemplateMarker = "circle_tangents tangent_from_external_point"
			spec.SubtypeInstruction = "Обязательно формат касательных: из внешней точки A к окружности проведены касательные AB и AC или одна касательная AT; радиус к точке касания перпендикулярен касательной; найти радиус, расстояние до центра или длину касательной."
		case "circle_inscribed":
			spec.SubtypeInstruction = "Работай с вписанной или описанной окружностью."
		case "circle_sector":
			spec.SubtypeInstruction = "Работай с сектором, сегментом, дугой или центральным углом."
		}
	case 17:
		spec.Format = "Формат СДАМ ГИА №17: свойства, диагонали или площадь четырёхугольника, обычно с чертежом."
		spec.Forbidden = "Не генерируй задачу про одну окружность как №16 и клетчатую бумагу как №18."
		switch target.SubtypeCode {
		case "quad_properties":
			spec.SubtypeInstruction = "Работай с параллелограммом, ромбом, прямоугольником или квадратом."
		case "quad_trapezoid":
			spec.SubtypeInstruction = "Работай с трапецией, основаниями, средней линией или высотой."
		case "quad_area":
			spec.SubtypeInstruction = "Работай с площадями четырёхугольников."
		case "quad_diagonals":
			spec.SubtypeInstruction = "Работай с диагоналями прямоугольника, ромба, квадрата или параллелограмма."
		}
	case 18:
		spec.Format = "Формат СДАМ ГИА №18: геометрия на клетчатой бумаге."
		spec.SubtypeInstruction = "Используй координаты точек на решётке; нужно найти расстояние, площадь, длину или среднюю линию."
		spec.Forbidden = "Не используй обычный чертёж без клетчатой решётки."
	case 19:
		spec.Format = "Формат СДАМ ГИА №19: выбрать номера верных геометрических утверждений."
		spec.SubtypeInstruction = "Дай ровно 3 утверждения. correct_answer - только цифры верных утверждений в возрастающем порядке, например 13. Не используй формат 1-A."
		spec.Forbidden = "Не проси вычислить длину, площадь или угол; это проверка истинности утверждений."
	}

	return spec
}

func buildGeneratePrompt(target tasks.Target, previousError string) string {
	entry := getCatalogEntry(target)
	spec := sdamgiaSpec(target)
	example := entry.Example
	if example == "" && len(entry.Examples) > 0 {
		example = entry.Examples[0].Question
	}

	var forbidden []string
	forbidden = append(forbidden, spec.Forbidden)
	forbidden = append(forbidden, entry.ForbiddenTopics...)

	var required []string
	required = append(required, entry.RequiredKeywords...)
	required = append(required, spec.SubtypeInstruction)

	visualRule := visualPromptRule(target)
	retryRule := ""
	if strings.TrimSpace(previousError) != "" {
		retryRule = "\nPrevious attempt was rejected. Fix this problem: " + previousError
	}

	return fmt.Sprintf(`Generate exactly one new Russian OGE math task in SDAM GIA style.

Target:
- oge_number: %d
- subtype_code: %s
- subtype title: %s
- subtype description: %s
- template marker: %s

Reference example structure, do not copy it verbatim:
%s

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

Strict answer rules:
- correct_answer must match ^-?[0-9]+(,[0-9]+)?$
- If the answer is a set/order of options or statements, write only digits without spaces, for example "231" or "13".
- Never return "1-A", "1) A", fractions like "2/3", words, units, or explanations in correct_answer.

Text rules:
- All user-facing text must be in Russian.
- Use LaTeX.
- Create a new task in the same format; do not copy SDAM GIA examples verbatim.

%s%s`,
		target.OgeNumber,
		target.SubtypeCode,
		nonEmpty(entry.Title, "selected subtype"),
		nonEmpty(entry.Description, "no catalog description"),
		spec.TemplateMarker,
		nonEmpty(example, "no catalog example"),
		spec.Format,
		strings.Join(nonEmptySlice(required), "\n"),
		strings.Join(nonEmptySlice(forbidden), "\n"),
		visualRule,
		retryRule,
	)
}

func visualPromptRule(target tasks.Target) string {
	switch tasks.VisualKindForTarget(target) {
	case tasks.VisualKindGraph:
		return `Required visual_data:
Add "visual_data" using this schema:
{"type":"graph","x_axis":{"min":-10,"max":10},"y_axis":{"min":-10,"max":10},"graphs":[{"id":"1","label":"y = x","points":[{"x":-2,"y":-2},{"x":0,"y":0},{"x":2,"y":2}]}]}
Use 1 to 4 graph series. Each series must have at least 3 coordinate points.`
	case tasks.VisualKindNumberLine:
		return `Required visual_data:
Add "visual_data" using this schema:
{"type":"number_line","axis":{"min":-5,"max":5},"interval":{"start":-1,"end":3,"start_closed":true,"end_closed":false},"points":[{"value":-1,"closed":true,"label":"-1"},{"value":3,"closed":false,"label":"3"}]}
The interval and points must match the condition and the correct option.`
	case tasks.VisualKindGeometry:
		return `Required visual_data:
Add "visual_data" using this schema:
{"type":"geometry","shape":"triangle","vertices":[{"label":"A","x":0,"y":0},{"label":"B","x":6,"y":0},{"label":"C","x":0,"y":8}],"labels":{"A":"A","B":"B","C":"C"},"segments":[{"from":"A","to":"B","label":"6"},{"from":"A","to":"C","label":"8"},{"from":"B","to":"C"}]}
For circle tasks use shape "circle_tangent", "circle_angles", "circle_chord" or "circle_sector" and you may add circles:[{"center":"O","through":"B"}].`
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

func sanitizeAIResponse(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")

	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start >= 0 && end > start {
		raw = raw[start : end+1]
	}

	raw = strings.ReplaceAll(raw, `\ \ `, "")
	raw = strings.ReplaceAll(raw, `\ \`, "")
	raw = strings.ReplaceAll(raw, `\(`, "")
	raw = strings.ReplaceAll(raw, `\)`, "")
	raw = strings.ReplaceAll(raw, `\[`, "")
	raw = strings.ReplaceAll(raw, `\]`, "")
	raw = strings.ReplaceAll(raw, `\\(`, "")
	raw = strings.ReplaceAll(raw, `\\)`, "")
	raw = strings.ReplaceAll(raw, `\\[`, "")
	raw = strings.ReplaceAll(raw, `\\]`, "")
	raw = strings.ReplaceAll(raw, `\begin{cases}`, "")
	raw = strings.ReplaceAll(raw, `\end{cases}`, "")
	raw = strings.ReplaceAll(raw, `\\begin{cases}`, "")
	raw = strings.ReplaceAll(raw, `\\end{cases}`, "")
	raw = strings.ReplaceAll(raw, `\begin`, "")
	raw = strings.ReplaceAll(raw, `\end`, "")

	reFrac := regexp.MustCompile(`\\frac\{([^}]+)\}\{([^}]+)\}`)
	raw = reFrac.ReplaceAllString(raw, "$1/$2")
	raw = regexp.MustCompile(`\\cdot`).ReplaceAllString(raw, "*")
	reSqrt := regexp.MustCompile(`\\sqrt\{([^}]+)\}`)
	raw = reSqrt.ReplaceAllString(raw, "sqrt($1)")
	raw = regexp.MustCompile(`\\{1,2}([a-zA-Z0-9{}()^/_\-+*=])`).ReplaceAllString(raw, "$1")
	raw = strings.ReplaceAll(raw, `\\`, "")
	raw = strings.ReplaceAll(raw, `\`, "")
	raw = strings.ReplaceAll(raw, "$", "")
	raw = regexp.MustCompile(`\s+`).ReplaceAllString(raw, " ")

	return strings.TrimSpace(raw)
}

func (c *Client) IsConfigured() bool {
	return c != nil && strings.TrimSpace(c.apiKey) != "" && strings.TrimSpace(c.baseURL) != "" && strings.TrimSpace(c.model) != ""
}

func (c *Client) GenerateTask(ctx context.Context, target tasks.Target) (tasks.GeneratedContent, error) {
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

		prompt := buildGeneratePrompt(target, errorText(lastErr))
		raw, err := c.Chat(systemPrompt, prompt, 2200, 0.25)
		if err != nil {
			lastErr = err
			continue
		}

		raw = sanitizeAIResponse(raw)

		var result tasks.GeneratedContent
		if err := json.Unmarshal([]byte(raw), &result); err != nil {
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

func errorText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func normalizeGeneratedContent(result *tasks.GeneratedContent) {
	result.Question = strings.TrimSpace(result.Question)
	result.CorrectAnswer = strings.TrimSpace(result.CorrectAnswer)
	result.SelfCheck = strings.TrimSpace(result.SelfCheck)
	result.ValidationNotes = strings.TrimSpace(result.ValidationNotes)
	if len(result.VisualData) == 0 {
		result.VisualData = nil
	}
	for i, step := range result.SolutionSteps {
		result.SolutionSteps[i] = strings.TrimSpace(step)
	}
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
	if containsForbiddenMathSyntax(result.Question) || containsForbiddenMathSyntax(strings.Join(result.SolutionSteps, " ")) {
		return fmt.Errorf("text contains LaTeX or forbidden math syntax")
	}
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

func containsForbiddenMathSyntax(text string) bool {
	return strings.Contains(text, `\`) || strings.Contains(text, "$")
}

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
- Normalize answers: treat comma and dot as decimal separator, ignore leading zeros
- Do NOT use LaTeX, arrows, or special math symbols. Use plain text only.`

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
	raw = sanitizeAIResponse(raw)

	var result tasks.CheckResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
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
- Maximum 2 sentences
- Do NOT use LaTeX, arrows, or special math symbols. Use plain text only.`

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
	raw = sanitizeAIResponse(raw)

	var parsed struct {
		Hint string `json:"hint"`
	}
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
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
- Use simple language for a 9th-grade student
- Do NOT use LaTeX, arrows, or special math symbols. Use plain text only.`

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
	raw = sanitizeAIResponse(raw)

	var result tasks.ExplainResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
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
	raw = sanitizeAIResponse(raw)

	var result diagnostic.Analysis
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return diagnostic.Analysis{}, fmt.Errorf("AnalyzeDiagnostic: could not parse AI JSON: %w", err)
	}
	result.Summary = strings.TrimSpace(result.Summary)
	for i := range result.WeakTopics {
		result.WeakTopics[i] = strings.TrimSpace(result.WeakTopics[i])
	}
	return result, nil
}
