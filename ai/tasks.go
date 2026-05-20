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
	AnswerFormat       string // Новый: явный формат ответа
}

var (
	catalogOnce sync.Once
	catalogData map[string]map[string]catalogEntry
	catalogErr  error
)

// Расширенная валидация ответов: целые, десятичные, последовательности цифр
var validAnswerRe = regexp.MustCompile(`^-?[0-9]+(,[0-9]+)?$|^[0-9]{2,4}$`)

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

func validateAnswer(target tasks.Target, answer string) error {
	answer = strings.TrimSpace(answer)
	if answer == "" {
		return fmt.Errorf("correct_answer is empty")
	}

	// Для заданий на соответствие/утверждения допускаем только цифры 2-4 знака
	if target.OgeNumber == 11 || target.OgeNumber == 19 {
		if matched, _ := regexp.MatchString(`^[0-9]{2,4}$`, answer); !matched {
			return fmt.Errorf("answer for task #%d must be 2-4 digits only, got: %q", target.OgeNumber, answer)
		}
		return nil
	}

	// Для остальных: целое или десятичное с запятой
	if !validAnswerRe.MatchString(answer) {
		return fmt.Errorf("correct_answer has unsupported format: %q (expected: integer, decimal with comma, or 2-4 digits for matching)", answer)
	}
	return nil
}

func sdamgiaSpec(target tasks.Target) sdamgiaTaskSpec {
	spec := sdamgiaTaskSpec{
		Forbidden:      "Не меняй номер задания и не уходи в соседние типы.",
		TemplateMarker: fmt.Sprintf("oge_%d_%s", target.OgeNumber, target.SubtypeCode),
		AnswerFormat:   "correct_answer: целое число или десятичная дробь с запятой",
	}

	// === Уточнения по форматам ответов ===
	switch target.OgeNumber {
	case 11:
		spec.AnswerFormat = "correct_answer: только последовательность цифр без пробелов, например 231 или 142"
	case 19:
		spec.AnswerFormat = "correct_answer: только цифры верных утверждений в возрастающем порядке, например 13 или 2"
	}

	// === Детализация по заданиям ===
	switch target.OgeNumber {
	case 6:
		spec.Format = "Задание №6: вычисление значения числового выражения."
		spec.Forbidden = "Не генерируй уравнения, неравенства, вероятность, графики или геометрию."
		switch target.SubtypeCode {
		case "numbers_fractions":
			spec.SubtypeInstruction = "Используй обыкновенные дроби, смешанные числа. Ответ — целое или десятичная дробь с запятой."
		case "numbers_decimal":
			spec.SubtypeInstruction = "Десятичные дроби, порядок действий. Ответ — десятичная дробь с запятой или целое."
		}
	case 7:
		spec.Format = "Задание №7: числовая прямая, сравнение чисел."
		spec.SubtypeInstruction = "Дай 3-4 варианта ответа или точки на числовой прямой. correct_answer — номер варианта или число."
		spec.Forbidden = "Не используй уравнения, вероятность, геометрию."
	case 8:
		spec.Format = "Задание №8: упрощение алгебраического выражения."
		spec.Forbidden = "Не генерируй уравнения как №9 и неравенства как №13."
		// ... подтипы
	case 9:
		spec.Format = "Задание №9: решить уравнение."
		spec.Forbidden = "Не используй неравенства и геометрию."
		// ... подтипы
	case 10:
		spec.Format = "Задание №10: вероятность."
		// ...
	case 11:
		spec.Format = "Задание №11: соответствие графиков функций."
		spec.SubtypeInstruction = "correct_answer — последовательность цифр (например 231). Не используй буквы, дефисы, формат 1-A."
		spec.Forbidden = "Не используй физические графики, таблицы, текстовые задачи без функции."
		// ВАЖНО: визуальные данные
		// (обрабатывается в visualPromptRule)
	case 12:
		spec.Format = "Задание №12: вычисление по формуле."
		// ...
	case 13:
		spec.Format = "Задание №13: решение неравенства."
		spec.SubtypeInstruction = "Дай 4 варианта с промежутками. correct_answer — номер варианта одной цифрой."
		// ...
	// ... остальные номера аналогично, с уточнениями
	}

	// === Геометрия: особые правила ===
	if target.OgeNumber >= 15 && target.OgeNumber <= 18 {
		spec.Forbidden += " Не меняй тип фигуры и не добавляй лишние элементы."
	}

	// === Задание 19: утверждения ===
	if target.OgeNumber == 19 {
		spec.Format = "Задание №19: выбор верных утверждений."
		spec.SubtypeInstruction = "Дай ровно 3 утверждения. correct_answer — цифры верных в возрастающем порядке, например 13."
		spec.Forbidden = "Не проси вычислить длину, площадь или угол."
	}

	return spec
}

func buildGeneratePrompt(target tasks.Target, previousError string) string {
	entry := getCatalogEntry(target)
	spec := sdamgiaSpec(target)
	
	// Берем пример, если есть
	example := entry.Example
	if example == "" && len(entry.Examples) > 0 {
		example = entry.Examples[0].Question
	}

	// Собираем списки запрещённого и требуемого
	var forbidden, required []string
	forbidden = append(forbidden, spec.Forbidden)
	forbidden = append(forbidden, entry.ForbiddenTopics...)
	required = append(required, entry.RequiredKeywords...)
	required = append(required, spec.SubtypeInstruction)

	// Визуальные данные
	visualRule := visualPromptRule(target)
	
	// Правило для повтора после ошибки
	retryRule := ""
	if strings.TrimSpace(previousError) != "" {
		retryRule = "\n❗ ПРЕДЫДУЩАЯ ПОПЫТКА ОТКЛОНЕНА. Исправь: " + previousError
	}

	// === ШАБЛОН ПРОМПТА (упрощённый и структурированный) ===
	return fmt.Sprintf(`[СИСТЕМНАЯ ИНСТРУКЦИЯ]
Ты — генератор заданий ОГЭ по математике в стиле СДАМ ГИА.
Верни ТОЛЬКО валидный JSON. Без markdown, без текста вне {}.

[ЦЕЛЬ]
Сгенерируй ОДНО новое задание:
• Номер: %d
• Подтип: %s (%s)
• Шаблон: %s

[ПРИМЕР СТРУКТУРЫ — не копируй текст, только формат]
Пример условия: %s

[ФОРМАТ ЗАДАНИЯ]
• Стиль: %s
• Требования подтипа: %s
• Запрещено: %s
• Формат ответа: %s

[ВИЗУАЛЬНЫЕ ДАННЫЕ]
%s

[СТРОГАЯ JSON-СХЕМА ОТВЕТА]
{
  "question": "строка на русском с условием",
  "correct_answer": "строка: число, десятичная дробь с запятой, или последовательность цифр",
  "solution_steps": ["шаг 1", "шаг 2", "шаг 3"],
  "self_check": "как ученик может проверить себя",
  "is_valid": true,
  "validation_notes": "",
  "visual_data": {} // только если требуется, см. выше
}

[ПРАВИЛА ДЛЯ correct_answer]
• Целое: "12", отрицательное: "-5"
• Десятичное: "3,14" (запятая, не точка!)
• Для №11, №19: только цифры подряд, например "231" или "13"
• НИКАКИХ: дробей 2/3, букв, пробелов, пояснений, единиц измерения

[ПРАВИЛА ДЛЯ ТЕКСТА]
• Весь текст условия — на русском
• Формулы — в формате простой строки (можно LaTeX, но без $)
• Не копируй примеры из каталога дословно

[ПОВТОР]
%s

Начинай ответ сразу с {`,
		// Подстановка переменных
		target.OgeNumber,
		target.SubtypeCode,
		nonEmpty(entry.Title, "выбранный подтип"),
		spec.TemplateMarker,
		nonEmpty(example, "нет примера в каталоге"),
		spec.Format,
		strings.Join(nonEmptySlice(required), "; "),
		strings.Join(nonEmptySlice(forbidden), "; "),
		spec.AnswerFormat,
		visualRule,
		retryRule,
	)
}

// === ВИЗУАЛЬНЫЕ ДАННЫЕ: структурированные инструкции ===
func visualPromptRule(target tasks.Target) string {
	kind := tasks.VisualKindForTarget(target)
	
	switch kind {
	case tasks.VisualKindGraph:
		return `• visual_data.type: "graph"
• axes: x_axis{min,max}, y_axis{min,max} — целые числа
• graphs: массив серий, каждая с:
  - id: строка
  - label: формула, например "y = x² - 2"
  - points: минимум 3 точки [{x:число, y:число}, ...]
• Все точки должны точно лежать на заявленной функции
• Пример: {"type":"graph","x_axis":{"min":-5,"max":5},"y_axis":{"min":-10,"max":10},"graphs":[{"id":"1","label":"y = x²","points":[{"x":-2,"y":4},{"x":0,"y":0},{"x":2,"y":4}]}]}`

	case tasks.VisualKindNumberLine:
		return `• visual_data.type: "number_line"
• axis: {min, max} — целые
• interval: {start, end, start_closed:bool, end_closed:bool}
• points: массив [{value:число, closed:bool, label:"строка"}]
• Пример: {"type":"number_line","axis":{"min":-5,"max":5},"interval":{"start":-2,"end":3,"start_closed":true,"end_closed":false},"points":[{"value":-2,"closed":true,"label":"A"},{"value":3,"closed":false,"label":"B"}]}`

	case tasks.VisualKindGeometry:
		return `• visual_data.type: "geometry"
• shape: "triangle", "trapezoid", "circle_tangent" и т.д.
• vertices: массив [{label:"A", x:число, y:число}, ...] — минимум 3
• segments: [{from:"A", to:"B", label:"6"}, ...] — опционально
• circles: [{center:"O", through:"B"}] — для окружностей
• Все координаты должны образовывать корректную фигуру (без пересечений сторон)
• Пример: {"type":"geometry","shape":"triangle","vertices":[{"label":"A","x":0,"y":0},{"label":"B","x":6,"y":0},{"label":"C","x":0,"y":8}],"segments":[{"from":"A","to":"B","label":"6"},{"from":"A","to":"C","label":"8"}]}`

	case tasks.VisualKindGrid:
		return `• visual_data.type: "grid"
• width, height: целые (размер решётки)
• points: [{label:"A", x:целое, y:целое}, ...] — только узлы сетки
• segments, fill: опционально
• Пример: {"type":"grid","width":8,"height":6,"points":[{"label":"A","x":1,"y":1},{"label":"B","x":6,"y":1}],"segments":[{"from":"A","to":"B"}]}`

	default:
		return `• visual_data: не добавляй (для этого типа не требуется)`
	}
}

// === ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ ===
func nonEmpty(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func nonEmptySlice(values []string) []string {
	out := make([]string, 0, len(values))
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v != "" {
			out = append(out, v)
		}
	}
	if len(out) == 0 {
		return []string{"нет доп. требований"}
	}
	return out
}

// === САНИТИЗАЦИЯ: раскомментировано и усилено ===
func sanitizeAIResponse(raw string) string {
	raw = strings.TrimSpace(raw)
	
	// Убираем markdown-блоки
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	
	// Вырезаем только JSON-объект
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start >= 0 && end > start {
		raw = raw[start : end+1]
	}
	
	// Убираем лишние слеши и математический синтаксис, который ломает JSON
	replacements := []struct{ from, to string }{
		{`\ \ `, ""}, {`\ \`, ""}, {`\(`, ""}, {`\)`, ""},
		{`\[`, ""}, {`\]`, ""}, {`\\(`, ""}, {`\\)`, ""},
		{`\\[`, ""}, {`\\]`, ""}, {`\begin{cases}`, ""}, {`\end{cases}`, ""},
		{`\\begin`, ""}, {`\\end`, ""}, {`\begin`, ""}, {`\end`, ""},
		{`\\frac`, `\frac`}, {`\\sqrt`, `\sqrt`}, {`\\cdot`, `*`},
		{`\\`, `\`}, // оставляем один слэш для экранирования
	}
	for _, r := range replacements {
		raw = strings.ReplaceAll(raw, r.from, r.to)
	}
	
	// Убираем $ (математические делимитеры), если они есть
	raw = strings.ReplaceAll(raw, "$", "")
	
	// Нормализуем пробелы
	raw = regexp.MustCompile(`\s+`).ReplaceAllString(raw, " ")
	
	return strings.TrimSpace(raw)
}

// === КЛИЕНТ ===
type Client struct {
	apiKey  string
	baseURL string
	model   string
}

func (c *Client) IsConfigured() bool {
	return c != nil && strings.TrimSpace(c.apiKey) != "" && 
		strings.TrimSpace(c.baseURL) != "" && strings.TrimSpace(c.model) != ""
}

func (c *Client) GenerateTask(ctx context.Context, target tasks.Target) (tasks.GeneratedContent, error) {
	if target.OgeNumber < 6 || target.OgeNumber > 19 {
		return tasks.GeneratedContent{}, fmt.Errorf("oge_number must be 6..19")
	}

	systemPrompt := "Return ONLY valid JSON. No markdown. No text outside {}. Start with { and end with }."
	
	// Температура ниже для строгих типов
	temperature := 0.15
	if target.OgeNumber == 11 || target.OgeNumber == 19 {
		temperature = 0.1 // минимальная креативность для соответствий/утверждений
	}
	
	var lastErr error

	for attempt := 1; attempt <= 3; attempt++ {
		select {
		case <-ctx.Done():
			return tasks.GeneratedContent{}, ctx.Err()
		default:
		}

		prompt := buildGeneratePrompt(target, errorText(lastErr))
		
		// Запрос к ИИ
		raw, err := c.Chat(systemPrompt, prompt, 2200, temperature)
		if err != nil {
			lastErr = fmt.Errorf("API error (attempt %d): %w", attempt, err)
			continue
		}

		// Санитизация
		raw = sanitizeAIResponse(raw)

		// Парсинг JSON
		var result tasks.GeneratedContent
		if err := json.Unmarshal([]byte(raw), &result); err != nil {
			lastErr = fmt.Errorf("JSON parse error: %w (raw: %.100s...)", err, raw)
			continue
		}

		// Нормализация
		normalizeGeneratedContent(&result)
		
		// Валидация
		if err := validateGeneratedTask(target, &result); err != nil {
			lastErr = fmt.Errorf("validation error: %w", err)
			continue
		}

		// Математическая проверка визуальных данных
		if tasks.RequiresVisualData(target) {
			if err := validateVisualMath(target, &result); err != nil {
				lastErr = fmt.Errorf("visual math error: %w", err)
				continue
			}
		}

		return result, nil
	}

	return tasks.GeneratedContent{}, fmt.Errorf("all 3 attempts failed: %w", lastErr)
}

// === НОВАЯ: математическая валидация визуальных данных ===
func validateVisualMath(target tasks.Target, result *tasks.GeneratedContent) error {
	if result.VisualData == nil {
		return nil
	}
	
	// Парсим visual_data
	var visual map[string]interface{}
	if err := json.Unmarshal(result.VisualData, &visual); err != nil {
		return fmt.Errorf("invalid visual_data JSON: %w", err)
	}
	
	typ, _ := visual["type"].(string)
	
	switch typ {
	case "graph":
		// Проверяем, что точки лежат на функции (упрощённо)
		// В продакшене: парсим label и проверяем y ≈ f(x)
		graphs, _ := visual["graphs"].([]interface{})
		for _, g := range graphs {
			gm, _ := g.(map[string]interface{})
			points, _ := gm["points"].([]interface{})
			if len(points) < 3 {
				return fmt.Errorf("graph must have at least 3 points")
			}
			// Здесь можно добавить проверку: все x уникальны, y соответствуют формуле
		}
		
	case "number_line":
		// Проверяем, что start < end
		interval, _ := visual["interval"].(map[string]interface{})
		if interval != nil {
			start, _ := interval["start"].(float64)
			end, _ := interval["end"].(float64)
			if start >= end {
				return fmt.Errorf("number_line: start (%v) must be < end (%v)", start, end)
			}
		}
		
	case "geometry":
		// Проверяем, что вершин >= 3
		verts, _ := visual["vertices"].([]interface{})
		if len(verts) < 3 {
			return fmt.Errorf("geometry: at least 3 vertices required")
		}
		// Можно добавить проверку на вырожденные фигуры (все точки на одной линии)
	}
	
	return nil
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
	
	for i := range result.SolutionSteps {
		result.SolutionSteps[i] = strings.TrimSpace(result.SolutionSteps[i])
	}
}

func validateGeneratedTask(target tasks.Target, result *tasks.GeneratedContent) error {
	if result == nil {
		return fmt.Errorf("generated content is nil")
	}
	
	if strings.TrimSpace(result.Question) == "" {
		return fmt.Errorf("question is empty")
	}
	
	if err := validateAnswer(target, result.CorrectAnswer); err != nil {
		return err
	}
	
	if len(result.SolutionSteps) < 3 {
		return fmt.Errorf("solution_steps must have at least 3 steps, got %d", len(result.SolutionSteps))
	}
	
	if !result.IsValid {
		return fmt.Errorf("is_valid=false: %s", result.ValidationNotes)
	}
	
	// Проверка темы (чтобы не было "уравнение" в задании №10 на вероятность)
	if err := validateTaskTheme(target, result.Question); err != nil {
		return err
	}
	
	// Визуальные данные
	if tasks.RequiresVisualData(target) {
		if result.VisualData == nil {
			return fmt.Errorf("visual_data is required for task #%d", target.OgeNumber)
		}
		// Нормализация и валидация структуры
		normalized, err := tasks.NormalizeAndValidateVisualData(target, result)
		if err != nil {
			return fmt.Errorf("visual_data validation: %w", err)
		}
		result.VisualData = normalized
	} else {
		result.VisualData = nil
		result.Graphs = nil
	}
	
	return nil
}

func validateTaskTheme(target tasks.Target, question string) error {
	text := strings.ToLower(question)
	
	switch target.OgeNumber {
	case 9:
		if strings.Contains(text, "неравен") {
			return fmt.Errorf("№9 must be equation, not inequality")
		}
	case 10:
		if strings.Contains(text, "решите уравнение") || strings.Contains(text, "неравенство") {
			return fmt.Errorf("№10 must be probability task")
		}
	case 13:
		if strings.Contains(text, "уравнение") && !strings.Contains(text, "неравен") {
			return fmt.Errorf("№13 must be inequality task")
		}
	case 16:
		if target.SubtypeCode == "circle_tangents" && !strings.Contains(text, "касател") {
			return fmt.Errorf("№16 circle_tangents must mention tangents")
		}
		if !strings.Contains(text, "окруж") && !strings.Contains(text, "касател") && !strings.Contains(text, "хорд") {
			return fmt.Errorf("№16 must be about circle")
		}
	case 19:
		if !strings.Contains(text, "утвержден") && !strings.Contains(text, "верн") {
			return fmt.Errorf("№19 must ask for true/false statements")
		}
	}
	
	return nil
}

// === ПРОВЕРКА ОТВЕТА УЧЕНИКА ===
func (c *Client) CheckAnswer(ctx context.Context, task tasks.Task, studentAnswer string) (tasks.CheckResult, error) {
	systemPrompt := `Ты — учитель математики, проверяющий ответ ученика на ОГЭ.
Верни ТОЛЬКО валидный JSON. Без markdown. Начни с { и закончи }.
Структура:
{
  "is_correct": true или false,
  "short_feedback": "одно предложение на русском: что сделал верно/неверно",
  "reason": "краткое объяснение на русском"
}
Правила:
• Весь текст на русском
• is_correct — boolean
• Будь строг: только математически верные ответы
• Нормализуй: запятая и точка как десятичный разделитель, игнорируй ведущие нули`

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
		return tasks.CheckResult{}, fmt.Errorf("CheckAnswer: marshal error: %w", err)
	}

	raw, err := c.Chat(systemPrompt, string(inputBytes), 300, 0.2)
	if err != nil {
		return tasks.CheckResult{}, fmt.Errorf("CheckAnswer: API error: %w", err)
	}
	
	raw = sanitizeAIResponse(raw)
	
	var result tasks.CheckResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return tasks.CheckResult{}, fmt.Errorf("CheckAnswer: JSON parse error: %w", err)
	}
	
	return result, nil
}

// === ПОДСКАЗКА ===
func (c *Client) Hint(ctx context.Context, task tasks.Task) (tasks.HintResult, error) {
	systemPrompt := `Ты — поддерживающий репетитор для ученика 9 класса.
Ученик ошибся. Дай мягкую подсказку, не раскрывая ответ.
Верни ТОЛЬКО JSON:
{
  "hint": "наводящий вопрос на русском, который направляет к решению"
}
Правила:
• Весь текст на русском
• Подсказка — вопрос, не утверждение
• НЕ раскрывай ответ или полный метод
• Будь конкретен к этой задаче
• Максимум 2 предложения`

	type hintInput struct {
		Question string `json:"question"`
	}
	
	inputBytes, err := json.Marshal(hintInput{Question: task.Question})
	if err != nil {
		return tasks.HintResult{}, fmt.Errorf("Hint: marshal error: %w", err)
	}

	raw, err := c.Chat(systemPrompt, string(inputBytes), 300, 0.6)
	if err != nil {
		return tasks.HintResult{}, fmt.Errorf("Hint: API error: %w", err)
	}
	
	raw = sanitizeAIResponse(raw)
	
	var parsed struct {
		Hint string `json:"hint"`
	}
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return tasks.HintResult{}, fmt.Errorf("Hint: JSON parse error: %w", err)
	}
	
	return tasks.HintResult{Hint: strings.TrimSpace(parsed.Hint)}, nil
}

// === ОБЪЯСНЕНИЕ ===
func (c *Client) Explain(ctx context.Context, task tasks.Task) (tasks.ExplainResult, error) {
	systemPrompt := `Ты — эксперт-учитель, объясняющий полное решение задачи ОГЭ.
Ученик не справился. Дай пошаговое объяснение.
Верни ТОЛЬКО JSON:
{
  "explanation": "полное объяснение на русском: дано, требуется, метод",
  "steps": ["шаг 1 с вычислением", "шаг 2", "шаг 3"]
}
Правила:
• Весь текст на русском
• steps: от 3 до 7 шагов
• Каждый шаг должен показывать вычисление явно
• Простой язык для 9 класса`

	type explainInput struct {
		Question      string `json:"question"`
		CorrectAnswer string `json:"correct_answer"`
	}
	
	inputBytes, err := json.Marshal(explainInput{
		Question:      task.Question,
		CorrectAnswer: task.CorrectAnswer,
	})
	if err != nil {
		return tasks.ExplainResult{}, fmt.Errorf("Explain: marshal error: %w", err)
	}

	raw, err := c.Chat(systemPrompt, string(inputBytes), 600, 0.4)
	if err != nil {
		return tasks.ExplainResult{}, fmt.Errorf("Explain: API error: %w", err)
	}
	
	raw = sanitizeAIResponse(raw)
	
	var result tasks.ExplainResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return tasks.ExplainResult{}, fmt.Errorf("Explain: JSON parse error: %w", err)
	}
	
	return result, nil
}

// === АНАЛИЗ ДИАГНОСТИКИ ===
func (c *Client) AnalyzeDiagnostic(ctx context.Context, answers []diagnostic.AnswerResult) (diagnostic.Analysis, error) {
	select {
	case <-ctx.Done():
		return diagnostic.Analysis{}, ctx.Err()
	default:
	}

	systemPrompt := `Ты анализируешь диагностику ученика по ОГЭ. Верни только валидный JSON без markdown.`
	
	payload, err := json.Marshal(answers)
	if err != nil {
		return diagnostic.Analysis{}, fmt.Errorf("AnalyzeDiagnostic: marshal error: %w", err)
	}

	prompt := fmt.Sprintf(`Проанализируй результаты диагностики по заданиям ОГЭ №6-19.
Ответы ученика: %s
Верни строго JSON:
{
  "summary": "краткий вывод на русском (1-2 предложения)",
  "weak_topics": ["тема 1", "тема 2"] // максимум 3 темы, конкретные названия
}`, string(payload))

	raw, err := c.Chat(systemPrompt, prompt, 600, 0.2)
	if err != nil {
		return diagnostic.Analysis{}, fmt.Errorf("AnalyzeDiagnostic: API error: %w", err)
	}
	
	raw = sanitizeAIResponse(raw)
	
	var result diagnostic.Analysis
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return diagnostic.Analysis{}, fmt.Errorf("AnalyzeDiagnostic: JSON parse error: %w", err)
	}
	
	result.Summary = strings.TrimSpace(result.Summary)
	for i := range result.WeakTopics {
		result.WeakTopics[i] = strings.TrimSpace(result.WeakTopics[i])
	}
	
	// Валидация: не более 3 тем
	if len(result.WeakTopics) > 3 {
		result.WeakTopics = result.WeakTopics[:3]
	}
	
	return result, nil
}
