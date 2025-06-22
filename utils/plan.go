package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/HenryArin/ReddmeitAlpha/models"
)

// PrintPlan prints the recommended subreddits with optional explanations.
func PrintPlan(plan models.RecommendationPlan) {
	sort.Strings(plan.ToAdd)
	sort.Strings(plan.ToRemove)

	if len(plan.ToAdd) > 0 {
		fmt.Println("To Add:")
		for _, sub := range plan.ToAdd {
			if explanation, ok := plan.Explanations[sub]; ok && explanation != "" {
				fmt.Printf(" + %s (%s)\n", sub, explanation)
			} else {
				fmt.Printf(" + %s\n", sub)
			}
		}
	}
	if len(plan.ToRemove) > 0 {
		fmt.Println("To Remove:")
		for _, sub := range plan.ToRemove {
			if explanation, ok := plan.Explanations[sub]; ok && explanation != "" {
				fmt.Printf(" - %s (%s)\n", sub, explanation)
			} else {
				fmt.Printf(" - %s\n", sub)
			}
		}
	}
	if len(plan.ToAdd) == 0 && len(plan.ToRemove) == 0 {
		fmt.Println("= No changes needed.")
	}
}

// SavePlanToFile saves the plan as a JSON file in the /logs directory.
func SavePlanToFile(plan models.RecommendationPlan, prompt string) {
	sort.Strings(plan.ToAdd)
	sort.Strings(plan.ToRemove)

	timestamp := time.Now().Format("2006-01-02")
	safePrompt := strings.ReplaceAll(prompt, " ", "_")
	safePrompt = strings.ReplaceAll(safePrompt, "/", "_")
	filename := fmt.Sprintf("logs/%s_%s.json", timestamp, safePrompt)

	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("❌ Failed to save plan: %v\n", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(plan); err != nil {
		fmt.Printf("❌ Failed to encode plan: %v\n", err)
		return
	}

	fmt.Printf("✅ Saved to: %s\n", filename)
}

// MergePlans combines two plans and merges explanations.
func MergePlans(a, b models.RecommendationPlan) models.RecommendationPlan {
	toAdd := map[string]bool{}
	toRemove := map[string]bool{}
	explanations := map[string]string{}

	for _, sub := range a.ToAdd {
		toAdd[sub] = true
	}
	for _, sub := range a.ToRemove {
		toRemove[sub] = true
	}
	for _, sub := range b.ToAdd {
		toAdd[sub] = true
	}
	for _, sub := range b.ToRemove {
		toRemove[sub] = true
	}

	for _, m := range []map[string]string{a.Explanations, b.Explanations} {
		for sub, reason := range m {
			explanations[sub] = reason
		}
	}

	// Avoid conflicts: a sub can't be in both lists
	for sub := range toAdd {
		if toRemove[sub] {
			delete(toAdd, sub)
			delete(toRemove, sub)
			delete(explanations, sub)
		}
	}

	return models.RecommendationPlan{
		ToAdd:        keys(toAdd),
		ToRemove:     keys(toRemove),
		Explanations: explanations,
	}
}

func keys(m map[string]bool) []string {
	out := []string{}
	for k := range m {
		out = append(out, k)
	}
	return out
}

// ParseSubredditPlan parses a GPT reply into a plan with explanations.
func ParseSubredditPlan(response string) models.RecommendationPlan {
	var toAdd, toRemove []string
	explanations := map[string]string{}

	lines := strings.Split(response, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "+ r/") || strings.HasPrefix(line, "+r/") {
			sub := extractSubreddit(line)
			if sub != "" {
				toAdd = append(toAdd, sub)
				explanations[sub] = extractExplanation(line)
			}
		} else if strings.HasPrefix(line, "- r/") || strings.HasPrefix(line, "-r/") {
			sub := extractSubreddit(line)
			if sub != "" {
				toRemove = append(toRemove, sub)
				explanations[sub] = extractExplanation(line)
			}
		}
	}

	return models.RecommendationPlan{
		ToAdd:        toAdd,
		ToRemove:     toRemove,
		Explanations: explanations,
	}
}

// extractSubreddit uses regex to find the r/subreddit pattern.
func extractSubreddit(line string) string {
	re := regexp.MustCompile(`r/[a-zA-Z0-9_]+`)
	return re.FindString(line)
}

// extractExplanation pulls any trailing explanation from the line.
func extractExplanation(line string) string {
	parts := strings.SplitN(line, " - ", 2)
	if len(parts) < 2 {
		parts = strings.SplitN(line, " – ", 2) // en-dash
	}
	if len(parts) == 2 {
		return strings.TrimSpace(parts[1])
	}
	return ""
}
