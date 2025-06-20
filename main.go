package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	openai "github.com/sashabaranov/go-openai"

	"github.com/HenryArin/ReddmeitAlpha/controllers"
	"github.com/HenryArin/ReddmeitAlpha/services"
	"github.com/HenryArin/ReddmeitAlpha/utils"
)

func main() {
	// Load environment variables
	utils.LoadEnv()

	// Get credentials
	accessToken := os.Getenv("REDDIT_ACCESS_TOKEN")
	username := os.Getenv("REDDIT_USERNAME")
	if accessToken == "" || username == "" {
		fmt.Println("❌ Missing environment variables.")
		return
	}

	// Prompt user for interests
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("🧠 What are you into lately? (describe your Reddit interests)")
	fmt.Print("> ")
	userPrompt, _ := reader.ReadString('\n')
	userPrompt = strings.TrimSpace(userPrompt)

	// Fetch subreddit activity
	fmt.Println("📥 Fetching your subreddit activity...")
	subscribed := services.FetchSubscribedSubreddits(accessToken)
	upvoted := services.FetchUserActivity(username, accessToken, "upvoted")
	commented := services.FetchUserActivity(username, accessToken, "comments")

	// Combine into active list
	combined := controllers.CombineSubredditStats(subscribed, upvoted, commented)
	activeList := controllers.FilterActiveSubreddits(combined, 2)

	var activeNames []string
	for _, sub := range activeList {
		activeNames = append(activeNames, strings.TrimPrefix(sub, "r/"))
	}

	// Build initial GPT-based prompt and memory
	var messages []openai.ChatCompletionMessage
	systemMsg := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: "You help users manage Reddit subscriptions. Format only as + r/Name, - r/Name, and = r/Name. Don't remove previously suggested items unless asked. No extra commentary.",
	}
	userMsg := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: services.BuildPrompt(userPrompt, activeNames),
	}
	messages = []openai.ChatCompletionMessage{systemMsg, userMsg}

	// Initial GPT suggestion
	fmt.Println("\n🤖 AI is generating recommendations...")
	reply, messages := services.RefineRecommendationsWithMemory(messages)
	plan := utils.ParseSubredditPlan(reply)
	utils.PrintPlan(plan)

	// Feedback loop
	for {
		fmt.Print("\n💬 Feedback? (say anything — GPT will revise), or press [enter] to confirm:\n> ")
		feedback, _ := reader.ReadString('\n')
		feedback = strings.TrimSpace(feedback)
		if feedback == "" {
			break
		}

		// Append user feedback and get revised plan
		messages = append(messages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: feedback})
		reply, messages = services.RefineRecommendationsWithMemory(messages)
		newPlan := utils.ParseSubredditPlan(reply)

		// Merge new suggestions with existing plan
		plan = mergePlans(plan, newPlan)
		utils.PrintPlan(plan)
	}

	// Final summary
	fmt.Println("\n📦 Final Summary:")
	utils.PrintPlan(plan)
	fmt.Println("\n✅ Finalized plan confirmed.")
	fmt.Println("\n🚀 You're ready to apply these changes via Reddit API!")
}

// mergePlans combines two plans, preserving all adds/removes/keeps, with keep overriding remove
func mergePlans(old, new utils.Plan) utils.Plan {
	// Combine lists
	adds := union(old.ToAdd, new.ToAdd)
	removes := union(old.ToRemove, new.ToRemove)
	keeps := union(old.ToKeep, new.ToKeep)

	// If a subreddit is kept, it should not be in removes or adds
	removes = subtract(removes, keeps)
	adds = subtract(adds, keeps)

	// If a subreddit is removed, it should not be in adds
	adds = subtract(adds, removes)

	return utils.Plan{
		ToAdd:    adds,
		ToRemove: removes,
		ToKeep:   keeps,
	}
}

// union returns the deduplicated union of two string slices
func union(a, b []string) []string {
	set := make(map[string]bool)
	for _, x := range a {
		set[x] = true
	}
	for _, x := range b {
		set[x] = true
	}
	result := make([]string, 0, len(set))
	for x := range set {
		result = append(result, x)
	}
	return result
}

// subtract removes any elements in b from a
func subtract(a, b []string) []string {
	setB := make(map[string]bool)
	for _, x := range b {
		setB[x] = true
	}
	result := make([]string, 0)
	for _, x := range a {
		if !setB[x] {
			result = append(result, x)
		}
	}
	return result
}
