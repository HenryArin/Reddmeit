package services

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/HenryArin/ReddmeitAlpha/models"
)

// ApplyPlan subscribes and unsubscribes based on the AI's recommendation plan.
func ApplyPlan(plan models.RecommendationPlan, accessToken string) {
	client := &http.Client{}

	for _, sub := range plan.ToAdd {
		performSubredditAction(client, accessToken, "sub", sub)
	}

	for _, sub := range plan.ToRemove {
		performSubredditAction(client, accessToken, "unsub", sub)
	}
}

// performSubredditAction calls the Reddit API to subscribe or unsubscribe
func performSubredditAction(client *http.Client, accessToken, action, subreddit string) {
	// Clean the subreddit name - remove "r/" prefix if present
	cleanSubreddit := strings.TrimPrefix(subreddit, "r/")

	form := url.Values{}
	form.Set("action", action)
	form.Set("sr_name", cleanSubreddit)

	req, _ := http.NewRequest("POST", "https://oauth.reddit.com/api/subscribe", strings.NewReader(form.Encode()))
	req.Header.Set("Authorization", "bearer "+accessToken)
	req.Header.Set("User-Agent", "reddmeitalpha/0.1")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ Failed %s on %s: %v\n", action, cleanSubreddit, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("⚠️  Reddit API returned %d for %s %s\n", resp.StatusCode, action, cleanSubreddit)
	} else {
		fmt.Printf("✅ %s: r/%s\n", strings.ToUpper(action), cleanSubreddit)
	}
}
