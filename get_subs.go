// ReddmeitAlpha - main.go
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	openai "github.com/sashabaranov/go-openai"
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
	prompt := flag.String("prompt", "", "User interest prompt for subreddit recommendations")
	flag.Parse()

	if *prompt == "" {
		fmt.Println("❗ Please provide a prompt with --prompt")
		os.Exit(1)
	}

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

	subscribed := fetchSubscribedSubreddits(accessToken)
	upvoted := fetchUpvotedSubreddits(username, accessToken)
	commented := fetchCommentedSubreddits(username, accessToken)

	combined := make(map[string]*SubredditStats)
	for sub := range subscribed {
		combined[sub] = &SubredditStats{Name: sub, Subscribed: true}
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

	fmt.Println("\n🔍 Combined Subreddit Interaction Stats:")
	for _, stat := range combined {
		fmt.Printf("r/%s - Subscribed: %v, Upvoted: %v, Commented: %v\n",
			stat.Name, stat.Subscribed, stat.Upvoted, stat.Commented)
	}

	fmt.Println("\n🔥 Most Active Subreddits:")
	var activeList []string
	for _, stat := range combined {
		score := 0
		if stat.Subscribed {
			score++
		}
		if stat.Upvoted {
			score++
		}
		if stat.Commented {
			score++
		}
		if score >= 2 {
			activeList = append(activeList, "r/"+stat.Name)
		}
	}

	fmt.Println("\n🤖 Recommended Subreddits (AI):")
	recommendations := getRecommendations(activeList, *prompt)
	fmt.Println(recommendations)
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
	subreddits := make(map[string]bool)
	after := ""
	for {
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Authorization", "bearer "+accessToken)
		req.Header.Set("User-Agent", "reddmeitalpha/0.1")

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
			subreddits[c["subreddit"].(string)] = true
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
	subreddits := make(map[string]bool)
	after := ""
	for {
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Authorization", "bearer "+accessToken)
		req.Header.Set("User-Agent", "reddmeitalpha/0.1")

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
			subreddits[c["subreddit"].(string)] = true
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

func getRecommendations(subs []string, userPrompt string) string {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("Missing OPENAI_API_KEY in .env")
	}

	client := openai.NewClient(apiKey)
	prompt := "The user is active in the following subreddits:\n\n"
	for _, sub := range subs {
		prompt += sub + "\n"
	}
	prompt += "\n\nThey said: \"" + userPrompt + "\""
	prompt += "\n\nSuggest subreddits to JOIN and LEAVE in bullet points."

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)

	if err != nil {
		log.Fatalf("OpenAI API error: %v\n", err)
	}
	return resp.Choices[0].Message.Content
}
