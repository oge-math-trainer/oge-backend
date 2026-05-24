package tasks

import (
	"strings"
	"testing"
)

func TestValidateGraphVisualData(t *testing.T) {
	data := VisualData{
		"type":   string(VisualKindGraph),
		"x_axis": map[string]any{"min": -5, "max": 5},
		"y_axis": map[string]any{"min": -5, "max": 5},
		"graphs": []any{
			map[string]any{
				"id":     "1",
				"points": []any{point(-1, -1), point(0, 0), point(1, 1)},
			},
		},
	}

	if err := ValidateVisualData(Target{OgeNumber: 11, SubtypeCode: "graphs_linear"}, data); err != nil {
		t.Fatalf("expected valid graph visual_data: %v", err)
	}
}

func TestValidateGraphVisualDataWithSeparatePlots(t *testing.T) {
	data := VisualData{
		"type":   string(VisualKindGraph),
		"x_axis": map[string]any{"min": -5, "max": 5},
		"y_axis": map[string]any{"min": -5, "max": 5},
		"plots": []any{
			map[string]any{
				"id":     "A",
				"points": []any{point(-1, 1), point(0, 0), point(1, 1)},
			},
			map[string]any{
				"id":     "B",
				"points": []any{point(-1, -1), point(0, 0), point(1, -1)},
			},
			map[string]any{
				"id":     "C",
				"points": []any{point(-1, 0), point(0, 1), point(1, 0)},
			},
		},
	}

	if err := ValidateVisualData(Target{OgeNumber: 11, SubtypeCode: "graphs_match"}, data); err != nil {
		t.Fatalf("expected valid graph plots visual_data: %v", err)
	}
}

func TestValidateGraphVisualDataRejectsTooFewPoints(t *testing.T) {
	data := VisualData{
		"type":   string(VisualKindGraph),
		"x_axis": map[string]any{"min": -5, "max": 5},
		"y_axis": map[string]any{"min": -5, "max": 5},
		"graphs": []any{
			map[string]any{
				"id":     "1",
				"points": []any{point(0, 0), point(1, 1)},
			},
		},
	}

	if err := ValidateVisualData(Target{OgeNumber: 11, SubtypeCode: "graphs_linear"}, data); err == nil {
		t.Fatal("expected graph validation error")
	}
}

func TestNormalizeGraphVisualDataSplitsLegacyGraphsIntoPlots(t *testing.T) {
	content := &GeneratedContent{
		VisualData: VisualData{
			"type":   string(VisualKindGraph),
			"x_axis": map[string]any{"min": -5, "max": 5},
			"y_axis": map[string]any{"min": -5, "max": 5},
			"graphs": []any{
				map[string]any{
					"id":     "A",
					"points": []any{point(-1, 1), point(0, 0), point(1, 1)},
				},
				map[string]any{
					"id":     "B",
					"points": []any{point(-1, -1), point(0, 0), point(1, -1)},
				},
			},
		},
	}

	data, err := NormalizeAndValidateVisualData(Target{OgeNumber: 11, SubtypeCode: "graphs_match"}, content)
	if err != nil {
		t.Fatalf("expected graph visual_data to normalize: %v", err)
	}

	plots, ok := data["plots"].([]any)
	if !ok {
		t.Fatalf("expected normalized plots array, got %T", data["plots"])
	}
	if len(plots) != 2 {
		t.Fatalf("expected 2 plots, got %d", len(plots))
	}
}

func TestValidateNumberLineRejectsInvalidInterval(t *testing.T) {
	data := VisualData{
		"type":     string(VisualKindNumberLine),
		"axis":     map[string]any{"min": -5, "max": 5},
		"interval": map[string]any{"start": 3, "end": 3, "start_closed": true, "end_closed": false},
	}

	if err := ValidateVisualData(Target{OgeNumber: 7, SubtypeCode: "numberline_compare"}, data); err == nil {
		t.Fatal("expected number_line validation error")
	}
}

func TestValidateGeometryRejectsLabelMismatch(t *testing.T) {
	data := VisualData{
		"type":     string(VisualKindGeometry),
		"shape":    "triangle",
		"vertices": []any{vertex("A", 0, 0), vertex("B", 1, 0), vertex("C", 0, 1)},
		"labels":   map[string]any{"A": "A", "B": "B", "D": "D"},
	}

	if err := ValidateVisualData(Target{OgeNumber: 15, SubtypeCode: "triangles_area"}, data); err == nil {
		t.Fatal("expected geometry validation error")
	}
}

func TestValidateGridVisualData(t *testing.T) {
	data := VisualData{
		"type":   string(VisualKindGrid),
		"width":  8,
		"height": 6,
		"points": []any{
			map[string]any{"label": "A", "x": 1, "y": 1},
			map[string]any{"label": "B", "x": 4, "y": 5},
		},
		"segments": []any{
			map[string]any{"from": "A", "to": "B"},
		},
	}

	if err := ValidateVisualData(Target{OgeNumber: 18, SubtypeCode: "grid_distance"}, data); err != nil {
		t.Fatalf("expected valid grid visual_data: %v", err)
	}
}

func TestOge19DoesNotRequireVisualData(t *testing.T) {
	target := Target{OgeNumber: 19, SubtypeCode: "logic_angles"}

	if RequiresVisualData(target) {
		t.Fatal("expected OGE 19 to work without visual_data")
	}
}

func TestFallbackGeometryContentMatchesOge17RectangleVisual(t *testing.T) {
	content := FallbackGeneratedContent(Target{OgeNumber: 17, SubtypeCode: "quad_rectangle"}, nil)

	if !strings.Contains(strings.ToLower(content.Question), "прямоугольник") {
		t.Fatalf("expected rectangle fallback question, got %q", content.Question)
	}
	if strings.Contains(strings.ToLower(content.Question), "треугольник") {
		t.Fatalf("fallback question should not mention triangle for OGE 17: %q", content.Question)
	}
	if content.CorrectAnswer != "20" {
		t.Fatalf("expected rectangle perimeter answer 20, got %q", content.CorrectAnswer)
	}
	if got := content.VisualData["shape"]; got != "rectangle" {
		t.Fatalf("expected rectangle visual shape, got %v", got)
	}
}
