package utils

import (
	"fmt"
	"strings"
)

// Plan holds lists of subreddit actions
type Plan struct {
	ToAdd    []string
	ToRemove []string
	ToKeep   []string
}

// ParseSubredditPlan parses GPT response into a Plan
func ParseSubredditPlan(response string) Plan {
	lines := strings.Split(response, "\n")
	var plan Plan

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) < 2 {
			continue
		}

		prefix := line[0]
		subreddit := strings.TrimSpace(line[1:])

		switch prefix {
		case '+':
			plan.ToAdd = append(plan.ToAdd, subreddit)
		case '-':
			plan.ToRemove = append(plan.ToRemove, subreddit)
		case '=':
			plan.ToKeep = append(plan.ToKeep, subreddit)
		}
	}

	return plan
}

// PrintPlan displays the actionable adds and removes
func PrintPlan(plan Plan) {
	if len(plan.ToAdd) > 0 {
		fmt.Println("To Add:")
		for _, sub := range plan.ToAdd {
			fmt.Printf(" + %s\n", sub)
		}
	}
	if len(plan.ToRemove) > 0 {
		fmt.Println("To Remove:")
		for _, sub := range plan.ToRemove {
			fmt.Printf(" - %s\n", sub)
		}
	}
}
