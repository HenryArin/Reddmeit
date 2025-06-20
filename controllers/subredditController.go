package controllers

import "github.com/HenryArin/ReddmeitAlpha/models"

func CombineSubredditStats(subscribed, upvoted, commented map[string]bool) map[string]*models.SubredditStats {
	combined := make(map[string]*models.SubredditStats)

	for sub := range subscribed {
		if combined[sub] == nil {
			combined[sub] = &models.SubredditStats{Name: sub}
		}
		combined[sub].Subscribed = true
	}

	for sub := range upvoted {
		if combined[sub] == nil {
			combined[sub] = &models.SubredditStats{Name: sub}
		}
		combined[sub].Upvoted = true
	}

	for sub := range commented {
		if combined[sub] == nil {
			combined[sub] = &models.SubredditStats{Name: sub}
		}
		combined[sub].Commented = true
	}

	return combined
}

func FilterActiveSubreddits(combined map[string]*models.SubredditStats, threshold int) []string {
	var active []string
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
		if score >= threshold {
			active = append(active, "r/"+stat.Name)
		}
	}
	return active
}
