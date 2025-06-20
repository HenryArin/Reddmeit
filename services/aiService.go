package services

import (
	"context"
	"log"
	"os"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

func GenerateSubredditRecommendations(userPrompt string, activeSubs []string) string {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("Missing OPENAI_API_KEY in environment")
	}

	client := openai.NewClient(apiKey)
	prompt := BuildPrompt(userPrompt, activeSubs)

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

func BuildPrompt(userPrompt string, activeSubs []string) string {
	var sb strings.Builder
	sb.WriteString("The user gave the following prompt describing their interests:\n\n")
	sb.WriteString(userPrompt + "\n\n")

	sb.WriteString("The user is currently active in these subreddits:\n")
	for _, sub := range activeSubs {
		sb.WriteString("r/" + sub + "\n")
	}

	sb.WriteString(`
Please recommend subreddit changes using this format:
+ r/something     // to subscribe
- r/oldsubreddit  // to unsubscribe
= r/keepsubreddit // to keep if needed

Avoid commentary or explanation. Keep only subreddit suggestions in output.
`)
	return sb.String()
}

func RefineRecommendationsWithMemory(messages []openai.ChatCompletionMessage) (string, []openai.ChatCompletionMessage) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("Missing OPENAI_API_KEY")
	}

	client := openai.NewClient(apiKey)

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    openai.GPT3Dot5Turbo,
			Messages: messages,
		},
	)

	if err != nil {
		log.Fatalf("GPT memory error: %v", err)
	}

	reply := resp.Choices[0].Message.Content
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: reply,
	})

	return reply, messages
}
