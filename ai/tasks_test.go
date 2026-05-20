package ai

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/oge-math-trainer/oge-backend.git/internal/tasks"
)

// Вспомогательная функция: создаёт GeneratedContent с визуальными данными
func generatedContentWithVisual(visual map[string]any) tasks.GeneratedContent {
	var visualRaw json.RawMessage
	if visual != nil {
		visualRaw, _ = json.Marshal(visual)
	}
	return tasks.GeneratedContent{
		Question:        "Сгенерированное задание.",
		CorrectAnswer:   "1",
		SolutionSteps:   []string{"Шаг 1", "Шаг 2", "Шаг 3"},
		SelfCheck:       "Проверка",
		IsValid:         true,
		ValidationNotes: "ok",
		VisualData:      visualRaw,
	}
}

// Тест: промпт для касательных к окружности содержит правильный маркер
func TestBuildGeneratePromptUsesCircleTangentsTemplate(t *testing.T) {
	prompt := buildGeneratePrompt(tasks.Target{OgeNumber: 16, SubtypeCode: "circle_tangents"}, "")

	if !strings.Contains(prompt, "circle_tangents tangent_from_external_point") {
		t.Fatalf("expected circle tangents template marker, got prompt:\n%s", prompt)
	}
	if !strings.Contains(prompt, `"visual_data"`) {
		t.Fatalf("expected visual_data schema in prompt")
	}
	if strings.Contains(prompt, "triangles_pythagor") {
		t.Fatalf("prompt leaked a neighboring triangle subtype")
	}
}

// Тест: валидация ответов для разных форматов
func TestValidateAnswerFormats(t *testing.T) {
	tests := []struct {
		name      string
		target    tasks.Target
		answer    string
		wantValid bool
	}{
		// Валидные ответы
		{"целое положительное", tasks.Target{OgeNumber: 6}, "12", true},
		{"целое отрицательное", tasks.Target{OgeNumber: 9}, "-3", true},
		{"десятичное с запятой", tasks.Target{OgeNumber: 7}, "4,5", true},
		{"последовательность цифр №11", tasks.Target{OgeNumber: 11}, "231", true},
		{"последовательность цифр №19", tasks.Target{OgeNumber: 19}, "13", true},
		
		// Невалидные ответы
		{"пустой ответ", tasks.Target{OgeNumber: 6}, "", false},
		{"формат 1-A", tasks.Target{OgeNumber: 11}, "1-A,2-B", false},
		{"дробь 2/3", tasks.Target{OgeNumber: 8}, "2/3", false},
		{"точка вместо запятой", tasks.Target{OgeNumber: 7}, "4.5", false},
		{"буквы в ответе", tasks.Target{OgeNumber: 6}, "abc", false},
		{"слишком короткая последовательность для №11", tasks.Target{OgeNumber: 11}, "5", false},
		{"слишком длинная последовательность для №19", tasks.Target{OgeNumber: 19}, "12345", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAnswer(tt.target, tt.answer)
			if tt.wantValid && err != nil {
				t.Fatalf("expected %q to be valid for target %+v: %v", tt.answer, tt.target, err)
			}
			if !tt.wantValid && err == nil {
				t.Fatalf("expected %q to be invalid for target %+v", tt.answer, tt.target)
			}
		})
	}
}

// Тест: валидация визуальных данных для разных типов заданий
func TestValidateGeneratedTaskVisualRequirements(t *testing.T) {
	tests := []struct {
		name    string
		target  tasks.Target
		content tasks.GeneratedContent
		wantErr bool
	}{
		{
			name:   "number line: визуал есть и валиден",
			target: tasks.Target{OgeNumber: 7, SubtypeCode: "numberline_compare"},
			content: generatedContentWithVisual(map[string]any{
				"type":     "number_line",
				"axis":     map[string]any{"min": -5, "max": 5},
				"interval": map[string]any{"start": -1, "end": 3, "start_closed": true, "end_closed": false},
			}),
			wantErr: false,
		},
		{
			name:   "graph: визуал есть и валиден",
			target: tasks.Target{OgeNumber: 11, SubtypeCode: "graphs_match"},
			content: generatedContentWithVisual(map[string]any{
				"type":   "graph",
				"x_axis": map[string]any{"min": -5, "max": 5},
				"y_axis": map[string]any{"min": -5, "max": 5},
				"graphs": []any{
					map[string]any{"points": []any{
						map[string]any{"x": -1, "y": -1},
						map[string]any{"x": 0, "y": 0},
						map[string]any{"x": 1, "y": 1},
					}},
				},
			}),
			wantErr: false,
		},
		{
			name:    "geometry: визуал обязателен, но отсутствует",
			target:  tasks.Target{OgeNumber: 15, SubtypeCode: "triangles_area"},
			content: generatedContentWithVisual(nil),
			wantErr: true,
		},
		{
			name:   "grid: визуал есть и валиден",
			target: tasks.Target{OgeNumber: 18, SubtypeCode: "grid_distance"},
			content: generatedContentWithVisual(map[string]any{
				"type":   "grid",
				"width":  8,
				"height": 6,
				"points": []any{
					map[string]any{"label": "A", "x": 1, "y": 1},
					map[string]any{"label": "B", "x": 4, "y": 5},
				},
				"segments": []any{map[string]any{"from": "A", "to": "B"}},
			}),
			wantErr: false,
		},
		{
			name:   "plain task: визуал не требуется",
			target: tasks.Target{OgeNumber: 6, SubtypeCode: "numbers_decimal"},
			content: func() tasks.GeneratedContent {
				c := generatedContentWithVisual(nil)
				c.Question = "Найдите значение выражения 3,5 + 2,5."
				return c
			}(),
			wantErr: false,
		},
		{
			name:   "oge 19: визуал не требуется",
			target: tasks.Target{OgeNumber: 19, SubtypeCode: "logic_angles"},
			content: func() tasks.GeneratedContent {
				c := generatedContentWithVisual(nil)
				c.Question = "Укажите номера верных утверждений. 1) Вертикальные углы равны. 2) Любой квадрат не является прямоугольником. 3) Смежные углы в сумме дают 180 градусов."
				c.CorrectAnswer = "13"
				return c
			}(),
			wantErr: false,
		},
		{
			name:   "number line: start >= end — ошибка",
			target: tasks.Target{OgeNumber: 7, SubtypeCode: "numberline_compare"},
			content: generatedContentWithVisual(map[string]any{
				"type":     "number_line",
				"axis":     map[string]any{"min": -5, "max": 5},
				"interval": map[string]any{"start": 3, "end": 1, "start_closed": true, "end_closed": false},
			}),
			wantErr: true,
		},
		{
			name:   "graph: меньше 3 точек — ошибка",
			target: tasks.Target{OgeNumber: 11, SubtypeCode: "graphs_linear"},
			content: generatedContentWithVisual(map[string]any{
				"type":   "graph",
				"x_axis": map[string]any{"min": -5, "max": 5},
				"y_axis": map[string]any{"min": -5, "max": 5},
				"graphs": []any{
					map[string]any{"points": []any{
						map[string]any{"x": 0, "y": 0},
						map[string]any{"x": 1, "y": 1},
					}},
				},
			}),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateGeneratedTask(tt.target, &tt.content)
			if tt.wantErr && err == nil {
				t.Fatal("expected validation error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected valid task, got error: %v", err)
			}
		})
	}
}

// === НОВЫЕ ТЕСТЫ ===

// Тест: санитизация ответа ИИ
func TestSanitizeAIResponse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "убирает markdown-блоки",
			input:    "```json\n{\"question\":\"test\"}\n```",
			expected: `{"question":"test"}`,
		},
		{
			name:     "вырезает только JSON из текста",
			input:    "Вот ответ: {\"question\":\"test\"} конец",
			expected: `{"question":"test"}`,
		},
		{
			name:     "убирает лишние слеши в математике",
			input:    `{"formula":"\sqrt{4}"}`,
			expected: `{"formula":"\sqrt{4}"}`, // один слэш остаётся
		},
		{
			name:     "убирает $ делимитеры",
			input:    `{"formula":"$x^2$"}`,
			expected: `{"formula":"x^2"}`,
		},
		{
			name:     "нормализует пробелы",
			input:    "{\"question\":\"test    with   spaces\"}",
			expected: `{"question":"test with spaces"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeAIResponse(tt.input)
			if got != tt.expected {
				t.Errorf("sanitizeAIResponse() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// Тест: промпт для разных типов заданий содержит правильные инструкции
func TestBuildGeneratePromptIncludesSubtypeInstructions(t *testing.T) {
	tests := []struct {
		name           string
		target         tasks.Target
		mustContain    []string
		mustNotContain []string
	}{
		{
			name:   "№11 graphs_quadratic: ответ — последовательность цифр",
			target: tasks.Target{OgeNumber: 11, SubtypeCode: "graphs_quadratic"},
			mustContain: []string{
				"последовательность цифр",
				"например 231",
				"не используй буквы",
			},
			mustNotContain: []string{"1-A", "формат ответа: число"},
		},
		{
			name:   "№19 logic_triangles: ответ — цифры верных утверждений",
			target: tasks.Target{OgeNumber: 19, SubtypeCode: "logic_triangles"},
			mustContain: []string{
				"ровно 3 утверждения",
				"цифры верных утверждений в возрастающем порядке",
				"например 13",
			},
			mustNotContain: []string{"вычислите", "найдите длину"},
		},
		{
			name:   "№7 numberline: ответ — число или номер варианта",
			target: tasks.Target{OgeNumber: 7, SubtypeCode: "numberline_compare"},
			mustContain: []string{
				"номер варианта или число",
				"десятичная дробь с запятой",
			},
			mustNotContain: []string{"последовательность цифр", "231"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt := buildGeneratePrompt(tt.target, "")
			for _, must := range tt.mustContain {
				if !strings.Contains(prompt, must) {
					t.Errorf("prompt must contain %q, got:\n%s", must, prompt)
				}
			}
			for _, mustNot := range tt.mustNotContain {
				if strings.Contains(prompt, mustNot) {
					t.Errorf("prompt must NOT contain %q, got:\n%s", mustNot, prompt)
				}
			}
		})
	}
}

// Тест: температура зависит от типа задания
func TestGenerateTaskUsesCorrectTemperature(t *testing.T) {
	// Этот тест требует мока AI-клиента, но можно проверить логику:
	// - Для №11 и №19 температура должна быть 0.1
	// - Для остальных — 0.15
	
	// Пока тестируем, что функция не паникует при разных target
	targets := []tasks.Target{
		{OgeNumber: 11, SubtypeCode: "graphs_match"},
		{OgeNumber: 19, SubtypeCode: "logic_angles"},
		{OgeNumber: 7, SubtypeCode: "numberline_compare"},
	}
	
	for _, target := range targets {
		prompt := buildGeneratePrompt(target, "")
		if !strings.Contains(prompt, "Верни ТОЛЬКО валидный JSON") {
			t.Errorf("prompt for target %+v missing strict JSON instruction", target)
		}
	}
}

// Тест: визуальная математическая валидация
func TestValidateVisualMath(t *testing.T) {
	tests := []struct {
		name    string
		target  tasks.Target
		visual  map[string]any
		wantErr bool
	}{
		{
			name:   "graph: 3+ точек — ок",
			target: tasks.Target{OgeNumber: 11},
			visual: map[string]any{
				"type": "graph",
				"graphs": []any{
					map[string]any{"points": []any{
						map[string]any{"x": 0, "y": 0},
						map[string]any{"x": 1, "y": 1},
						map[string]any{"x": 2, "y": 4},
					}},
				},
			},
			wantErr: false,
		},
		{
			name:   "graph: меньше 3 точек — ошибка",
			target: tasks.Target{OgeNumber: 11},
			visual: map[string]any{
				"type": "graph",
				"graphs": []any{
					map[string]any{"points": []any{
						map[string]any{"x": 0, "y": 0},
						map[string]any{"x": 1, "y": 1},
					}},
				},
			},
			wantErr: true,
		},
		{
			name:   "number_line: start < end — ок",
			target: tasks.Target{OgeNumber: 7},
			visual: map[string]any{
				"type": "number_line",
				"interval": map[string]any{"start": -2, "end": 3},
			},
			wantErr: false,
		},
		{
			name:   "number_line: start >= end — ошибка",
			target: tasks.Target{OgeNumber: 7},
			visual: map[string]any{
				"type": "number_line",
				"interval": map[string]any{"start": 3, "end": 1},
			},
			wantErr: true,
		},
		{
			name:   "geometry: 3+ вершины — ок",
			target: tasks.Target{OgeNumber: 15},
			visual: map[string]any{
				"type": "geometry",
				"vertices": []any{
					map[string]any{"label": "A", "x": 0, "y": 0},
					map[string]any{"label": "B", "x": 3, "y": 0},
					map[string]any{"label": "C", "x": 0, "y": 4},
				},
			},
			wantErr: false,
		},
		{
			name:   "geometry: меньше 3 вершин — ошибка",
			target: tasks.Target{OgeNumber: 15},
			visual: map[string]any{
				"type": "geometry",
				"vertices": []any{
					map[string]any{"label": "A", "x": 0, "y": 0},
					map[string]any{"label": "B", "x": 3, "y": 0},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := generatedContentWithVisual(tt.visual)
			err := validateVisualMath(tt.target, &content)
			if tt.wantErr && err == nil {
				t.Fatal("expected visual math validation error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected valid visual math, got error: %v", err)
			}
		})
	}
}
