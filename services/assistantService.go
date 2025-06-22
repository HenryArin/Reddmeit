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
	// 1) If user asked to view subs, short-circuit
	if controllers.ParseConversationIntent(userPrompt).ShowSubList {
		reply := buildSubsListing(subscribed, upvoted, commented)
		return AssistantResult{ViewOnly: true, Reply: reply}, nil
	}

	// 2) Else, use OpenAI to generate suggestions
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
1) Suggest subs to ADD with "+ r/Subreddit" when the user describes new interests.
2) Suggest subs to REMOVE with "- r/Subreddit" when the user explicitly asks to remove, prune, or unsubscribe.
Use no other text.`,
	}
	user := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: BuildPrompt(userPrompt, activeNames),
	}

	req := openai.ChatCompletionRequest{
		Model:    openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{system, user},
	}
	resp, err := client.CreateChatCompletion(context.Background(), req)
	if err != nil {
		return AssistantResult{}, err
	}
	raw := resp.Choices[0].Message.Content

	// Parse subreddit plan
	plan := utils.ParseSubredditPlan(raw)

	lp := strings.ToLower(userPrompt)

	// Only allow removal if user prompt says so
	if !strings.Contains(lp, "remove") && !strings.Contains(lp, "prune") && !strings.Contains(lp, "unsubscribe") {
		plan.ToRemove = nil
	}

	// Override removes if user says "keep r/xyz"
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

	// Filter out already subscribed subs
	var filtered []string
	for _, sub := range plan.ToAdd {
		name := strings.TrimPrefix(sub, "r/")
		if !subscribed[name] {
			filtered = append(filtered, sub)
		}
	}
	plan.ToAdd = filtered

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
