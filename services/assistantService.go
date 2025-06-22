package services

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/HenryArin/ReddmeitAlpha/controllers"
	"github.com/HenryArin/ReddmeitAlpha/models"
	"github.com/HenryArin/ReddmeitAlpha/utils"
	openai "github.com/sashabaranov/go-openai"
)

var lastUserInterest string
var lastPlan models.RecommendationPlan

type AssistantResult struct {
	ViewOnly bool
	Reply    string
	Plan     models.RecommendationPlan
}

func HandleRequest(userPrompt string, intent controllers.Intent, subscribed, upvoted, commented map[string]bool) (AssistantResult, error) {
	if !intent.ShowSubList && strings.TrimSpace(userPrompt) != "" {
		lastUserInterest = userPrompt
	}

	if intent.ShowSubList {
		reply := buildSubsListing(subscribed, upvoted, commented)
		return AssistantResult{ViewOnly: true, Reply: reply}, nil
	}

	if isExclusionOnlyRequest(userPrompt) && hasValidLastPlan() {
		return handleExclusionRequest(userPrompt, lastPlan, subscribed)
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
		Content: `You are a Reddit assistant helping users manage their subreddit subscriptions.

Your task:
1. Suggest subreddit additions and removals in clearly grouped categories.
2. Use emoji category headers to group related subreddits. Examples:
   🥐 Baking:
   💪 Fitness:
   🎮 Gaming:
   🍷 Alcohol:
   🚗 Cars:
   🧠 Learning:
   🧘 Wellness:
   📚 Books:

Formatting Rules:
+ r/Subreddit – short reason (for adds)
- r/Subreddit – short reason (for removes)

3. Always group subreddits under the correct category heading.
4. Only REMOVE subreddits that clearly relate to the user's removal intent.
5. Keep explanations concise and helpful.
6. Output should be readable in markdown/plaintext format — no extra commentary.
7. If no relevant results, respond with:
   🤖 No strong subreddit matches. Try rephrasing or being more specific?`,
	}

	promptInput := userPrompt
	if intent.RegenerateAdds && lastUserInterest != "" {
		promptInput = lastUserInterest
	}

	user := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: BuildPrompt(promptInput, intent, subscribed),
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

	lastPlan = plan
	plan = deduplicatePlan(plan)

	if !intent.RemoveMode {
		var filtered []string
		lp := strings.ToLower(userPrompt)
		for _, sub := range plan.ToRemove {
			name := strings.ToLower(strings.TrimPrefix(sub, "r/"))
			if strings.Contains(lp, name) || strings.Contains(lp, "r/"+name) {
				filtered = append(filtered, sub)
			}
		}
		plan.ToRemove = filtered
	}

	plan = handleKeepRequests(userPrompt, plan)
	plan.ToAdd = filterAlreadySubscribed(plan.ToAdd, subscribed)
	plan = handleExclusions(userPrompt, plan)
	plan = preventOverlap(plan)
	lastPlan = plan

	if len(plan.ToAdd) == 0 && len(plan.ToRemove) == 0 {
		return AssistantResult{
			ViewOnly: false,
			Reply:    "🤖 No strong subreddit matches. Try rephrasing or being more specific?",
			Plan:     plan,
		}, nil
	}

	lastPlan = plan
	return AssistantResult{ViewOnly: false, Reply: raw, Plan: plan}, nil
}

func isExclusionOnlyRequest(userPrompt string) bool {
	lp := strings.ToLower(strings.TrimSpace(userPrompt))
	exclusionKeywords := []string{"don't add", "dont add", "skip", "no "}
	hasExclusionKeyword := false
	for _, keyword := range exclusionKeywords {
		if strings.Contains(lp, keyword) {
			hasExclusionKeyword = true
			break
		}
	}
	if !hasExclusionKeyword {
		return false
	}

	cleaned := lp
	cleaned = strings.ReplaceAll(cleaned, "don't add", "")
	cleaned = strings.ReplaceAll(cleaned, "dont add", "")
	cleaned = strings.ReplaceAll(cleaned, "skip", "")
	cleaned = strings.ReplaceAll(cleaned, "no", "")
	cleaned = strings.ReplaceAll(cleaned, "r/", "")
	cleaned = strings.Trim(cleaned, " .,!?")

	words := strings.Fields(cleaned)
	meaningfulWords := 0
	for _, word := range words {
		if len(word) > 2 && !isCommonWord(word) {
			meaningfulWords++
		}
	}
	return meaningfulWords <= 1
}

func isCommonWord(word string) bool {
	commonWords := map[string]bool{
		"the": true, "and": true, "or": true, "but": true, "in": true,
		"on": true, "at": true, "to": true, "for": true, "of": true,
		"with": true, "by": true, "from": true, "up": true, "about": true,
		"into": true, "through": true, "during": true, "before": true,
		"after": true, "above": true, "below": true, "between": true,
		"among": true, "is": true, "are": true, "was": true, "were": true,
		"be": true, "been": true, "being": true, "have": true, "has": true,
		"had": true, "do": true, "does": true, "did": true, "will": true,
		"would": true, "could": true, "should": true, "may": true, "might": true,
		"must": true, "can": true, "please": true, "thanks": true, "thank": true,
	}
	return commonWords[strings.ToLower(word)]
}

func hasValidLastPlan() bool {
	result := len(lastPlan.ToAdd) > 0 || len(lastPlan.ToRemove) > 0
	fmt.Printf("🔍 Has valid last plan: %t (ToAdd: %d, ToRemove: %d)\n", result, len(lastPlan.ToAdd), len(lastPlan.ToRemove))
	return result
}

func handleExclusionRequest(userPrompt string, plan models.RecommendationPlan, subscribed map[string]bool) (AssistantResult, error) {
	newPlan := models.RecommendationPlan{
		ToAdd:    make([]string, len(plan.ToAdd)),
		ToRemove: []string{},
	}
	copy(newPlan.ToAdd, plan.ToAdd)
	newPlan = handleExclusions(userPrompt, newPlan)
	response := generateExclusionResponse(plan.ToAdd, newPlan.ToAdd)
	lastPlan = newPlan
	return AssistantResult{
		ViewOnly: false,
		Reply:    response,
		Plan:     newPlan,
	}, nil
}

func generateExclusionResponse(originalAdds, filteredAdds []string) string {
	if len(originalAdds) == len(filteredAdds) {
		return "🤖 No changes made. The subreddit you mentioned wasn't in the recommendations."
	}
	excluded := []string{}
	filteredMap := make(map[string]bool)
	for _, sub := range filteredAdds {
		filteredMap[sub] = true
	}
	for _, sub := range originalAdds {
		if !filteredMap[sub] {
			excluded = append(excluded, sub)
		}
	}
	response := "✅ Updated recommendations:\n\n"
	if len(filteredAdds) > 0 {
		response += "📚 To Add:\n"
		for _, sub := range filteredAdds {
			response += fmt.Sprintf("+ %s\n", sub)
		}
	}
	if len(excluded) > 0 {
		response += "\n❌ Excluded based on your feedback:\n"
		for _, sub := range excluded {
			response += fmt.Sprintf("- %s\n", sub)
		}
	}
	return response
}

func handleKeepRequests(userPrompt string, plan models.RecommendationPlan) models.RecommendationPlan {
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
					if strings.ToLower(r) != sub {
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
	return plan
}

func deduplicatePlan(plan models.RecommendationPlan) models.RecommendationPlan {
	seenAdd := make(map[string]bool)
	var uniqueAdd []string
	for _, sub := range plan.ToAdd {
		normalizedSub := strings.ToLower(strings.TrimPrefix(sub, "r/"))
		if !seenAdd[normalizedSub] {
			seenAdd[normalizedSub] = true
			uniqueAdd = append(uniqueAdd, sub)
		}
	}
	seenRemove := make(map[string]bool)
	var uniqueRemove []string
	for _, sub := range plan.ToRemove {
		normalizedSub := strings.ToLower(strings.TrimPrefix(sub, "r/"))
		if !seenRemove[normalizedSub] {
			seenRemove[normalizedSub] = true
			uniqueRemove = append(uniqueRemove, sub)
		}
	}
	return models.RecommendationPlan{
		ToAdd:    uniqueAdd,
		ToRemove: uniqueRemove,
	}
}

func filterAlreadySubscribed(toAdd []string, subscribed map[string]bool) []string {
	var filtered []string
	for _, sub := range toAdd {
		name := strings.TrimPrefix(sub, "r/")
		if !subscribed[name] {
			filtered = append(filtered, sub)
		}
	}
	return filtered
}

func handleExclusions(userPrompt string, plan models.RecommendationPlan) models.RecommendationPlan {
	excludeSubs := make(map[string]bool)
	lp := strings.ToLower(userPrompt)
	if strings.Contains(lp, "don't add") || strings.Contains(lp, "dont add") || strings.Contains(lp, "skip") || strings.Contains(lp, "no ") {
		clean := strings.ReplaceAll(lp, "don't add", "")
		clean = strings.ReplaceAll(clean, "dont add", "")
		clean = strings.ReplaceAll(clean, "skip", "")
		clean = strings.ReplaceAll(clean, "no", "")
		parts := strings.FieldsFunc(clean, func(c rune) bool {
			return c == ',' || c == ' ' || c == '.' || c == '!' || c == '?'
		})
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			if strings.HasPrefix(part, "r/") {
				subName := strings.TrimPrefix(part, "r/")
				excludeSubs[strings.ToLower(subName)] = true
			} else {
				excludeSubs[strings.ToLower(part)] = true
			}
		}
	}
	filteredAdd := []string{}
	for _, sub := range plan.ToAdd {
		name := strings.ToLower(strings.TrimPrefix(sub, "r/"))
		if !excludeSubs[name] {
			filteredAdd = append(filteredAdd, sub)
		} else {
			fmt.Printf("❌ Skipped %s based on your feedback.\n", sub)
		}
	}
	plan.ToAdd = filteredAdd
	return plan
}

func preventOverlap(plan models.RecommendationPlan) models.RecommendationPlan {
	removeMap := make(map[string]bool)
	for _, r := range plan.ToRemove {
		removeMap[strings.ToLower(strings.TrimPrefix(r, "r/"))] = true
	}
	finalAdd := []string{}
	for _, sub := range plan.ToAdd {
		name := strings.ToLower(strings.TrimPrefix(sub, "r/"))
		if !removeMap[name] {
			finalAdd = append(finalAdd, sub)
		}
	}
	plan.ToAdd = finalAdd
	return plan
}

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

// im into books, novels, book collecting, and mangas
