package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/oge-math-trainer/oge-backend.git/internal/tasks"
)

// sanitizeAIResponse очищает ответ AI от типичных LaTeX артефактов
func sanitizeAIResponse(raw string) string {
	raw = strings.TrimSpace(raw)

	// Убираем markdown
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")

	// Извлекаем JSON
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start >= 0 && end > start {
		raw = raw[start : end+1]
	}

	// === ОСНОВНАЯ ЧИСТКА ===

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

	// 3. Убираем \begin и \end
	raw = strings.ReplaceAll(raw, `\begin{cases}`, "")
	raw = strings.ReplaceAll(raw, `\end{cases}`, "")
	raw = strings.ReplaceAll(raw, `\\begin{cases}`, "")
	raw = strings.ReplaceAll(raw, `\\end{cases}`, "")
	raw = strings.ReplaceAll(raw, `\begin`, "")
	raw = strings.ReplaceAll(raw, `\end`, "")

	// 4. Заменяем \frac{a}{b} → a/b
	reFrac := regexp.MustCompile(`\\frac\{([^}]+)\}\{([^}]+)\}`)
	raw = reFrac.ReplaceAllString(raw, "$1/$2")

	reCdot := regexp.MustCompile(`\\cdot`)
	raw = reCdot.ReplaceAllString(raw, "*")

	// 5. Заменяем \sqrt{a} → sqrt(a)
	reSqrt := regexp.MustCompile(`\\sqrt\{([^}]+)\}`)
	raw = reSqrt.ReplaceAllString(raw, "sqrt($1)")

	// 6. Убираем бэкслеши перед буквами, цифрами и спецсимволами
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

// validAnswerRe проверяет что ответ содержит только цифры, запятую и минус
var validAnswerRe = regexp.MustCompile(`^-?[0-9]+(,[0-9]+)?$`)

func validateAnswer(answer string) error {
	answer = strings.TrimSpace(answer)
	if answer == "" {
		return fmt.Errorf("correct_answer пустой")
	}
	if !validAnswerRe.MatchString(answer) {
		return fmt.Errorf("correct_answer содержит недопустимые символы: %q", answer)
	}
	return nil
}

// IsConfigured проверяет что клиент настроен (ключ задан)
func (c *Client) IsConfigured() bool {
	return c.apiKey != ""
}

// GenerateTask генерирует задание ОГЭ по математике
func (c *Client) GenerateTask(ctx context.Context, target tasks.Target) (tasks.GeneratedContent, error) {

	systemPrompt := `You are an expert mathematics teacher for Russian OGE exam.

Generate ONE math problem. 

Respond ONLY with valid JSON, nothing else. No explanations, no markdown, no code blocks.

{
  "question": "Текст задачи на русском языке",
  "correct_answer": "только цифры, минус и запятая",
  "solution_steps": ["шаг 1", "шаг 2", "шаг 3"],
  "self_check": "как проверить ответ",
  "is_valid": true,
  "validation_notes": ""
}

СТРОГО ЗАПРЕЩЕНО:
- Использовать любой LaTeX: \( \), \[ \], \begin, \end, \\(
- Использовать двойные бэкслеши \\
- Использовать \frac, \sqrt, \begin{cases} и любые другие LaTeX команды
- Использовать символы $ для математики

Для записи формул используй ТОЛЬКО простой текст:
Правильно: "x^2 + 3x - 5 = 0" или "2x + y = 5"
Неправильно: anything with \( or \begin

All text must be in Russian.
correct_answer — только цифры, запятая и минус, без пробелов.
Do not add any extra text outside JSON.

    // ... остальной код без изменений

For oge_number 11:
correct_answer may contain graph matching notation like:
"1-A,2-B"
or inequalities like:
"a>0,b<0"

  if oge_number is 11, you MUST add a "graphs" field to the JSON.

"graphs" is an array of 1 to 4 objects depending on the task type.

Each graph object must contain:

{
	"id": "1",
	"type": "linear" | "quadratic" | "hyperbola",
	"coefficients": [numbers],
	"x_min": -10,
	"x_max": 10,
	"y_min": -10,
	"y_max": 10
}

Rules for graph types:
* linear:
  y = ax + b
  coefficients format: [a, b]
* quadratic:
  y = ax^2 + bx + c
  coefficients format: [a, b, c]
* hyperbola:
  y = k / x
  coefficients format: [k]
Rules:
- coefficients must contain only numbers
- all ranges must be integers
- do NOT include formula strings
- do NOT include LaTeX
- graphs must be mathematically correct

If the task requires matching graphs to formulas, "correct_answer" must be like "1-A,2-B,3-C".

For all other oge_number values, do NOT include the "graphs" field.

IMPORTANT JSON RULES:
CRITICAL RULE FOR FRACTIONS:
- Use only plain text: 1/2, 3/4, 5/2, a/b
- NEVER use \frac{}, \\frac{}, or any LaTeX commands
- Do not use backslashes at all in mathematical expressions
IMPORTANT:
- Do NOT use any LaTeX at all.
- Do NOT use backslashes before ( ) [ ] or letters.
- Output clean JSON only.
NEVER use LaTeX delimiters: \( \), \[ \], $$ $$
NEVER use backslashes in mathematical expressions
NEVER escape parentheses
Use only plain UTF-8 text
Example correct: "x^2 + 2x - 3 = 0"
Example wrong: "(x^2 + 2x - 3 = 0)"
FINAL STRICT RULE:
You MUST NOT use any of these: \(  \)  \[  \]  \\(  \\)  \\[  \\]  \begin  \end
All math expressions must be written in plain text only, for example: 
"x + y = 5" or "2x - y = 1" or "x = 2"
Never wrap math in any LaTeX delimiters.
`

	type generateInput struct {
		OgeNumber   int    `json:"oge_number"`
		SubtypeCode string `json:"subtype_code"`
	}
	inputBytes, err := json.Marshal(generateInput{
		OgeNumber:   target.OgeNumber,
		SubtypeCode: target.SubtypeCode,
	})
	if err != nil {
		return tasks.GeneratedContent{}, fmt.Errorf("GenerateTask: %w", err)
	}

	raw, err := c.Chat(systemPrompt, string(inputBytes), 800, 0.7)
	fmt.Println("RAW:", raw)
	if err != nil {
		return tasks.GeneratedContent{}, fmt.Errorf("GenerateTask: %w", err)
	}

	fmt.Println("========== RAW BEFORE SANITIZE ==========")
	fmt.Println(raw)

	raw = sanitizeAIResponse(raw)

	fmt.Println("========== RAW AFTER SANITIZE ==========")
	fmt.Println(raw)

	var result tasks.GeneratedContent
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return tasks.GeneratedContent{}, fmt.Errorf("GenerateTask: не удалось распарсить ответ AI: %w", err)
	}

	// валидация correct_answer
	if err := validateAnswer(result.CorrectAnswer); err != nil {
		return tasks.GeneratedContent{}, fmt.Errorf("GenerateTask: %w", err)
	}

	return result, nil
}

// CheckAnswer проверяет ответ ученика
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

	// studentAnswer передаётся в промпт
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

	// cleanLatex вызывается ДО json.Unmarshal
	raw = sanitizeAIResponse(raw)

	var result tasks.CheckResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return tasks.CheckResult{}, fmt.Errorf("CheckAnswer: не удалось распарсить ответ AI: %w", err)
	}

	return result, nil
}

// Hint возвращает подсказку первого уровня (наводящий вопрос)
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
	inputBytes, err := json.Marshal(hintInput{
		Question: task.Question,
	})
	if err != nil {
		return tasks.HintResult{}, fmt.Errorf("Hint: %w", err)
	}

	raw, err := c.Chat(systemPrompt, string(inputBytes), 300, 0.6)
	if err != nil {
		return tasks.HintResult{}, fmt.Errorf("Hint: %w", err)
	}

	// cleanLatex вызывается ДО json.Unmarshal
	raw = sanitizeAIResponse(raw)

	var parsed struct {
		Hint string `json:"hint"`
	}
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return tasks.HintResult{}, fmt.Errorf("Hint: не удалось распарсить ответ AI: %w", err)
	}

	return tasks.HintResult{Hint: parsed.Hint}, nil
}

// Explain возвращает полный пошаговый разбор задачи
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

	// cleanLatex вызывается ДО json.Unmarshal
	raw = sanitizeAIResponse(raw)

	var result tasks.ExplainResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return tasks.ExplainResult{}, fmt.Errorf("Explain: не удалось распарсить ответ AI: %w", err)
	}

	return result, nil
}
