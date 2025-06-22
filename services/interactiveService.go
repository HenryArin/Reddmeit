package services

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/HenryArin/ReddmeitAlpha/models"
	"github.com/HenryArin/ReddmeitAlpha/utils"
)

// RunInteractiveSession handles the interactive user loop for AI subreddit planning
func RunInteractiveSession() error {
	utils.LoadEnv()
	token := os.Getenv("REDDIT_ACCESS_TOKEN")
	user := os.Getenv("REDDIT_USERNAME")
	if token == "" || user == "" {
		return fmt.Errorf("missing REDDIT_ACCESS_TOKEN or REDDIT_USERNAME")
	}

	reader := bufio.NewReader(os.Stdin)
	finalPlan := models.RecommendationPlan{}

	for {
		fmt.Print("🧠 What are you into? (or ask 'show subs')\n> ")
		prompt, _ := reader.ReadString('\n')
		prompt = strings.TrimSpace(prompt)
		lowerPrompt := strings.ToLower(prompt)

		// 🆕 Check for "show", "review", or "summary"
		if lowerPrompt == "show" || lowerPrompt == "review" || lowerPrompt == "summary" {
			fmt.Println("\n📋 Current plan so far:")
			utils.PrintPlan(finalPlan)
			fmt.Print("Would you like to add or remove anything else? (yes/no)\n> ")
			resp, _ := reader.ReadString('\n')
			resp = strings.ToLower(strings.TrimSpace(resp))
			if resp != "yes" {
				break
			}
			continue
		}

		// Fetch current user activity
		subscribed := FetchSubscribedSubreddits(token)
		upvoted := FetchUserActivity(user, token, "upvoted")
		commented := FetchUserActivity(user, token, "comments")

		// Get AI recommendation
		result, err := HandleRequest(prompt, subscribed, upvoted, commented)
		if err != nil {
			return fmt.Errorf("assistant error: %w", err)
		}

		if result.ViewOnly {
			fmt.Println(result.Reply)
		} else {
			fmt.Println("🤖 AI recommendations:")
			utils.PrintPlan(result.Plan)
			finalPlan = utils.MergePlans(finalPlan, result.Plan)
		}

		fmt.Print("Would you like to add or remove anything else? (yes/no)\n> ")
		resp, _ := reader.ReadString('\n')
		resp = strings.ToLower(strings.TrimSpace(resp))
		if resp != "yes" {
			break
		}
	}

	// Final confirmation
	fmt.Println("\n✅ Final Recommendation:")
	utils.PrintPlan(finalPlan)

	fmt.Print("⚠️  Apply these changes? (yes/no)\n> ")
	confirm, _ := reader.ReadString('\n')
	confirm = strings.ToLower(strings.TrimSpace(confirm))

	if confirm != "yes" {
		fmt.Println("❌ Changes canceled.")
		return nil
	}

	// Save and apply
	utils.SavePlanToFile(finalPlan, "interactive_session")
	ApplyPlan(finalPlan, token)

	// Summary
	if len(finalPlan.ToAdd) > 0 {
		fmt.Printf("✅ Added %d subreddit(s)\n", len(finalPlan.ToAdd))
	}
	if len(finalPlan.ToRemove) > 0 {
		fmt.Printf("✅ Removed %d subreddit(s)\n", len(finalPlan.ToRemove))
	}
	if len(finalPlan.ToAdd) == 0 && len(finalPlan.ToRemove) == 0 {
		fmt.Println("✅ No changes needed - your subreddit list looks good!")
	}

	fmt.Println("\n🎉 All done!")
	return nil
}
