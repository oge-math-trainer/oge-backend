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

	systemPrompt := `You are an expert mathematics teacher for Russian OGE exam preparation.

Generate exactly ONE OGE math problem.

Respond ONLY with valid JSON.
No markdown.
No explanations.
No code blocks.
No text before or after JSON.

JSON format:

{
  "question": "problem text in Russian",
  "correct_answer": "answer",
  "solution_steps": ["step 1", "step 2", "step 3"],
  "self_check": "how to verify the answer",
  "is_valid": true,
  "validation_notes": ""
}

Rules:
- All text must be in Russian.
- Problem must match real OGE difficulty and subtype.
- solution_steps: 3 to 7 steps.
- Use plain text math only.
- Never use LaTeX.
- Never use backslashes.
- Never use markdown formatting.
- Mathematical expressions must look like:
  "x^2 + 3x - 5 = 0"
  "2x + y = 7"
  "3/4"

Forbidden:
\(
\)
\[
\]
$$
\frac
\sqrt
\begin
\end

FOR OGE NUMBER 11:
- Add field "graphs".
- graphs must contain 1 to 4 graph objects.

Graph object format:

{
  "id": "1",
  "type": "linear",
  "coefficients": [2, 1],
  "x_min": -10,
  "x_max": 10,
  "y_min": -10,
  "y_max": 10
}

Graph rules:
- linear => [a, b] for y = ax + b
- quadratic => [a, b, c] for y = ax^2 + bx + c
- hyperbola => [k] for y = k/x
- coefficients must contain only numbers
- ranges must be integers
- do not include formula strings

For graph matching tasks:
correct_answer may be:
"1-A,2-B,3-C"

For coefficient sign tasks:
correct_answer may be:
"a>0,b<0"

For all other tasks:
Do NOT include "graphs".
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
