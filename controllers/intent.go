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
	RemoveMode         bool
	RawFeedback        string
}

// ParseConversationIntent examines the user input to determine their intent.
func ParseConversationIntent(input string) Intent {
	lc := strings.ToLower(input)
	intent := Intent{RawFeedback: input}

	// Regenerate additions
	if strings.Contains(lc, "give me") && strings.Contains(lc, "more") && strings.Contains(lc, "subs") {
		intent.RegenerateAdds = true
	}

	// Regenerate removals
	if strings.Contains(lc, "remove more") || strings.Contains(lc, "prune more") {
		intent.RegenerateRemoves = true
	}

	// Clear removal suggestions
	if strings.Contains(lc, "keep all") || strings.Contains(lc, "clear removes") {
		intent.ClearRemoves = true
	}

	// New topic prompt
	if strings.Contains(lc, "suggest more") && strings.Contains(lc, "subs") {
		intent.NewTopicPrompt = input
	}

	// Show current subscriptions
	if containsAny(lc, []string{
		"show subs", "show my subs", "show me my subs", "show subreddits",
		"show my subreddits", "show me my subreddits", "list my subs",
		"list my subreddits", "what am i subscribed to", "what communities",
		"what subreddits", "which ones do i follow", "subs i use",
	}) {
		intent.ShowSubList = true
	}

	// General fuzzy removal intent (we let GPT handle the category)
	if containsAny(lc, []string{
		"remove", "unsubscribe", "prune", "get rid", "delete", "i don't want",
	}) {
		intent.RemoveMode = true
		intent.NewTopicPrompt = input // Forward full phrase to GPT
		intent.RegenerateRemoves = true
	}

	return intent
}

func containsAny(input string, keywords []string) bool {
	for _, k := range keywords {
		if strings.Contains(input, k) {
			return true
		}
	}
	return false
}
