package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/oge-math-trainer/oge-backend.git/internal/app"
	"github.com/oge-math-trainer/oge-backend.git/internal/diagnostic"
	"github.com/oge-math-trainer/oge-backend.git/internal/tasks"
)

const (
	defaultAIRequestTimeout  = 30 * time.Second
	extendedAIRequestTimeout = 90 * time.Second
)

type Client struct {
	baseURL    string
	apiKey     string
	model      string
	httpClient *http.Client
}

func NewClient(baseURL, apiKey, model string) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     strings.TrimSpace(apiKey),
		model:      strings.TrimSpace(model),
		httpClient: &http.Client{},
	}
}

func (c *Client) IsConfigured() bool {
	return c != nil && strings.TrimSpace(c.apiKey) != "" && strings.TrimSpace(c.model) != "" && strings.TrimSpace(c.baseURL) != ""
}

func (c *Client) GenerateTask(ctx context.Context, target tasks.Target) (tasks.GeneratedContent, error) {
	var out tasks.GeneratedContent

	timeout := defaultAIRequestTimeout
	if tasks.RequiresExtendedAITimeout(target) {
		timeout = extendedAIRequestTimeout
	}
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, timeout)
	defer cancel()

	// 🔧 Расширенный промпт с поддержкой графиков для №11
	graphsRule := ""
	if target.OgeNumber == 11 {
		graphsRule = `
6. СПЕЦИАЛЬНО ДЛЯ oge_number=11: добавь поле "graphs" — массив из 1-4 объектов.
   Каждый объект: {"id":"1","type":"linear|quadratic|hyperbola","coefficients":[числа],"x_min":-10,"x_max":10,"y_min":-10,"y_max":10}
   - linear: y=ax+b → coefficients=[a,b]
   - quadratic: y=ax²+bx+c → coefficients=[a,b,c]  
   - hyperbola: y=k/x → coefficients=[k]
   - Все коэффициенты — только числа, без формул и LaTeX.`
	}

	prompt := fmt.Sprintf(`Ты — составитель заданий ОГЭ по математике (6-19).
Параметры: oge_number=%d, subtype_code=%s.
Требования:
1. Условие текстовое, без рисунков и таблиц, уровень base/middle.
2. Ответ — однозначное число или десятичная дробь (только цифры, запятая, минус).
3. Реши задачу пошагово (минимум 3 шага).
4. Самопроверь ответ.
5. Верни ТОЛЬКО валидный JSON без markdown, без пояснений, начинай с { и заканчивай }:
{"question":"текст","correct_answer":"число","solution_steps":["шаг 1"],"self_check":"проверка","is_valid":true,"validation_notes":"соответствует ОГЭ"}%s`,
		target.OgeNumber, target.SubtypeCode, graphsRule)
	prompt += visualDataPromptRule(target)

	if err := c.completeJSON(ctx, prompt, &out); err != nil {
		return tasks.GeneratedContent{}, err
	}

	// 🔧 Очищаем текст от знаков доллара (LaTeX) перед сохранением/возвратом
	// 🔍 Логируем для отладки
	if strings.Contains(out.Question, "$") {
		log.Printf("CLEANING question: %q -> %q", out.Question, cleanLaTeX(out.Question))
	}
	out.Question = cleanLaTeX(out.Question)

	for i, step := range out.SolutionSteps {
		if strings.Contains(step, "$") {
			log.Printf("CLEANING step[%d]: %q -> %q", i, step, cleanLaTeX(step))
		}
		out.SolutionSteps[i] = cleanLaTeX(step)
	}

	return out, nil
}

func visualDataPromptRule(target tasks.Target) string {
	switch tasks.VisualKindForTarget(target) {
	case tasks.VisualKindGraph:
		return `

VISUAL DATA RULE:
Add "visual_data" as a JSON object. Do not wrap it in a string.
Required schema:
"visual_data": {
  "type": "graph",
  "x_axis": {"min": -10, "max": 10},
  "y_axis": {"min": -10, "max": 10},
  "graphs": [
    {"id": "1", "label": "y = x", "points": [{"x": -2, "y": -2}, {"x": 0, "y": 0}, {"x": 2, "y": 2}]}
  ]
}
Each graph must contain at least 3 coordinate points. All coordinates and axis bounds must be numbers.`
	case tasks.VisualKindNumberLine:
		return `

VISUAL DATA RULE:
Add "visual_data" as a JSON object. Do not wrap it in a string.
Required schema:
"visual_data": {
  "type": "number_line",
  "axis": {"min": -10, "max": 10},
  "interval": {"start": -2, "end": 3, "start_closed": true, "end_closed": false},
  "points": [{"value": -2, "closed": true, "label": "-2"}, {"value": 3, "closed": false, "label": "3"}]
}
The interval start must be strictly less than end. The *_closed and point closed fields must be booleans.`
	case tasks.VisualKindGeometry:
		return `

VISUAL DATA RULE:
Add "visual_data" as a JSON object. Do not wrap it in a string.
Required schema:
"visual_data": {
  "type": "geometry",
  "shape": "triangle",
  "vertices": [{"label": "A", "x": 0, "y": 0}, {"label": "B", "x": 6, "y": 0}, {"label": "C", "x": 0, "y": 8}],
  "labels": {"A": "A", "B": "B", "C": "C"},
  "segments": [{"from": "A", "to": "B"}, {"from": "A", "to": "C"}, {"from": "B", "to": "C"}]
}
There must be at least 3 vertices. Every vertex must use exactly label, x, and y for its required keys. Labels must match vertices.`
	default:
		return `

VISUAL DATA RULE:
For this task type, omit "visual_data".`
	}
}

func (c *Client) CheckAnswer(ctx context.Context, task tasks.Task, studentAnswer string) (tasks.CheckResult, error) {
	var out tasks.CheckResult
	prompt := fmt.Sprintf(`Задача: %q
Правильный ответ: %q
Ответ ученика: %q
Проверь, можно ли засчитать ответ ученика. Учитывай точку/запятую, пробелы и математическую эквивалентность.
Верни только JSON:
{"is_correct":true,"short_feedback":"краткий комментарий","reason":"почему так"}`,
		task.Question, task.CorrectAnswer, studentAnswer)
	if err := c.completeJSON(ctx, prompt, &out); err != nil {
		return tasks.CheckResult{}, err
	}

	// Очищаем feedback от LaTeX
	out.ShortFeedback = sanitizeAIResponse(out.ShortFeedback)
	out.Reason = sanitizeAIResponse(out.Reason)

	return out, nil
}

func (c *Client) Hint(ctx context.Context, task tasks.Task) (tasks.HintResult, error) {
	var out tasks.HintResult
	prompt := fmt.Sprintf(`Дай одну полезную подсказку к задаче без готового ответа.
Задача: %q
Правильный ответ не раскрывай.
Верни только JSON: {"hint":"подсказка"}`, task.Question)
	if err := c.completeJSON(ctx, prompt, &out); err != nil {
		return tasks.HintResult{}, err
	}
	return out, nil
}

func (c *Client) Explain(ctx context.Context, task tasks.Task) (tasks.ExplainResult, error) {
	var out tasks.ExplainResult
	prompt := fmt.Sprintf(`Объясни решение задачи простым языком.
Задача: %q
Правильный ответ: %q
Верни только JSON:
{"explanation":"объяснение","steps":["шаг 1","шаг 2"]}`, task.Question, task.CorrectAnswer)
	if err := c.completeJSON(ctx, prompt, &out); err != nil {
		return tasks.ExplainResult{}, err
	}
	return out, nil
}

func (c *Client) AnalyzeDiagnostic(ctx context.Context, answers []diagnostic.AnswerResult) (diagnostic.Analysis, error) {
	var out diagnostic.Analysis
	payload, _ := json.Marshal(answers)
	prompt := fmt.Sprintf(`Проанализируй результаты диагностики ученика по ОГЭ 6-19.
Ответы: %s
Верни СТРОГО JSON без markdown и пояснений. Начинай сразу с { и заканчивай }:
{"summary":"краткий анализ","weak_topics":["тема 1","тема 2"]}`, string(payload))
	if err := c.completeJSON(ctx, prompt, &out); err != nil {
		return diagnostic.Analysis{}, err
	}

	// Очищаем summary и weak_topics от LaTeX
	out.Summary = sanitizeAIResponse(out.Summary)
	for i := range out.WeakTopics {
		out.WeakTopics[i] = sanitizeAIResponse(out.WeakTopics[i])
	}

	return out, nil
}

func (c *Client) completeJSON(ctx context.Context, prompt string, out any) error {
	if !c.IsConfigured() {
		return app.AIUnavailable(errors.New("AITUNNEL_API_KEY is empty"))
	}
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, defaultAIRequestTimeout)
		defer cancel()
	}

	body := chatRequest{
		Model: c.model,
		Messages: []chatMessage{
			{Role: "system", Content: "Ты возвращаешь ТОЛЬКО валидный JSON. Никакого markdown, никаких пояснений вне JSON. Начинай сразу с { и заканчивай }."},
			{Role: "user", Content: prompt},
		},
		Temperature: 0.3,
	}
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return app.Internal(err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(bodyJSON))
	if err != nil {
		return app.Internal(err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return app.AIUnavailable(err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return app.AIUnavailable(err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return app.AIUnavailable(fmt.Errorf("ai service status %d: %s", resp.StatusCode, string(respBody)))
	}

	var chatResp chatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return app.AIUnavailable(err)
	}
	if len(chatResp.Choices) == 0 {
		return app.AIUnavailable(errors.New("ai service returned no choices"))
	}

	rawContent := strings.TrimSpace(chatResp.Choices[0].Message.Content)
	debugLogf("ai raw response: %s", rawContent)
	content := rawContent

	// Чистим markdown блоки
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	// Если AI вернул текст с лишним, ищем JSON между { и }
	if !strings.HasPrefix(content, "{") {
		start := strings.Index(content, "{")
		end := strings.LastIndex(content, "}")
		if start != -1 && end != -1 && end > start {
			content = content[start : end+1]
		}
	}

	// 🧹 Очищаем LaTeX ДО парсинга JSON
	content = sanitizeAIResponse(content)

	// 🔍 Логируем сырой ответ перед парсингом
	debugLogf("ai cleaned response: %s", content)

	if err := json.Unmarshal([]byte(content), out); err != nil {
		// 🔍 Логируем конкретную ошибку парсинга
		log.Printf("ai json parse failed: err=%v raw_prefix=%.200s", err, content)
		return app.AIUnavailable(fmt.Errorf("invalid ai json: %w", err))
	}

	// 🔍 Дополнительная валидация после парсинга
	if gen, ok := out.(*tasks.GeneratedContent); ok {
		if strings.TrimSpace(gen.Question) == "" {
			log.Printf("ai validation failed: missing question")
			return app.Validation("AI returned empty question")
		}
		if strings.TrimSpace(gen.CorrectAnswer) == "" {
			log.Printf("ai validation failed: missing correct_answer")
			return app.Validation("AI returned empty correct_answer")
		}
		if len(gen.SolutionSteps) < 3 {
			log.Printf("ai validation failed: solution_steps=%d need>=3", len(gen.SolutionSteps))
			return app.Validation("AI returned too few solution steps")
		}
	}

	return nil
}

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
}

func debugLogf(format string, args ...any) {
	log.Printf("DEBUG "+format, args...)
}

// sanitizeAIResponse очищает ответ AI от типичных LaTeX артефактов
// Это предотвращает ошибку "ambiguous parameter type" в PostgreSQL
func sanitizeAIResponse(raw string) string {
	raw = strings.TrimSpace(raw)

	// Убираем markdown блоки
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")

	// Извлекаем JSON между { и }
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start >= 0 && end > start {
		raw = raw[start : end+1]
	}

	// === ОСНОВНАЯ ЧИСТКА LaTeX ===

	// 1. Заменяем `\ \` (бэкслеш-пробел-бэкслеш-пробел) на пусто
	raw = strings.ReplaceAll(raw, `\ \ `, "")
	raw = strings.ReplaceAll(raw, `\ \`, "")

	// 2. Убираем LaTeX-обрамление
	raw = strings.ReplaceAll(raw, `\(`, "")
	raw = strings.ReplaceAll(raw, `\)`, "")
	raw = strings.ReplaceAll(raw, `\[`, "")
	raw = strings.ReplaceAll(raw, `\]`, "")
	raw = strings.ReplaceAll(raw, `\\(`, "")
	raw = strings.ReplaceAll(raw, `\\)`, "")
	raw = strings.ReplaceAll(raw, `\\[`, "")
	raw = strings.ReplaceAll(raw, `\\]`, "")

	// 3. Убираем \begin и \end блоки
	raw = strings.ReplaceAll(raw, `\begin{cases}`, "")
	raw = strings.ReplaceAll(raw, `\end{cases}`, "")
	raw = strings.ReplaceAll(raw, `\\begin{cases}`, "")
	raw = strings.ReplaceAll(raw, `\\end{cases}`, "")
	raw = strings.ReplaceAll(raw, `\begin`, "")
	raw = strings.ReplaceAll(raw, `\end`, "")

	// 4. Заменяем \frac{a}{b} → a/b
	reFrac := regexp.MustCompile(`\\frac\{([^}]+)\}\{([^}]+)\}`)
	raw = reFrac.ReplaceAllString(raw, "$1/$2")

	// 5. Заменяем \sqrt{a} → sqrt(a)
	reSqrt := regexp.MustCompile(`\\sqrt\{([^}]+)\}`)
	raw = reSqrt.ReplaceAllString(raw, "sqrt($1)")

	// 6. Убираем бэкслеши перед буквами, цифрами и спецсимволами
	// Это обрабатывает `\14/5` -> `14/5`, `\x` -> `x`, и т.д.
	raw = regexp.MustCompile(`\\{1,2}([a-zA-Z0-9{}()^/_\-+*=])`).ReplaceAllString(raw, "$1")

	// 7. Убираем оставшиеся двойные бэкслеши и одиночные
	raw = strings.ReplaceAll(raw, `\\`, "")
	raw = strings.ReplaceAll(raw, `\`, "")

	// 8. Убираем знаки доллара
	raw = strings.ReplaceAll(raw, "$", "")

	// 9. Очищаем множественные пробелы (оставляем один)
	raw = regexp.MustCompile(`\s+`).ReplaceAllString(raw, " ")

	return strings.TrimSpace(raw)
}

// cleanLaTeX удаляет ВСЕ знаки доллара из текста (для обратной совместимости)
func cleanLaTeX(text string) string {
	return strings.ReplaceAll(text, "$", "")
}
