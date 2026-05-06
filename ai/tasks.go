package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/oge-math-trainer/oge-backend.git/internal/tasks"
)

// cleanLatex убирает LaTeX-разметку которую добавляют некоторые модели
func cleanLatex(s string) string {
	s = strings.ReplaceAll(s, `\(`, "")
	s = strings.ReplaceAll(s, `\)`, "")
	s = strings.ReplaceAll(s, `\[`, "")
	s = strings.ReplaceAll(s, `\]`, "")
	s = strings.ReplaceAll(s, `\cdot`, "x")
	s = strings.ReplaceAll(s, "→", "=>")
	s = strings.ReplaceAll(s, "≠", "!=")
	s = strings.ReplaceAll(s, "≤", "<=")
	s = strings.ReplaceAll(s, "≥", ">=")
	s = strings.ReplaceAll(s, "±", "+/-")
	return s
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
	systemPrompt := `You are an expert mathematics teacher specializing in Russian OGE exam preparation.
Generate one math problem for the OGE exam.
Respond ONLY with valid JSON. No markdown, no code blocks, no text before or after. Start with { and end with }.
Return exactly this structure:
{
  "question": "full problem statement in Russian",
  "correct_answer": "digits/comma/minus only, no spaces, no letters",
  "solution_steps": ["step 1", "step 2", "step 3"],
  "self_check": "how to verify the answer",
  "is_valid": true,
  "validation_notes": ""
}
Rules:
- All text in Russian
- correct_answer must contain only digits (0-9), minus (-), comma (,). No spaces, no letters, no dots.
- solution_steps minimum 3 steps, maximum 7
- Problem must match real OGE difficulty for 9th grade
- Do NOT use LaTeX, arrows (->), or any special math symbols. Use plain text only.`

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

	// cleanLatex вызывается ДО json.Unmarshal
	raw = cleanLatex(raw)

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
	raw = cleanLatex(raw)

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
	raw = cleanLatex(raw)

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
	raw = cleanLatex(raw)

	var result tasks.ExplainResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return tasks.ExplainResult{}, fmt.Errorf("Explain: не удалось распарсить ответ AI: %w", err)
	}

	return result, nil
}
