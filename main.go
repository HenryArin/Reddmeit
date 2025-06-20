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
	utils.LoadEnv()

	accessToken := os.Getenv("REDDIT_ACCESS_TOKEN")
	username := os.Getenv("REDDIT_USERNAME")

	if accessToken == "" || username == "" {
		fmt.Println("❌ Missing environment variables.")
		return
	}

	// Step 1: Ask user for interests
	fmt.Print("🧠 What are you into lately? (describe your Reddit interests)\n> ")
	reader := bufio.NewReader(os.Stdin)
	userPrompt, _ := reader.ReadString('\n')
	userPrompt = strings.TrimSpace(userPrompt)

	// Step 2: Fetch subreddit activity
	fmt.Println("📥 Fetching your subreddit activity...")
	subscribed := services.FetchSubscribedSubreddits(accessToken)
	upvoted := services.FetchUserActivity(username, accessToken, "upvoted")
	commented := services.FetchUserActivity(username, accessToken, "comments")

	combined := controllers.CombineSubredditStats(subscribed, upvoted, commented)
	activeList := controllers.FilterActiveSubreddits(combined, 2)

	var activeNames []string
	for _, sub := range activeList {
		activeNames = append(activeNames, strings.TrimPrefix(sub, "r/"))
	}

	// Step 3: Build GPT prompt and message history
	var messages []openai.ChatCompletionMessage
	initialPrompt := services.BuildPrompt(userPrompt, activeNames)

	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: initialPrompt,
	})

	// Step 4: Initial GPT response
	reply, messages := services.RefineRecommendationsWithMemory(messages)
	fmt.Println("\n📝 Suggested Changes:")
	fmt.Println(reply)
	parsed := controllers.ParseRecommendationOutput(reply)

	// Step 5: Feedback loop
	feedbackReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("\n🧾 Current Plan:")

		if len(parsed.Add) > 0 {
			fmt.Println("To Add:")
			for _, sub := range parsed.Add {
				fmt.Println(" +", sub)
			}
		}
		if len(parsed.Remove) > 0 {
			fmt.Println("To Remove:")
			for _, sub := range parsed.Remove {
				fmt.Println(" -", sub)
			}
		}
		if len(parsed.Keep) > 0 {
			fmt.Println("Kept:")
			for _, sub := range parsed.Keep {
				fmt.Println(" =", sub)
			}
		}

		fmt.Print("\n💬 Feedback? (say anything — GPT will revise), or press [enter] to confirm:\n> ")
		feedback, _ := feedbackReader.ReadString('\n')
		feedback = strings.TrimSpace(feedback)

		if feedback == "" {
			// Final summary before confirmation
			fmt.Println("\n📦 Final Summary:")
			if len(parsed.Add) > 0 {
				fmt.Println("✅ Add:")
				for _, sub := range parsed.Add {
					fmt.Println(" +", sub)
				}
			}
			if len(parsed.Remove) > 0 {
				fmt.Println("❌ Remove:")
				for _, sub := range parsed.Remove {
					fmt.Println(" -", sub)
				}
			}
			if len(parsed.Keep) > 0 {
				fmt.Println("✔ Keep:")
				for _, sub := range parsed.Keep {
					fmt.Println(" =", sub)
				}
			}

			fmt.Println("\n✅ Finalized plan confirmed.")
			break
		}

		// Add feedback to message history
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: feedback,
		})

		fmt.Println("🤖 Asking GPT to revise based on your feedback...")

		reply, messages = services.RefineRecommendationsWithMemory(messages)
		fmt.Println("\n📝 Updated Plan from AI:")
		fmt.Println(reply)

		parsed = controllers.ParseRecommendationOutput(reply)
	}

	// Placeholder for Phase 3
	fmt.Println("\n🚀 You're ready to apply these changes via Reddit API!")
}
