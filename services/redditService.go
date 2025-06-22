package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func FetchSubscribedSubreddits(accessToken string) map[string]bool {
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

		var parsed struct {
			Data struct {
				Children []struct {
					Data struct {
						DisplayName string `json:"display_name"`
					} `json:"data"`
				} `json:"children"`
				After string `json:"after"`
			} `json:"data"`
		}
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

func FetchUserActivity(username, accessToken, activityType string) map[string]bool {
	client := &http.Client{}
	subreddits := make(map[string]bool)
	after := ""
	url := fmt.Sprintf("https://oauth.reddit.com/user/%s/%s?limit=100", username, activityType)
	headers := map[string]string{
		"Authorization": "bearer " + accessToken,
		"User-Agent":    "reddmeitalpha/0.1",
	}

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

		if resp.StatusCode != 200 {
			fmt.Printf("Reddit API error (%s): %d\n", activityType, resp.StatusCode)
			fmt.Println("Response body:", string(body))
			break
		}

		var parsed map[string]interface{}
		if err := json.Unmarshal(body, &parsed); err != nil {
			fmt.Println("Failed to parse JSON:", err)
			fmt.Println("Raw response:", string(body))
			break
		}

		rawData, ok := parsed["data"]
		if !ok {
			fmt.Println("Missing 'data' in Reddit response")
			fmt.Println("Raw response:", string(body))
			break
		}

		data, ok := rawData.(map[string]interface{})
		if !ok {
			fmt.Printf("'data' is not a map[string]interface{} (got %T)\n", rawData)
			fmt.Println("Raw response:", string(body))
			break
		}

		children, ok := data["children"].([]interface{})
		if !ok {
			fmt.Println("Missing or invalid 'children' array in data")
			break
		}

		for _, child := range children {
			childMap, ok := child.(map[string]interface{})
			if !ok {
				continue
			}
			childData, ok := childMap["data"].(map[string]interface{})
			if !ok {
				continue
			}
			subreddit, ok := childData["subreddit"].(string)
			if ok {
				subreddits[subreddit] = true
			}
		}

		afterRaw, ok := data["after"]
		afterStr, isStr := afterRaw.(string)
		if !ok || !isStr || afterStr == "" {
			break
		}
		after = afterStr
		url = fmt.Sprintf("https://oauth.reddit.com/user/%s/%s?limit=100&after=%s", username, activityType, after)
	}

	return subreddits
}

func FetchUpvotedSubreddits(username, accessToken string) map[string]bool {
	return FetchUserActivity(username, accessToken, "upvoted")
}

func FetchCommentedSubreddits(username, accessToken string) map[string]bool {
	return FetchUserActivity(username, accessToken, "comments")
}
