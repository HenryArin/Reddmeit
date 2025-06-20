package controllers

import (
	"strings"

	"github.com/HenryArin/ReddmeitAlpha/models"
)

func ParseRecommendationOutput(output string) models.Recommendation {
	lines := strings.Split(output, "\n")
	seen := make(map[string]bool)

	rec := models.Recommendation{}

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "+ r/") {
			sub := strings.TrimSpace(strings.TrimPrefix(line, "+ "))
			if !seen[sub] {
				rec.Add = append(rec.Add, sub)
				seen[sub] = true
			}
		} else if strings.HasPrefix(line, "- r/") {
			sub := strings.TrimSpace(strings.TrimPrefix(line, "- "))
			if !seen[sub] {
				rec.Remove = append(rec.Remove, sub)
				seen[sub] = true
			}
		} else if strings.HasPrefix(line, "= r/") {
			sub := strings.TrimSpace(strings.TrimPrefix(line, "= "))
			if !seen[sub] {
				rec.Keep = append(rec.Keep, sub)
				seen[sub] = true
			}
		}
	}

	return rec
}
