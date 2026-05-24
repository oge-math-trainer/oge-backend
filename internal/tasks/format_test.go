package tasks

import "testing"

func TestValidateUserFacingTextRejectsTikZ(t *testing.T) {
	content := GeneratedContent{
		Question:      `\begin{tikzpicture}\draw (0,0) -- (1,0);\end{tikzpicture}`,
		SolutionSteps: []string{"step 1", "step 2", "step 3"},
		SelfCheck:     "check",
	}

	if err := ValidateUserFacingText(content); err == nil {
		t.Fatal("expected TikZ validation error")
	}
}

func TestValidateUserFacingTextAllowsOrdinaryLatex(t *testing.T) {
	content := GeneratedContent{
		Question:      `Найдите значение \frac{3}{4} + \sqrt{16}.`,
		SolutionSteps: []string{`Вычислим \sqrt{16} = 4.`, `Сложим значения.`, `Получим ответ.`},
		SelfCheck:     `Обычный LaTeX без рисунка допустим.`,
	}

	if err := ValidateUserFacingText(content); err != nil {
		t.Fatalf("expected ordinary LaTeX to be valid: %v", err)
	}
}
