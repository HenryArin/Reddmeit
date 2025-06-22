package controllers

import (
	"strings"
)

type Intent struct {
	RegenerateAdds     bool
	RegenerateRemoves  bool
	ClearRemoves       bool
	NewTopicPrompt     string
	ShowSubList        bool
	RawFeedback        string
}

// ParseConversationIntent examines the user input to determine their intent.
func ParseConversationIntent(input string) Intent {
	lc := strings.ToLower(input)
	intent := Intent{RawFeedback: input}

	// Regenerate additions: e.g. "give me more subs"
	if strings.Contains(lc, "give me") && strings.Contains(lc, "more") && strings.Contains(lc, "subs") {
		intent.RegenerateAdds = true
	}

	// Regenerate removals: e.g. "remove more", "prune more"
	if strings.Contains(lc, "remove more") || strings.Contains(lc, "prune more") {
		intent.RegenerateRemoves = true
	}

	// Clear all removal suggestions: "keep all", "clear removes"
	if strings.Contains(lc, "keep all") || strings.Contains(lc, "clear removes") {
		intent.ClearRemoves = true
	}

	// New topic prompt: "suggest more subs"
	if strings.Contains(lc, "suggest more") && strings.Contains(lc, "subs") {
		intent.NewTopicPrompt = input
	}

	// Show current subscriptions: multiple phrasings
	if strings.Contains(lc, "show subs") ||
		strings.Contains(lc, "show my subs") ||
		strings.Contains(lc, "show me my subs") ||
		strings.Contains(lc, "show subreddits") ||
		strings.Contains(lc, "show my subreddits") ||
		strings.Contains(lc, "show me my subreddits") ||
		strings.Contains(lc, "list my subs") ||
		strings.Contains(lc, "list my subreddits") ||
		strings.Contains(lc, "what am i subscribed to") ||
		strings.Contains(lc, "what communities") ||
		strings.Contains(lc, "what subreddits") ||
		strings.Contains(lc, "which ones do i follow") ||
		strings.Contains(lc, "subs i use") {
		intent.ShowSubList = true
	}

	return intent
}
