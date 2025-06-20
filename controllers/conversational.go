package controllers

import (
	"strings"
)

type Intent struct {
	RegenerateAdds     bool
	RegenerateRemoves  bool
	ClearRemoves       bool
	NewTopicPrompt     string
	RawFeedback        string
}

// Parses conversational feedback and determines what to do
func ParseConversationIntent(input string) Intent {
	input = strings.ToLower(input)
	intent := Intent{RawFeedback: input}

	if strings.Contains(input, "give me") && strings.Contains(input, "more") && strings.Contains(input, "subs") {
		intent.RegenerateAdds = true
	}

	if strings.Contains(input, "remove more") || strings.Contains(input, "prune more") {
		intent.RegenerateRemoves = true
	}

	if strings.Contains(input, "keep all") || strings.Contains(input, "clear removes") {
		intent.ClearRemoves = true
	}

	if strings.Contains(input, "suggest more") && strings.Contains(input, "subs") {
		intent.NewTopicPrompt = input // Let the AI reprocess this full line
	}

	return intent
}
