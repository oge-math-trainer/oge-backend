package ai

import "testing"

func TestCleanLaTeX(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple formula",
			input:    "$x^2 + b_1 = 0$",
			expected: "x^2 + b_1 = 0",
		},
		{
			name:     "double dollar",
			input:    "$$formula$$",
			expected: "formula",
		},
		{
			name:     "text with dollars",
			input:    "В ряду $25$ мест",
			expected: "В ряду 25 мест",
		},
		{
			name:     "no dollars",
			input:    "Обычный текст без формул",
			expected: "Обычный текст без формул",
		},
		{
			name:     "mixed content",
			input:    "Найдите $x$, если $x + 5 = 10$",
			expected: "Найдите x, если x + 5 = 10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanLaTeX(tt.input)
			if result != tt.expected {
				t.Errorf("cleanLaTeX(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
