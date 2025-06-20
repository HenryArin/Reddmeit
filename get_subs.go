package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

type SubredditStats struct {
	Name       string
	Subscribed bool
	Upvoted    bool
	Commented  bool
}

type SubredditData struct {
	DisplayName string `json:"display_name"`
	Title       string `json:"title"`
}

type SubredditChild struct {
	Data SubredditData `json:"data"`
}

type SubredditResponse struct {
	Data struct {
		Children []SubredditChild `json:"children"`
		After    string           `json:"after"`
	} `json:"data"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	accessToken := os.Getenv("REDDIT_ACCESS_TOKEN")
	if accessToken == "" {
		log.Fatal("Missing REDDIT_ACCESS_TOKEN in .env")
	}

	username := os.Getenv("REDDIT_USERNAME")
	if username == "" {
		log.Fatal("Missing REDDIT_USERNAME in .env")
	}

	// Fetch all 3 sets
	subscribed := fetchSubscribedSubreddits(accessToken)
	upvoted := fetchUpvotedSubreddits(username, accessToken)
	commented := fetchCommentedSubreddits(username, accessToken)

	// Merge them into a combined map
	combined := make(map[string]*SubredditStats)

	for sub := range subscribed {
		if combined[sub] == nil {
			combined[sub] = &SubredditStats{Name: sub}
		}
		combined[sub].Subscribed = true
	}

	for sub := range upvoted {
		if combined[sub] == nil {
			combined[sub] = &SubredditStats{Name: sub}
		}
		combined[sub].Upvoted = true
	}

	for sub := range commented {
		if combined[sub] == nil {
			combined[sub] = &SubredditStats{Name: sub}
		}
		combined[sub].Commented = true
	}

	// Print combined stats
	fmt.Println("\n🔍 Combined Subreddit Interaction Stats:")
	for _, stat := range combined {
		fmt.Printf("r/%s - Subscribed: %v, Upvoted: %v, Commented: %v\n",
			stat.Name, stat.Subscribed, stat.Upvoted, stat.Commented)
	}
}

func fetchSubscribedSubreddits(accessToken string) map[string]bool {
	client := &http.Client{}
	after := ""
	subreddits := make(map[string]bool)

	for {
		url := "https://oauth.reddit.com/subreddits/mine/subscriber?limit=100"
		if after != "" {
			url += "&after=" + after
		}

		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Authorization", "bearer "+accessToken)
		req.Header.Set("User-Agent", "reddmeitalpha/0.1")

		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		var parsed SubredditResponse
		json.Unmarshal(body, &parsed)

		for _, child := range parsed.Data.Children {
			subreddits[child.Data.DisplayName] = true
		}

		if parsed.Data.After == "" || parsed.Data.After == "null" {
			break
		}
		after = parsed.Data.After
	}

	return subreddits
}

func fetchUpvotedSubreddits(username, accessToken string) map[string]bool {
	client := &http.Client{}
	url := fmt.Sprintf("https://oauth.reddit.com/user/%s/upvoted?limit=100", username)
	headers := map[string]string{
		"Authorization": "bearer " + accessToken,
		"User-Agent":    "reddmeitalpha/0.1",
	}

	subreddits := make(map[string]bool)
	after := ""

	for {
		req, _ := http.NewRequest("GET", url, nil)
		for k, v := range headers {
			req.Header.Set(k, v)
		}

		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		var parsed map[string]interface{}
		json.Unmarshal(body, &parsed)

		data := parsed["data"].(map[string]interface{})
		children := data["children"].([]interface{})

		for _, child := range children {
			c := child.(map[string]interface{})["data"].(map[string]interface{})
			subreddit := c["subreddit"].(string)
			subreddits[subreddit] = true
		}

		afterRaw := data["after"]
		if afterRaw == nil || afterRaw == "" {
			break
		}
		after = afterRaw.(string)
		url = fmt.Sprintf("https://oauth.reddit.com/user/%s/upvoted?limit=100&after=%s", username, after)
	}

	return subreddits
}

func fetchCommentedSubreddits(username, accessToken string) map[string]bool {
	client := &http.Client{}
	url := fmt.Sprintf("https://oauth.reddit.com/user/%s/comments?limit=100", username)
	headers := map[string]string{
		"Authorization": "bearer " + accessToken,
		"User-Agent":    "reddmeitalpha/0.1",
	}

	subreddits := make(map[string]bool)
	after := ""

	for {
		req, _ := http.NewRequest("GET", url, nil)
		for k, v := range headers {
			req.Header.Set(k, v)
		}

		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		var parsed map[string]interface{}
		json.Unmarshal(body, &parsed)

		data := parsed["data"].(map[string]interface{})
		children := data["children"].([]interface{})

		for _, child := range children {
			c := child.(map[string]interface{})["data"].(map[string]interface{})
			subreddit := c["subreddit"].(string)
			subreddits[subreddit] = true
		}

		afterRaw := data["after"]
		if afterRaw == nil || afterRaw == "" {
			break
		}
		after = afterRaw.(string)
		url = fmt.Sprintf("https://oauth.reddit.com/user/%s/comments?limit=100&after=%s", username, after)
	}

	return subreddits
}
