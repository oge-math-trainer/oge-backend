package main

import (
	"fmt"
	"testing"

	"github.com/joho/godotenv"
	"github.com/oge-math-trainer/oge-backend.git/ai"
)

func TestChat(t *testing.T) {
	godotenv.Load(".env")

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
