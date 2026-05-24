package ai

import (
	"strings"
	"testing"

	"github.com/oge-math-trainer/oge-backend.git/internal/tasks"
)

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

func TestBuildGeneratePromptUsesSeparateGraphPlots(t *testing.T) {
	prompt := buildGeneratePrompt(tasks.Target{OgeNumber: 11, SubtypeCode: "graphs_match"}, "")

	if !strings.Contains(prompt, `"plots"`) {
		t.Fatalf("expected graph prompt to request separate plots, got prompt:\n%s", prompt)
	}
	if !strings.Contains(prompt, "separate graph pictures") {
		t.Fatalf("expected graph prompt to forbid combining curves into one picture")
	}
}

func TestBuildGeneratePromptForbidsDrawingLatexInText(t *testing.T) {
	prompt := buildGeneratePrompt(tasks.Target{OgeNumber: 7, SubtypeCode: "numberline_compare"}, "")

	for _, token := range []string{"TikZ", "PGFPlots", `\draw`, "visual_data"} {
		if !strings.Contains(prompt, token) {
			t.Fatalf("expected prompt to mention %q", token)
		}
	}
}

func TestValidateAnswerFormats(t *testing.T) {
	valid := []string{"12", "-3", "4,5", "231"}
	for _, answer := range valid {
		if err := validateAnswer(answer); err != nil {
			t.Fatalf("expected %q to be valid: %v", answer, err)
		}
	}

	invalid := []string{"", "1-A,2-B", "2/3", "4.5"}
	for _, answer := range invalid {
		if err := validateAnswer(answer); err == nil {
			t.Fatalf("expected %q to be invalid", answer)
		}
	}
}

func TestValidateGeneratedTaskVisualRequirements(t *testing.T) {
	tests := []struct {
		name    string
		target  tasks.Target
		content tasks.GeneratedContent
		wantErr bool
	}{
		{
			name:   "number line required and accepted",
			target: tasks.Target{OgeNumber: 7, SubtypeCode: "numberline_compare"},
			content: generatedContentWithVisual(tasks.VisualData{
				"type":     string(tasks.VisualKindNumberLine),
				"axis":     map[string]any{"min": -5, "max": 5},
				"interval": map[string]any{"start": -1, "end": 3, "start_closed": true, "end_closed": false},
			}),
		},
		{
			name:   "graph required and accepted",
			target: tasks.Target{OgeNumber: 11, SubtypeCode: "graphs_match"},
			content: generatedContentWithVisual(tasks.VisualData{
				"type":   string(tasks.VisualKindGraph),
				"x_axis": map[string]any{"min": -5, "max": 5},
				"y_axis": map[string]any{"min": -5, "max": 5},
				"graphs": []any{
					map[string]any{"points": []any{map[string]any{"x": -1, "y": -1}, map[string]any{"x": 0, "y": 0}, map[string]any{"x": 1, "y": 1}}},
				},
			}),
		},
		{
			name:    "geometry required",
			target:  tasks.Target{OgeNumber: 15, SubtypeCode: "triangles_area"},
			content: generatedContentWithVisual(nil),
			wantErr: true,
		},
		{
			name:   "grid required and accepted",
			target: tasks.Target{OgeNumber: 18, SubtypeCode: "grid_distance"},
			content: generatedContentWithVisual(tasks.VisualData{
				"type":   string(tasks.VisualKindGrid),
				"width":  8,
				"height": 6,
				"points": []any{
					map[string]any{"label": "A", "x": 1, "y": 1},
					map[string]any{"label": "B", "x": 4, "y": 5},
				},
				"segments": []any{map[string]any{"from": "A", "to": "B"}},
			}),
		},
		{
			name:   "plain task can omit visual",
			target: tasks.Target{OgeNumber: 6, SubtypeCode: "numbers_decimal"},
			content: func() tasks.GeneratedContent {
				c := generatedContentWithVisual(nil)
				c.Question = "Найдите значение выражения 3,5 + 2,5."
				return c
			}(),
		},
		{
			name:   "oge 19 can omit visual",
			target: tasks.Target{OgeNumber: 19, SubtypeCode: "logic_angles"},
			content: func() tasks.GeneratedContent {
				c := generatedContentWithVisual(nil)
				c.Question = "Укажите номера верных утверждений. 1) Вертикальные углы равны. 2) Любой квадрат не является прямоугольником. 3) Смежные углы в сумме дают 180 градусов."
				c.CorrectAnswer = "13"
				return c
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateGeneratedTask(tt.target, &tt.content)
			if tt.wantErr && err == nil {
				t.Fatal("expected validation error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected valid task: %v", err)
			}
		})
	}
}

func generatedContentWithVisual(visual tasks.VisualData) tasks.GeneratedContent {
	return tasks.GeneratedContent{
		Question:        "Сгенерированное задание.",
		CorrectAnswer:   "1",
		SolutionSteps:   []string{"step 1", "step 2", "step 3"},
		SelfCheck:       "check",
		IsValid:         true,
		ValidationNotes: "ok",
		VisualData:      visual,
	}
}
