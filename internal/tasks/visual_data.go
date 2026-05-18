package tasks

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

type VisualKind string

const (
	VisualKindGraph      VisualKind = "graph"
	VisualKindNumberLine VisualKind = "number_line"
	VisualKindGeometry   VisualKind = "geometry"
	VisualKindGrid       VisualKind = "grid"
)

func VisualKindForTarget(target Target) VisualKind {
	subtype := strings.ToLower(strings.TrimSpace(target.SubtypeCode))
	switch {
	case target.OgeNumber == 11 || strings.HasPrefix(subtype, "graphs_"):
		return VisualKindGraph
	case target.OgeNumber == 7 || strings.Contains(subtype, "numberline") || strings.HasPrefix(subtype, "ineq_") || strings.Contains(subtype, "inequal"):
		return VisualKindNumberLine
	case target.OgeNumber == 18 || strings.HasPrefix(subtype, "grid_"):
		return VisualKindGrid
	case IsGeometryTarget(target):
		return VisualKindGeometry
	default:
		return ""
	}
}

func RequiresVisualData(target Target) bool {
	return VisualKindForTarget(target) != ""
}

func RequiresExtendedAITimeout(target Target) bool {
	return target.OgeNumber == 11 || target.OgeNumber == 19 || IsGeometryTarget(target)
}

func IsGeometryTarget(target Target) bool {
	subtype := strings.ToLower(strings.TrimSpace(target.SubtypeCode))
	if target.OgeNumber >= 15 && target.OgeNumber <= 17 {
		return true
	}
	for _, prefix := range []string{"triangles_", "circle_", "quad_"} {
		if strings.HasPrefix(subtype, prefix) {
			return true
		}
	}
	return false
}

func NormalizeAndValidateVisualData(target Target, content *GeneratedContent) (VisualData, error) {
	if !RequiresVisualData(target) {
		return nil, nil
	}
	if content == nil {
		return nil, fmt.Errorf("visual_data validation failed: generated content is nil")
	}

	data := content.VisualData
	if len(data) == 0 && VisualKindForTarget(target) == VisualKindGraph && len(content.Graphs) > 0 {
		data = visualDataFromLegacyGraphs(content.Graphs)
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("visual_data is required for %s tasks", VisualKindForTarget(target))
	}
	if err := ValidateVisualData(target, data); err != nil {
		return nil, err
	}
	return data, nil
}

func ValidateVisualData(target Target, data VisualData) error {
	kind := VisualKindForTarget(target)
	if kind == "" {
		return nil
	}
	if len(data) == 0 {
		return fmt.Errorf("visual_data must be a non-empty JSON object")
	}

	gotType, err := stringFromMap(data, "type", "visual_data.type")
	if err != nil {
		return err
	}
	if gotType != string(kind) {
		return fmt.Errorf("visual_data.type must be %q, got %q", kind, gotType)
	}

	switch kind {
	case VisualKindGraph:
		return validateGraphVisualData(data)
	case VisualKindNumberLine:
		return validateNumberLineVisualData(data)
	case VisualKindGeometry:
		return validateGeometryVisualData(data)
	case VisualKindGrid:
		return validateGridVisualData(data)
	default:
		return fmt.Errorf("unsupported visual_data.type %q", gotType)
	}
}

func FallbackGeneratedContent(target Target, reason error) GeneratedContent {
	note := "generated from static fallback"
	if reason != nil {
		note = "generated from static fallback after AI validation error: " + reason.Error()
	}

	switch VisualKindForTarget(target) {
	case VisualKindGraph:
		return GeneratedContent{
			Question:        "На рисунке изображен график функции y = x. Найдите значение y при x = 2.",
			CorrectAnswer:   "2",
			SolutionSteps:   []string{"Подставим x = 2 в формулу y = x.", "Получаем y = 2.", "Ответ: 2."},
			SelfCheck:       "Точка (2; 2) лежит на прямой y = x.",
			IsValid:         true,
			ValidationNotes: note,
			VisualData:      FallbackVisualData(target),
		}
	case VisualKindNumberLine:
		return GeneratedContent{
			Question:        "На координатной прямой отмечен отрезок от -2 до 3, обе границы включены. Сколько целых чисел содержит этот отрезок?",
			CorrectAnswer:   "6",
			SolutionSteps:   []string{"Перечислим целые числа от -2 до 3.", "Получаем -2, -1, 0, 1, 2, 3.", "Всего 6 чисел."},
			SelfCheck:       "Проверьте, что обе граничные точки закрашены.",
			IsValid:         true,
			ValidationNotes: note,
			VisualData:      FallbackVisualData(target),
		}
	case VisualKindGeometry:
		return GeneratedContent{
			Question:        "В прямоугольном треугольнике ABC катеты AB = 6 и AC = 8. Найдите гипотенузу BC.",
			CorrectAnswer:   "10",
			SolutionSteps:   []string{"По теореме Пифагора BC^2 = AB^2 + AC^2.", "BC^2 = 6^2 + 8^2 = 36 + 64 = 100.", "BC = 10."},
			SelfCheck:       "Тройка 6, 8, 10 является пифагоровой.",
			IsValid:         true,
			ValidationNotes: note,
			VisualData:      FallbackVisualData(target),
		}
	case VisualKindGrid:
		return GeneratedContent{
			Question:        "На клетчатой бумаге отмечены точки A(1; 1) и B(4; 5). Найдите длину отрезка AB, если сторона клетки равна 1.",
			CorrectAnswer:   "5",
			SolutionSteps:   []string{"По горизонтали точки отличаются на 3 клетки.", "По вертикали точки отличаются на 4 клетки.", "По теореме Пифагора AB = sqrt(3^2 + 4^2) = 5."},
			SelfCheck:       "Тройки 3, 4, 5 дают длину 5.",
			IsValid:         true,
			ValidationNotes: note,
			VisualData:      FallbackVisualData(target),
		}
	default:
		return GeneratedContent{
			Question:        "Решите уравнение x + 2 = 5.",
			CorrectAnswer:   "3",
			SolutionSteps:   []string{"Перенесем 2 в правую часть.", "Получаем x = 5 - 2.", "x = 3."},
			SelfCheck:       "3 + 2 = 5.",
			IsValid:         true,
			ValidationNotes: note,
		}
	}
}

func FallbackVisualData(target Target) VisualData {
	switch VisualKindForTarget(target) {
	case VisualKindGraph:
		return VisualData{
			"type":   string(VisualKindGraph),
			"x_axis": map[string]any{"min": -5, "max": 5},
			"y_axis": map[string]any{"min": -5, "max": 5},
			"graphs": []any{
				map[string]any{
					"id":     "1",
					"label":  "y = x",
					"points": []any{point(-2, -2), point(0, 0), point(2, 2)},
				},
			},
		}
	case VisualKindNumberLine:
		return VisualData{
			"type":     string(VisualKindNumberLine),
			"axis":     map[string]any{"min": -5, "max": 5},
			"interval": map[string]any{"start": -2, "end": 3, "start_closed": true, "end_closed": true},
			"points": []any{
				map[string]any{"value": -2, "closed": true, "label": "-2"},
				map[string]any{"value": 3, "closed": true, "label": "3"},
			},
		}
	case VisualKindGeometry:
		return fallbackGeometryVisualData(target)
	case VisualKindGrid:
		return VisualData{
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
	default:
		return nil
	}
}

func fallbackGeometryVisualData(target Target) VisualData {
	subtype := strings.ToLower(strings.TrimSpace(target.SubtypeCode))
	switch {
	case target.OgeNumber == 16 && strings.Contains(subtype, "tangent"):
		return VisualData{
			"type":  string(VisualKindGeometry),
			"shape": "circle_tangent",
			"vertices": []any{
				vertex("A", -4, 0),
				vertex("O", 0, 0),
				vertex("B", -1.8, 2.4),
				vertex("C", -1.8, -2.4),
			},
			"labels": map[string]any{"A": "A", "O": "O", "B": "B", "C": "C"},
			"circles": []any{
				map[string]any{"center": "O", "through": "B"},
			},
			"segments": []any{
				map[string]any{"from": "A", "to": "B", "label": "касательная"},
				map[string]any{"from": "A", "to": "C", "label": "касательная"},
				map[string]any{"from": "O", "to": "B", "label": "r"},
				map[string]any{"from": "O", "to": "C", "label": "r"},
				map[string]any{"from": "A", "to": "O", "style": "dashed"},
			},
		}
	case target.OgeNumber == 16:
		return VisualData{
			"type":  string(VisualKindGeometry),
			"shape": "circle",
			"vertices": []any{
				vertex("O", 0, 0),
				vertex("A", 3, 0),
				vertex("B", -2.2, 2),
			},
			"labels": map[string]any{"O": "O", "A": "A", "B": "B"},
			"circles": []any{
				map[string]any{"center": "O", "through": "A"},
			},
			"segments": []any{
				map[string]any{"from": "O", "to": "A", "label": "r"},
				map[string]any{"from": "A", "to": "B", "label": "хорда"},
			},
		}
	case target.OgeNumber == 17 && strings.Contains(subtype, "trapezoid"):
		return VisualData{
			"type":  string(VisualKindGeometry),
			"shape": "trapezoid",
			"vertices": []any{
				vertex("A", 0, 0),
				vertex("B", 8, 0),
				vertex("C", 6, 3),
				vertex("D", 2, 3),
			},
			"labels": map[string]any{"A": "A", "B": "B", "C": "C", "D": "D"},
			"segments": []any{
				map[string]any{"from": "A", "to": "B"},
				map[string]any{"from": "B", "to": "C"},
				map[string]any{"from": "C", "to": "D"},
				map[string]any{"from": "D", "to": "A"},
			},
		}
	case target.OgeNumber == 17:
		return VisualData{
			"type":  string(VisualKindGeometry),
			"shape": "rectangle",
			"vertices": []any{
				vertex("A", 0, 0),
				vertex("B", 6, 0),
				vertex("C", 6, 4),
				vertex("D", 0, 4),
			},
			"labels": map[string]any{"A": "A", "B": "B", "C": "C", "D": "D"},
			"segments": []any{
				map[string]any{"from": "A", "to": "B"},
				map[string]any{"from": "B", "to": "C"},
				map[string]any{"from": "C", "to": "D"},
				map[string]any{"from": "D", "to": "A"},
			},
		}
	default:
		return VisualData{
			"type":     string(VisualKindGeometry),
			"shape":    "triangle",
			"vertices": []any{vertex("A", 0, 0), vertex("B", 6, 0), vertex("C", 0, 8)},
			"labels":   map[string]any{"A": "A", "B": "B", "C": "C"},
			"segments": []any{
				map[string]any{"from": "A", "to": "B"},
				map[string]any{"from": "A", "to": "C"},
				map[string]any{"from": "B", "to": "C"},
			},
		}
	}
}

func validateGraphVisualData(data VisualData) error {
	if err := validateAxis(data["x_axis"], "visual_data.x_axis"); err != nil {
		return err
	}
	if err := validateAxis(data["y_axis"], "visual_data.y_axis"); err != nil {
		return err
	}

	if rawGraphs, ok := data["graphs"]; ok {
		graphs, err := arrayFromValue(rawGraphs, "visual_data.graphs")
		if err != nil {
			return err
		}
		if len(graphs) == 0 {
			return fmt.Errorf("visual_data.graphs must contain at least one graph")
		}
		for i, rawGraph := range graphs {
			graph, err := objectFromValue(rawGraph, fmt.Sprintf("visual_data.graphs[%d]", i))
			if err != nil {
				return err
			}
			if err := validateGraphSeries(graph, fmt.Sprintf("visual_data.graphs[%d]", i)); err != nil {
				return err
			}
		}
		return nil
	}

	return validateGraphSeries(data, "visual_data")
}

func validateGraphSeries(graph map[string]any, path string) error {
	points, err := arrayFromValue(graph["points"], path+".points")
	if err != nil {
		return err
	}
	if len(points) < 3 {
		return fmt.Errorf("%s.points must contain at least 3 coordinate pairs", path)
	}
	for i, rawPoint := range points {
		if err := validateCoordinatePair(rawPoint, fmt.Sprintf("%s.points[%d]", path, i)); err != nil {
			return err
		}
	}
	return nil
}

func validateNumberLineVisualData(data VisualData) error {
	if err := validateAxis(data["axis"], "visual_data.axis"); err != nil {
		return err
	}

	interval, err := objectFromValue(data["interval"], "visual_data.interval")
	if err != nil {
		return err
	}
	start, err := numberFromMap(interval, "start", "visual_data.interval.start")
	if err != nil {
		return err
	}
	end, err := numberFromMap(interval, "end", "visual_data.interval.end")
	if err != nil {
		return err
	}
	if !(start < end) {
		return fmt.Errorf("visual_data.interval.start must be strictly less than visual_data.interval.end")
	}
	if _, err := boolFromMap(interval, "start_closed", "visual_data.interval.start_closed"); err != nil {
		return err
	}
	if _, err := boolFromMap(interval, "end_closed", "visual_data.interval.end_closed"); err != nil {
		return err
	}

	if rawPoints, ok := data["points"]; ok {
		points, err := arrayFromValue(rawPoints, "visual_data.points")
		if err != nil {
			return err
		}
		for i, rawPoint := range points {
			pointObj, err := objectFromValue(rawPoint, fmt.Sprintf("visual_data.points[%d]", i))
			if err != nil {
				return err
			}
			if _, err := numberFromMap(pointObj, "value", fmt.Sprintf("visual_data.points[%d].value", i)); err != nil {
				return err
			}
			if _, err := boolFromMap(pointObj, "closed", fmt.Sprintf("visual_data.points[%d].closed", i)); err != nil {
				return err
			}
		}
	}

	return nil
}

func validateGeometryVisualData(data VisualData) error {
	if _, err := stringFromMap(data, "shape", "visual_data.shape"); err != nil {
		return err
	}

	vertices, err := arrayFromValue(data["vertices"], "visual_data.vertices")
	if err != nil {
		return err
	}
	if len(vertices) < 3 {
		return fmt.Errorf("visual_data.vertices must contain at least 3 vertices")
	}

	vertexLabels := make(map[string]struct{}, len(vertices))
	for i, rawVertex := range vertices {
		vertexObj, err := objectFromValue(rawVertex, fmt.Sprintf("visual_data.vertices[%d]", i))
		if err != nil {
			return err
		}
		label, err := stringFromMap(vertexObj, "label", fmt.Sprintf("visual_data.vertices[%d].label", i))
		if err != nil {
			return err
		}
		if _, exists := vertexLabels[label]; exists {
			return fmt.Errorf("visual_data.vertices[%d].label duplicates vertex %q", i, label)
		}
		vertexLabels[label] = struct{}{}
		if _, err := numberFromMap(vertexObj, "x", fmt.Sprintf("visual_data.vertices[%d].x", i)); err != nil {
			return err
		}
		if _, err := numberFromMap(vertexObj, "y", fmt.Sprintf("visual_data.vertices[%d].y", i)); err != nil {
			return err
		}
	}

	labels, err := objectFromValue(data["labels"], "visual_data.labels")
	if err != nil {
		return err
	}
	if len(labels) != len(vertexLabels) {
		return fmt.Errorf("visual_data.labels must contain exactly one label for each vertex")
	}
	for label := range vertexLabels {
		rawLabel, ok := labels[label]
		if !ok {
			return fmt.Errorf("visual_data.labels is missing label for vertex %q", label)
		}
		if _, err := stringFromValue(rawLabel, "visual_data.labels."+label); err != nil {
			return err
		}
	}
	for label := range labels {
		if _, ok := vertexLabels[label]; !ok {
			return fmt.Errorf("visual_data.labels contains unknown vertex %q", label)
		}
	}

	if rawSegments, ok := data["segments"]; ok {
		segments, err := arrayFromValue(rawSegments, "visual_data.segments")
		if err != nil {
			return err
		}
		for i, rawSegment := range segments {
			segment, err := objectFromValue(rawSegment, fmt.Sprintf("visual_data.segments[%d]", i))
			if err != nil {
				return err
			}
			from, err := stringFromMap(segment, "from", fmt.Sprintf("visual_data.segments[%d].from", i))
			if err != nil {
				return err
			}
			to, err := stringFromMap(segment, "to", fmt.Sprintf("visual_data.segments[%d].to", i))
			if err != nil {
				return err
			}
			if _, ok := vertexLabels[from]; !ok {
				return fmt.Errorf("visual_data.segments[%d].from references unknown vertex %q", i, from)
			}
			if _, ok := vertexLabels[to]; !ok {
				return fmt.Errorf("visual_data.segments[%d].to references unknown vertex %q", i, to)
			}
		}
	}

	return nil
}

func validateGridVisualData(data VisualData) error {
	width, err := numberFromMap(data, "width", "visual_data.width")
	if err != nil {
		return err
	}
	height, err := numberFromMap(data, "height", "visual_data.height")
	if err != nil {
		return err
	}
	if width <= 0 || height <= 0 {
		return fmt.Errorf("visual_data.width and visual_data.height must be positive")
	}

	points, err := arrayFromValue(data["points"], "visual_data.points")
	if err != nil {
		return err
	}
	if len(points) < 2 {
		return fmt.Errorf("visual_data.points must contain at least 2 points")
	}

	pointLabels := make(map[string]struct{}, len(points))
	for i, rawPoint := range points {
		pointObj, err := objectFromValue(rawPoint, fmt.Sprintf("visual_data.points[%d]", i))
		if err != nil {
			return err
		}
		label, err := stringFromMap(pointObj, "label", fmt.Sprintf("visual_data.points[%d].label", i))
		if err != nil {
			return err
		}
		if _, exists := pointLabels[label]; exists {
			return fmt.Errorf("visual_data.points[%d].label duplicates point %q", i, label)
		}
		pointLabels[label] = struct{}{}
		x, err := numberFromMap(pointObj, "x", fmt.Sprintf("visual_data.points[%d].x", i))
		if err != nil {
			return err
		}
		y, err := numberFromMap(pointObj, "y", fmt.Sprintf("visual_data.points[%d].y", i))
		if err != nil {
			return err
		}
		if x < 0 || x > width || y < 0 || y > height {
			return fmt.Errorf("visual_data.points[%d] must be inside the grid", i)
		}
	}

	segments, err := arrayFromValue(data["segments"], "visual_data.segments")
	if err != nil {
		return err
	}
	if len(segments) == 0 {
		return fmt.Errorf("visual_data.segments must contain at least one segment")
	}
	for i, rawSegment := range segments {
		segment, err := objectFromValue(rawSegment, fmt.Sprintf("visual_data.segments[%d]", i))
		if err != nil {
			return err
		}
		from, err := stringFromMap(segment, "from", fmt.Sprintf("visual_data.segments[%d].from", i))
		if err != nil {
			return err
		}
		to, err := stringFromMap(segment, "to", fmt.Sprintf("visual_data.segments[%d].to", i))
		if err != nil {
			return err
		}
		if _, ok := pointLabels[from]; !ok {
			return fmt.Errorf("visual_data.segments[%d].from references unknown point %q", i, from)
		}
		if _, ok := pointLabels[to]; !ok {
			return fmt.Errorf("visual_data.segments[%d].to references unknown point %q", i, to)
		}
	}

	if rawFill, ok := data["fill"]; ok {
		fill, err := arrayFromValue(rawFill, "visual_data.fill")
		if err != nil {
			return err
		}
		for i, rawLabel := range fill {
			label, err := stringFromValue(rawLabel, fmt.Sprintf("visual_data.fill[%d]", i))
			if err != nil {
				return err
			}
			if _, ok := pointLabels[label]; !ok {
				return fmt.Errorf("visual_data.fill[%d] references unknown point %q", i, label)
			}
		}
	}

	return nil
}

func validateAxis(raw any, path string) error {
	axis, err := objectFromValue(raw, path)
	if err != nil {
		return err
	}
	min, err := numberFromMap(axis, "min", path+".min")
	if err != nil {
		return err
	}
	max, err := numberFromMap(axis, "max", path+".max")
	if err != nil {
		return err
	}
	if !(min < max) {
		return fmt.Errorf("%s.min must be strictly less than %s.max", path, path)
	}
	return nil
}

func validateCoordinatePair(raw any, path string) error {
	if pointObj, err := objectFromValue(raw, path); err == nil {
		if _, err := numberFromMap(pointObj, "x", path+".x"); err != nil {
			return err
		}
		if _, err := numberFromMap(pointObj, "y", path+".y"); err != nil {
			return err
		}
		return nil
	}

	pointArray, err := arrayFromValue(raw, path)
	if err != nil {
		return fmt.Errorf("%s must be an object with x/y numbers or a [x,y] pair", path)
	}
	if len(pointArray) != 2 {
		return fmt.Errorf("%s must contain exactly two coordinates", path)
	}
	if _, err := numberFromValue(pointArray[0], path+"[0]"); err != nil {
		return err
	}
	if _, err := numberFromValue(pointArray[1], path+"[1]"); err != nil {
		return err
	}
	return nil
}

func visualDataFromLegacyGraphs(graphs []GraphInfo) VisualData {
	xMin, xMax, yMin, yMax := -10, 10, -10, 10
	if len(graphs) > 0 && graphs[0].XMin < graphs[0].XMax && graphs[0].YMin < graphs[0].YMax {
		xMin, xMax, yMin, yMax = graphs[0].XMin, graphs[0].XMax, graphs[0].YMin, graphs[0].YMax
	}

	series := make([]any, 0, len(graphs))
	for i, graph := range graphs {
		id := strings.TrimSpace(graph.ID)
		if id == "" {
			id = fmt.Sprintf("%d", i+1)
		}
		series = append(series, map[string]any{
			"id":     id,
			"label":  strings.TrimSpace(graph.Type),
			"points": legacyGraphPoints(graph, xMin, xMax),
		})
	}

	return VisualData{
		"type":   string(VisualKindGraph),
		"x_axis": map[string]any{"min": xMin, "max": xMax},
		"y_axis": map[string]any{"min": yMin, "max": yMax},
		"graphs": series,
	}
}

func legacyGraphPoints(graph GraphInfo, xMin, xMax int) []any {
	xs := []float64{float64(xMin), float64(xMin+xMax) / 2, float64(xMax)}
	if strings.EqualFold(graph.Type, "hyperbola") {
		xs = []float64{-2, -1, 1, 2}
	}

	points := make([]any, 0, len(xs))
	for _, x := range xs {
		y, ok := legacyGraphY(graph, x)
		if !ok {
			continue
		}
		points = append(points, point(x, y))
	}
	if len(points) >= 3 {
		return points
	}
	return []any{point(-2, -2), point(0, 0), point(2, 2)}
}

func legacyGraphY(graph GraphInfo, x float64) (float64, bool) {
	switch strings.ToLower(strings.TrimSpace(graph.Type)) {
	case "linear":
		if len(graph.Coefficients) < 2 {
			return 0, false
		}
		return graph.Coefficients[0]*x + graph.Coefficients[1], true
	case "quadratic":
		if len(graph.Coefficients) < 3 {
			return 0, false
		}
		return graph.Coefficients[0]*x*x + graph.Coefficients[1]*x + graph.Coefficients[2], true
	case "hyperbola":
		if len(graph.Coefficients) < 1 || x == 0 {
			return 0, false
		}
		return graph.Coefficients[0] / x, true
	default:
		return 0, false
	}
}

func point(x, y any) map[string]any {
	return map[string]any{"x": x, "y": y}
}

func vertex(label string, x, y any) map[string]any {
	return map[string]any{"label": label, "x": x, "y": y}
}

func stringFromMap(obj map[string]any, key, path string) (string, error) {
	raw, ok := obj[key]
	if !ok {
		return "", fmt.Errorf("%s is required", path)
	}
	return stringFromValue(raw, path)
}

func stringFromValue(raw any, path string) (string, error) {
	value, ok := raw.(string)
	if !ok {
		return "", fmt.Errorf("%s must be a string", path)
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("%s must not be empty", path)
	}
	return value, nil
}

func numberFromMap(obj map[string]any, key, path string) (float64, error) {
	raw, ok := obj[key]
	if !ok {
		return 0, fmt.Errorf("%s is required", path)
	}
	return numberFromValue(raw, path)
}

func numberFromValue(raw any, path string) (float64, error) {
	var value float64
	switch n := raw.(type) {
	case int:
		value = float64(n)
	case int8:
		value = float64(n)
	case int16:
		value = float64(n)
	case int32:
		value = float64(n)
	case int64:
		value = float64(n)
	case uint:
		value = float64(n)
	case uint8:
		value = float64(n)
	case uint16:
		value = float64(n)
	case uint32:
		value = float64(n)
	case uint64:
		value = float64(n)
	case float32:
		value = float64(n)
	case float64:
		value = n
	case json.Number:
		parsed, err := n.Float64()
		if err != nil {
			return 0, fmt.Errorf("%s must be a number", path)
		}
		value = parsed
	default:
		return 0, fmt.Errorf("%s must be a number", path)
	}
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return 0, fmt.Errorf("%s must be a finite number", path)
	}
	return value, nil
}

func boolFromMap(obj map[string]any, key, path string) (bool, error) {
	raw, ok := obj[key]
	if !ok {
		return false, fmt.Errorf("%s is required", path)
	}
	value, ok := raw.(bool)
	if !ok {
		return false, fmt.Errorf("%s must be a boolean", path)
	}
	return value, nil
}

func objectFromValue(raw any, path string) (map[string]any, error) {
	switch obj := raw.(type) {
	case VisualData:
		return map[string]any(obj), nil
	case map[string]any:
		return obj, nil
	default:
		return nil, fmt.Errorf("%s must be a JSON object", path)
	}
}

func arrayFromValue(raw any, path string) ([]any, error) {
	if raw == nil {
		return nil, fmt.Errorf("%s is required", path)
	}
	switch arr := raw.(type) {
	case []any:
		return arr, nil
	default:
		return nil, fmt.Errorf("%s must be an array", path)
	}
}
