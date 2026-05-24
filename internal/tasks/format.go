package tasks

import (
	"fmt"
	"strings"
)

var unsupportedVisualLatexTokens = []string{
	"tikzpicture",
	"pgfplots",
	`\begin{axis}`,
	`\addplot`,
	`\draw`,
	`\filldraw`,
	`\foreach`,
	`\node`,
	`\path`,
}

func ValidateUserFacingText(content GeneratedContent) error {
	if err := validateTextField("question", content.Question); err != nil {
		return err
	}
	for i, step := range content.SolutionSteps {
		if err := validateTextField(fmt.Sprintf("solution_steps[%d]", i), step); err != nil {
			return err
		}
	}
	if err := validateTextField("self_check", content.SelfCheck); err != nil {
		return err
	}
	return nil
}

func validateTextField(field, text string) error {
	normalized := strings.ToLower(strings.TrimSpace(text))
	for _, token := range unsupportedVisualLatexTokens {
		if strings.Contains(normalized, token) {
			return fmt.Errorf("%s contains unsupported visual LaTeX token %q; diagrams must be returned through visual_data", field, token)
		}
	}
	return nil
}
