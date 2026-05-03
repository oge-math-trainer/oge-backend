package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/oge-math-trainer/oge-backend.git/internal/app"
	"github.com/oge-math-trainer/oge-backend.git/internal/diagnostic"
	"github.com/oge-math-trainer/oge-backend.git/internal/tasks"
)

type Client struct {
	baseURL    string
	apiKey     string
	model      string
	httpClient *http.Client
}

func NewClient(baseURL, apiKey, model string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  strings.TrimSpace(apiKey),
		model:   strings.TrimSpace(model),
		httpClient: &http.Client{
			Timeout: 45 * time.Second,
		},
	}
}

func (c *Client) IsConfigured() bool {
	return c != nil && strings.TrimSpace(c.apiKey) != "" && strings.TrimSpace(c.model) != "" && strings.TrimSpace(c.baseURL) != ""
}

func (c *Client) GenerateTask(ctx context.Context, target tasks.Target) (tasks.GeneratedContent, error) {
	var out tasks.GeneratedContent
	prompt := fmt.Sprintf(`Ты — составитель заданий ОГЭ по математике (6-19).
Параметры: oge_number=%d, subtype_code=%s.
Требования:
1. Условие текстовое, без рисунков и таблиц, уровень base/middle.
2. Ответ — однозначное число или десятичная дробь.
3. Реши задачу пошагово.
4. Самопроверь ответ.
5. Верни только JSON:
{"question":"текст","correct_answer":"число","solution_steps":["шаг 1"],"self_check":"проверка","is_valid":true,"validation_notes":"соответствует ОГЭ"}`,
		target.OgeNumber, target.SubtypeCode)
	if err := c.completeJSON(ctx, prompt, &out); err != nil {
		return tasks.GeneratedContent{}, err
	}
	return out, nil
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
Верни только JSON:
{"summary":"краткий анализ","weak_topics":["тема 1","тема 2"]}`, string(payload))
	if err := c.completeJSON(ctx, prompt, &out); err != nil {
		return diagnostic.Analysis{}, err
	}
	return out, nil
}

func (c *Client) completeJSON(ctx context.Context, prompt string, out any) error {
	if !c.IsConfigured() {
		return app.AIUnavailable(errors.New("OPENROUTER_API_KEY is empty"))
	}

	body := chatRequest{
		Model: c.model,
		Messages: []chatMessage{
			{Role: "system", Content: "Ты возвращаешь только валидный JSON без markdown и пояснений."},
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
	req.Header.Set("HTTP-Referer", "http://localhost")
	req.Header.Set("X-Title", "oge-ai-trainer")

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
		return app.AIUnavailable(fmt.Errorf("openrouter status %d", resp.StatusCode))
	}

	var chatResp chatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return app.AIUnavailable(err)
	}
	if len(chatResp.Choices) == 0 {
		return app.AIUnavailable(errors.New("openrouter returned no choices"))
	}

	content := strings.TrimSpace(chatResp.Choices[0].Message.Content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)
	if err := json.Unmarshal([]byte(content), out); err != nil {
		return app.AIUnavailable(err)
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
