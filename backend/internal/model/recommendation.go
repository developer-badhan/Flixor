package model

/*
 * RecommendRequest is what the client sends.
 * "mode" can be "rule" | "ai" | "hybrid" (default: hybrid)
 * If no mode is sent, we run both engines and merge results.
*/
type RecommendRequest struct {
	Mode  string `json:"mode" binding:"omitempty,oneof=rule ai hybrid"`
	Limit int    `json:"limit" binding:"omitempty,min=1,max=50"`
}

/*
 * RecommendedMovie is a slim movie view used in recommendation responses.
 * We deliberately keep it light — only what the frontend needs to render a card.
 */
type RecommendedMovie struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Genre       []string `json:"genre"`
	Thumbnail   string   `json:"thumbnail"`
	Year        string   `json:"year"`
	ViewCount   int64    `json:"view_count"`
	Source      string   `json:"source"` 
	MatchReason string   `json:"match_reason"`
}

/*
 * RecommendResponse is the full API response.
 * Mode: which engine ran ("rule" | "ai" | "hybrid")
 * TotalResults: total movies returned (before pagination)
 * Movies: array of RecommendedMovie
 */
type RecommendResponse struct {
	Mode         string             `json:"mode"`
	TotalResults int                `json:"total_results"`
	Movies       []RecommendedMovie `json:"movies"`
}