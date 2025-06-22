package models

// Recommendation is used internally to structure + / - / = suggestions from GPT
type Recommendation struct {
	Add    []string
	Remove []string
	Keep   []string
}

// RecommendationPlan represents the plan the assistant will execute and/or save
type RecommendationPlan struct {
	ToAdd    []string `json:"to_add"`
	ToRemove []string `json:"to_remove"`
	ViewOnly bool     `json:"view_only,omitempty"`
	Reply    string   `json:"reply,omitempty"`
	Explanations map[string]string `json:"explanations,omitempty"`
}
