package data

import (
	"context"

	"sitecraft/internal/models"
)

// GetReviews returns published reviews.
func (r *Repository) GetReviews(ctx context.Context) ([]models.ReviewItem, error) {
	rows, err := r.queries.ListPublishedReviews(ctx)
	if err != nil {
		return nil, err
	}

	reviews := make([]models.ReviewItem, len(rows))
	for i, row := range rows {
		var photos []string
		if row.Photo != "" {
			photos = []string{row.Photo}
		}
		reviews[i] = models.ReviewItem{
			ID:          row.ID,
			AuthorName:  row.AuthorName,
			Rating:      int(row.Rating),
			Comment:     row.Comment,
			OrderID:     row.OrderID,
			Photos:      photos,
			IsPublished: row.IsPublished == 1,
			Source:      row.Source,
			CreatedAt:   row.CreatedAt,
		}
	}
	return reviews, nil
}

func (r *Repository) ApproveReview(ctx context.Context, id string) error {
	_, err := r.queries.ApproveReview(ctx, id)
	return err
}

// ToggleReview toggles is_published between 1 and 0.
func (r *Repository) ToggleReview(ctx context.Context, id string) error {
	_, err := r.rawDB.ExecContext(ctx, "UPDATE reviews SET is_published = CASE WHEN is_published = 1 THEN 0 ELSE 1 END WHERE id = ?", id)
	return err
}

// DeleteReview deletes a review by ID.
func (r *Repository) DeleteReview(ctx context.Context, id string) error {
	_, err := r.rawDB.ExecContext(ctx, "DELETE FROM reviews WHERE id = ?", id)
	return err
}
