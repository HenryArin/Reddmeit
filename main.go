package main

import (
	"fmt"
	"os"

	"github.com/HenryArin/ReddmeitAlpha/controllers"
	"github.com/HenryArin/ReddmeitAlpha/services"
	"github.com/HenryArin/ReddmeitAlpha/utils"
)

func main() {
	// Load .env variables like REDDIT_ACCESS_TOKEN, REDDIT_USERNAME, OPENAI_API_KEY
	utils.LoadEnv()

	accessToken := os.Getenv("REDDIT_ACCESS_TOKEN")
	username := os.Getenv("REDDIT_USERNAME")

	if accessToken == "" || username == "" {
		fmt.Println("❌ Missing environment variables. Make sure .env is set correctly.")
		return
	}

	fmt.Println("📥 Fetching subreddit data...")

	// Fetch subreddit data from Reddit API
	subscribed := services.FetchSubscribedSubreddits(accessToken)
	upvoted := services.FetchUserActivity(username, accessToken, "upvoted")
	commented := services.FetchUserActivity(username, accessToken, "comments")

	// Combine stats into a single map
	combined := controllers.CombineSubredditStats(subscribed, upvoted, commented)

	fmt.Println("\n🔍 Combined Subreddit Interaction Stats:")
	for name, stat := range combined {
		fmt.Printf("r/%s - Subscribed: %v, Upvoted: %v, Commented: %v\n",
			name, stat.Subscribed, stat.Upvoted, stat.Commented)
	}

	// Filter by active subreddits (score >= 2)
	activeList := controllers.FilterActiveSubreddits(combined, 2)

	fmt.Println("\n🔥 Most Active Subreddits:")
	for _, sub := range activeList {
		fmt.Println(sub)
	}

	// Get OpenAI GPT-based subreddit recommendations
	fmt.Println("\n🤖 Recommended Subreddits (AI):")
	recommendations := services.GenerateSubredditRecommendations(activeList)
	fmt.Println(recommendations)
}
