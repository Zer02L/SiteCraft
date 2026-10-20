package models

// ReviewItem represents a customer review with rating and optional photos (Verified UGC).
type ReviewItem struct {
	ID           string   `json:"id"`
	AuthorName   string   `json:"author_name"`
	City         string   `json:"city,omitempty"`
	ServiceTitle string   `json:"service_title,omitempty"`
	Rating       int      `json:"rating"` // 1 to 5
	Comment      string   `json:"comment"`
	OrderID      string   `json:"order_id,omitempty"` // Verified purchase/service link
	OrderCode    string   `json:"order_code,omitempty"`
	Photos       []string `json:"photos,omitempty"`
	IsPublished  bool     `json:"is_published"`
	Source       string   `json:"source,omitempty"` // "track", "site"
	CreatedAt    string   `json:"created_at,omitempty"`
}

// ReviewAggregate calculates summary ratings for Schema.org AggregateRating.
type ReviewAggregate struct {
	TotalReviews int     `json:"total_reviews"`
	AverageScore float64 `json:"average_score"`
	BestRating   int     `json:"best_rating"`
	WorstRating  int     `json:"worst_rating"`
}
