package tasks

import "testing"

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
