package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/oge-math-trainer/oge-backend.git/ai"
)

func TestChat(t *testing.T) {
	godotenv.Load(".env")
	if os.Getenv("AITUNNEL_API_KEY") == "" {
		t.Skip("AITUNNEL_API_KEY is not set; skipping external AI integration test")
	}

	client, err := ai.NewClient()
	if err != nil {
		t.Fatal(err)
	}

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
