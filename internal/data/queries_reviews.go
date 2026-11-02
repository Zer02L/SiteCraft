package data

import (
	"fmt"

	"sitecraft/internal/models"

	"ganigeorgiev.com/pocketbase/pocketbase/core"
)

// GetReviews retrieves published customer reviews for public display
func GetReviews(app core.App) []models.ReviewItem {
	records, err := app.FindRecordsByFilter("reviews", "is_published = true", "-created", 0, 0)
	if err != nil || len(records) == 0 {
		return nil
	}

	var items []models.ReviewItem
	for _, r := range records {
		items = append(items, models.ReviewItem{
			ID:          r.Id,
			AuthorName:  r.GetString("author_name"),
			Rating:      r.GetInt("rating"),
			Comment:     r.GetString("comment"),
			OrderID:     r.GetString("order_id"),
			IsPublished: r.GetBool("is_published"),
			Source:      r.GetString("source"),
			CreatedAt:   r.GetString("created"),
		})
	}
	return items
}

// GetReviewAggregate computes aggregate metrics for Schema.org JSON-LD
func GetReviewAggregate(reviews []models.ReviewItem) models.ReviewAggregate {
	if len(reviews) == 0 {
		return models.ReviewAggregate{
			TotalReviews: 0,
			AverageScore: 5.0,
			BestRating:   5,
			WorstRating:  1,
		}
	}

	total := len(reviews)
	sum := 0
	for _, r := range reviews {
		sum += r.Rating
	}
	avg := float64(sum) / float64(total)

	return models.ReviewAggregate{
		TotalReviews: total,
		AverageScore: avg,
		BestRating:   5,
		WorstRating:  1,
	}
}

// CreateReview saves a customer review with is_published = false (pending moderation)
func CreateReview(app core.App, review models.ReviewItem) (*core.Record, error) {
	col, err := app.FindCollectionByNameOrId("reviews")
	if err != nil {
		return nil, err
	}

	rec := core.NewRecord(col)
	rec.Set("author_name", review.AuthorName)
	rec.Set("rating", review.Rating)
	rec.Set("comment", htmlSanitizer.Sanitize(review.Comment))
	rec.Set("order_id", review.OrderID)
	rec.Set("is_published", review.IsPublished)
	rec.Set("source", review.Source)

	if err := app.Save(rec); err != nil {
		return nil, err
	}
	return rec, nil
}

// GetAllReviewsAdmin retrieves all reviews for client admin moderation
func GetAllReviewsAdmin(app core.App) []models.ReviewItem {
	records, err := app.FindRecordsByFilter("reviews", "", "-created", 0, 0)
	if err != nil || len(records) == 0 {
		return nil
	}

	var items []models.ReviewItem
	for _, r := range records {
		items = append(items, models.ReviewItem{
			ID:          r.Id,
			AuthorName:  r.GetString("author_name"),
			Rating:      r.GetInt("rating"),
			Comment:     r.GetString("comment"),
			OrderID:     r.GetString("order_id"),
			IsPublished: r.GetBool("is_published"),
			Source:      r.GetString("source"),
			CreatedAt:   r.GetString("created"),
		})
	}
	return items
}

// ToggleReviewPublish toggles the publication status of a review
func ToggleReviewPublish(app core.App, id string) error {
	rec, err := app.FindRecordById("reviews", id)
	if err != nil || rec == nil {
		return fmt.Errorf("отзыв не найден")
	}
	current := rec.GetBool("is_published")
	rec.Set("is_published", !current)
	return app.Save(rec)
}

// DeleteReview permanently deletes a review
func DeleteReview(app core.App, id string) error {
	rec, err := app.FindRecordById("reviews", id)
	if err != nil || rec == nil {
		return fmt.Errorf("отзыв не найден")
	}
	return app.Delete(rec)
}
