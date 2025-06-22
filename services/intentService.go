package services

import (
	"context"
	"os"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

type IntentType string

const (
	ShowSubs           IntentType = "show_subs"
	RegenerateAdds     IntentType = "regenerate_adds"
	RegenerateRemoves  IntentType = "regenerate_removes"
	ClearRemoves       IntentType = "clear_removes"
	NewPrompt          IntentType = "new_prompt"
	RefineRemoves      IntentType = "refine_removes"
	RemoveOnlyIntent   IntentType = "remove_only"
	None               IntentType = "none"
)

func GetIntentFromGPT(input string) IntentType {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		panic("Missing OPENAI_API_KEY")
	}

	client := openai.NewClient(apiKey)

	system := openai.ChatCompletionMessage{
		Role: openai.ChatMessageRoleSystem,
		Content: `
You are an intent classifier for a Reddit assistant. Respond with ONE of the following keywords only:

- "show_subs" → if they want to see current subreddit subscriptions
- "regenerate_adds" → if they want more subreddit suggestions to add
- "regenerate_removes" → if they want more subs to remove
- "refine_removes" → if they want to remove some and keep others
- "remove_only" → if they only want to prune or clean inactive subs
- "clear_removes" → if they want to cancel/remove all previous removals
- "new_prompt" → if it's a completely new topic
- "none" → if it's just confirmation or mild feedback

Return one keyword only. No extra text.
`,
	}

	user := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: input,
	}

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{system, user},
		},
	)
	if err != nil {
		panic("Intent GPT error: " + err.Error())
	}

	result := strings.TrimSpace(resp.Choices[0].Message.Content)

	switch IntentType(result) {
	case ShowSubs, RegenerateAdds, RegenerateRemoves, ClearRemoves,
		NewPrompt, RefineRemoves, RemoveOnlyIntent, None:
		return IntentType(result)
	default:
		return None
	}
}
