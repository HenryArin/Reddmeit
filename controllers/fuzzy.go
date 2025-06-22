package controllers

import "strings"

// MatchesFuzzyRemove checks if the input expresses an intent to remove something.
func MatchesFuzzyRemove(input string) bool {
	lc := strings.ToLower(input)
	return strings.Contains(lc, "remove") ||
		strings.Contains(lc, "delete") ||
		strings.Contains(lc, "get rid of") ||
		strings.Contains(lc, "unsub") ||
		strings.Contains(lc, "leave") ||
		strings.Contains(lc, "drop") ||
		strings.Contains(lc, "no more")
}
