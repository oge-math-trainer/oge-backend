package main

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/oge-math-trainer/oge-backend.git/ai"
	"github.com/oge-math-trainer/oge-backend.git/internal/tasks"
)

func TestChat(t *testing.T) {
	client := requireAIClient(t)

	result, err := client.Chat(
		"Respond ONLY with valid JSON. No markdown. Return: {\"topic\": \"string\", \"task_number\": 0, \"condition\": \"string\", \"answer\": \"string\", \"hint1\": \"string\", \"explanation\": \"string\"}. All text in Russian.",
		"{\"topic\": \"Линейные уравнения\", \"task_number\": 4}",
		800,
		0.7,
	)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(result)
}

func TestGenerateTask(t *testing.T) {
	client := requireAIClient(t)

	target := tasks.Target{
		OgeNumber:   6,
		SubtypeCode: "numbers_integers",
	}

	result, err := client.GenerateTask(context.Background(), target)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("Question:", result.Question)
	fmt.Println("Answer:", result.CorrectAnswer)
	fmt.Println("IsValid:", result.IsValid)
	fmt.Println("Steps:", result.SolutionSteps)
}

func TestHint(t *testing.T) {
	client := requireAIClient(t)

	task := tasks.Task{
		Question:      "Решите уравнение 3(x - 4) + 2x = 7x - 10.",
		CorrectAnswer: "-1",
	}

	result, err := client.Hint(context.Background(), task)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("Hint:", result.Hint)
}

func TestExplain(t *testing.T) {
	client := requireAIClient(t)

	task := tasks.Task{
		Question:      "Решите уравнение 3(x - 4) + 2x = 7x - 10.",
		CorrectAnswer: "-1",
	}

	result, err := client.Explain(context.Background(), task)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("Explanation:", result.Explanation)
	fmt.Println("Steps:", result.Steps)
}

func requireAIClient(t *testing.T) *ai.Client {
	t.Helper()

	_ = godotenv.Load(".env")
	if os.Getenv("AITUNNEL_API_KEY") == "" {
		t.Skip("AITUNNEL_API_KEY is not set; skipping external AI integration test")
	}

	client, err := ai.NewClient()
	if err != nil {
		t.Fatal(err)
	}
	return client
}
