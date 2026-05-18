package ai

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// --- Структуры запроса ---

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens"`
	Temperature float64   `json:"temperature"`
}

// --- Структуры ответа ---

type ChatChoice struct {
	Message Message `json:"message"`
}

type ChatResponse struct {
	Choices []ChatChoice `json:"choices"`
}

// --- Клиент ---

type Client struct {
	apiKey     string
	baseURL    string
	model      string
	httpClient *http.Client
}

const defaultModel = "qwen/qwen3-235b-a22b-2507"

// NewClient создаёт нового AI-клиента из переменных окружения
func NewClient() (*Client, error) {
	apiKey := os.Getenv("AITUNNEL_API_KEY")
	if apiKey == "" {
		return nil, errors.New("AITUNNEL_API_KEY не задан в .env")
	}

	baseURL := os.Getenv("AITUNNEL_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.aitunnel.ru/v1"
	}

	return &Client{
		apiKey:  apiKey,
		baseURL: baseURL,
		model:   envOrDefault("AITUNNEL_MODEL", defaultModel),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

func NewConfiguredClient(baseURL, apiKey, model string) *Client {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		baseURL = "https://api.aitunnel.ru/v1"
	}
	model = strings.TrimSpace(model)
	if model == "" {
		model = defaultModel
	}

	return &Client{
		apiKey:  strings.TrimSpace(apiKey),
		baseURL: baseURL,
		model:   model,
		httpClient: &http.Client{
			Timeout: 600 * time.Second,
		},
	}
}

// Chat отправляет запрос к AI и возвращает текст ответа
func (c *Client) Chat(systemPrompt string, userMessage string, maxTokens int, temperature float64) (string, error) {
	// Собираем тело запроса
	reqBody := ChatRequest{
		Model: c.model,
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userMessage},
		},
		MaxTokens:   maxTokens,
		Temperature: temperature,
	}

	// Сериализуем в JSON
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("ошибка сериализации запроса: %w", err)
	}

	// Создаём HTTP запрос
	req, err := http.NewRequest("POST", c.baseURL+"/chat/completions", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("ошибка создания запроса: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	// Отправляем с ретраями
	var resp *http.Response
	for attempt := 1; attempt <= 3; attempt++ {
		resp, err = c.httpClient.Do(req)
		if err != nil {
			if attempt == 3 {
				return "", fmt.Errorf("ошибка запроса после 3 попыток: %w", err)
			}
			time.Sleep(time.Duration(attempt) * time.Second)
			continue
		}
		break
	}
	defer resp.Body.Close()

	// Читаем тело ответа
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	// Проверяем статус
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("AI вернул статус %d: %s", resp.StatusCode, string(respBytes))
	}

	// Парсим ответ
	var chatResp ChatResponse
	if err := json.Unmarshal(respBytes, &chatResp); err != nil {
		return "", fmt.Errorf("ошибка парсинга ответа: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", errors.New("AI вернул пустой список ответов")
	}

	// Очищаем markdown-обёртку если модель её добавила
	content := chatResp.Choices[0].Message.Content
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	return content, nil
}

func envOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
