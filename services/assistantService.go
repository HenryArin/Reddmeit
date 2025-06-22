package services

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	openai "github.com/sashabaranov/go-openai"
	"github.com/HenryArin/ReddmeitAlpha/controllers"
	"github.com/HenryArin/ReddmeitAlpha/models"
	"github.com/HenryArin/ReddmeitAlpha/utils"
)

// AssistantResult bundles the assistant output.
type AssistantResult struct {
	ViewOnly bool
	Reply    string
	Plan     models.RecommendationPlan
}

// HandleRequest processes the user input and returns subreddit recommendations or views.
func HandleRequest(userPrompt string, subscribed, upvoted, commented map[string]bool) (AssistantResult, error) {
	intent := controllers.ParseConversationIntent(userPrompt)

	if intent.ShowSubList {
		reply := buildSubsListing(subscribed, upvoted, commented)
		return AssistantResult{ViewOnly: true, Reply: reply}, nil
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return AssistantResult{}, fmt.Errorf("OPENAI_API_KEY unset")
	}
	client := openai.NewClient(apiKey)

	active := controllers.FilterActiveSubreddits(
		controllers.CombineSubredditStats(subscribed, upvoted, commented),
		2,
	)

	var activeNames []string
	for _, s := range active {
		activeNames = append(activeNames, strings.TrimPrefix(s, "r/"))
	}

	system := openai.ChatCompletionMessage{
		Role: openai.ChatMessageRoleSystem,
		Content: `You are a Reddit assistant.
1. When suggesting subs to ADD, format like: "+ r/Subreddit – explanation why it's relevant".
2. When suggesting subs to REMOVE, format like: "- r/Subreddit – why it's no longer relevant".
Only remove subreddits clearly related to the user's topic.
Keep the explanation brief.`,
	}

	user := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: BuildPrompt(userPrompt, intent, subscribed),
	}

	req := openai.ChatCompletionRequest{
		Model:    openai.GPT4o,
		Messages: []openai.ChatCompletionMessage{system, user},
	}
	resp, err := client.CreateChatCompletion(context.Background(), req)
	if err != nil {
		return AssistantResult{}, err
	}
	raw := resp.Choices[0].Message.Content

	plan := utils.ParseSubredditPlan(raw)

	// Allow removal only if intent detected
	if !intent.RemoveMode {
		plan.ToRemove = nil
	}

	// Remove any subs the user said "keep"
	lp := strings.ToLower(userPrompt)
	if strings.Contains(lp, "keep") {
		words := strings.Fields(lp)
		for _, word := range words {
			word = strings.Trim(strings.ToLower(word), ".,!?")
			if strings.HasPrefix(word, "r/") {
				sub := word
				newRemove := []string{}
				removed := false
				for _, r := range plan.ToRemove {
					if r != sub {
						newRemove = append(newRemove, r)
					} else {
						fmt.Printf("🛑 Removed %s from unsubscribe list.\n", sub)
						removed = true
					}
				}
				if removed {
					plan.ToRemove = newRemove
				}
			}
		}
	}

	// Remove duplicates (already subscribed)
	var filtered []string
	for _, sub := range plan.ToAdd {
		name := strings.TrimPrefix(sub, "r/")
		if !subscribed[name] {
			filtered = append(filtered, sub)
		}
	}
	plan.ToAdd = filtered

	// If there's nothing to add/remove, say so
	if len(plan.ToAdd) == 0 && len(plan.ToRemove) == 0 {
		return AssistantResult{
			ViewOnly: false,
			Reply:    "🤖 No strong subreddit matches. Try rephrasing or being more specific?",
			Plan:     plan,
		}, nil
	}

	return AssistantResult{ViewOnly: false, Reply: raw, Plan: plan}, nil
}

// buildSubsListing returns a formatted list of your current subs
func buildSubsListing(subscribed, upvoted, commented map[string]bool) string {
	var subsList, upvotedList, commentedList []string

	for sub := range subscribed {
		subsList = append(subsList, "= r/"+sub)
	}
	for sub := range upvoted {
		if !subscribed[sub] {
			upvotedList = append(upvotedList, "* r/"+sub+" (upvoted-only)")
		}
	}
	for sub := range commented {
		if !subscribed[sub] && !upvoted[sub] {
			commentedList = append(commentedList, "* r/"+sub+" (commented-only)")
		}
	}

	sort.Strings(subsList)
	sort.Strings(upvotedList)
	sort.Strings(commentedList)

	var sb strings.Builder
	sb.WriteString("📋 Your active subreddits:\n")
	for _, line := range subsList {
		sb.WriteString(line + "\n")
	}
	if len(upvotedList) > 0 {
		sb.WriteString("\n🔼 Upvoted-only subs:\n")
		for _, line := range upvotedList {
			sb.WriteString(line + "\n")
		}
	}
	if len(commentedList) > 0 {
		sb.WriteString("\n🗨️  Commented-only subs:\n")
		for _, line := range commentedList {
			sb.WriteString(line + "\n")
		}
	}

	return sb.String()
}
