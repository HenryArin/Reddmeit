package services

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/HenryArin/ReddmeitAlpha/controllers"
	openai "github.com/sashabaranov/go-openai"
)

// GenerateSubredditRecommendations uses OpenAI to get subreddit changes based on user input
func GenerateSubredditRecommendations(userPrompt string, subscribed map[string]bool) string {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("Missing OPENAI_API_KEY in environment")
	}

	client := openai.NewClient(apiKey)

	// Parse user's intent
	intent := controllers.ParseConversationIntent(userPrompt)

	// Prepare prompt based on user's request
	prompt := BuildPrompt(userPrompt, intent, subscribed)

	// Send request to GPT
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

// BuildPrompt builds a dynamic prompt depending on intent
func BuildPrompt(userPrompt string, intent controllers.Intent, subscribed map[string]bool) string {
	var sb strings.Builder

	if intent.RemoveMode {
		// System-instructed safe prompt for removals
		sb.WriteString(fmt.Sprintf(`The user said: "%s"

They want to remove subreddit topics related to that input.

Your task:
- ONLY suggest subreddit names that clearly relate to the topic the user wants removed.
- DO NOT suggest anything unrelated.
- Format removals as:
  - r/subredditname - short explanation
- Also provide explanations in parentheses so they can decide.

Their current subscriptions are:
%s
`, userPrompt, formatSubList(subscribed)))
	} else {
		// Default discovery prompt
		sb.WriteString("The user gave the following prompt describing their interests:\n\n")
		sb.WriteString(userPrompt + "\n\n")
		sb.WriteString("The user is currently active in these subreddits:\n")
		for name := range subscribed {
			sb.WriteString("r/" + name + "\n")
		}

		sb.WriteString(`
Please recommend subreddit changes using this format:
+ r/something     // to subscribe
- r/oldsubreddit  // to unsubscribe
= r/keepsubreddit // to keep if needed

Avoid commentary or explanation. Keep only subreddit suggestions in output.
`)
	}
	return sb.String()
}

// formatSubList turns map[string]bool into a comma-separated r/sub list
func formatSubList(subs map[string]bool) string {
	var list []string
	for name := range subs {
		list = append(list, "r/"+name)
	}
	return strings.Join(list, ", ")
}

// RefineRecommendationsWithMemory lets GPT iterate with memory on multi-turn conversations
func RefineRecommendationsWithMemory(messages []openai.ChatCompletionMessage) (string, []openai.ChatCompletionMessage) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("Missing OPENAI_API_KEY")
	}

	client := openai.NewClient(apiKey)

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    openai.GPT4o,
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
