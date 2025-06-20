package services

import (
	"context"
	"log"
	"os"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

func GenerateSubredditRecommendations(subs []string) string {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("Missing OPENAI_API_KEY in environment")
	}

	client := openai.NewClient(apiKey)

	prompt := buildPrompt(subs)

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)

	if err != nil {
		log.Fatalf("OpenAI API error: %v", err)
	}

	return resp.Choices[0].Message.Content
}

func buildPrompt(subs []string) string {
	var sb strings.Builder
	sb.WriteString("Based on the following subreddits the user is active in:\n\n")
	for _, sub := range subs {
		sb.WriteString("r/" + sub + "\n")
	}
	sb.WriteString("\nPlease recommend subreddits they might enjoy, and briefly explain why.")
	return sb.String()
}
