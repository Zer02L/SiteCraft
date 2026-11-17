package data

import (
	"context"

	"sitecraft/internal/db"
	"sitecraft/internal/models"

	"github.com/google/uuid"
)

// GetLeads returns recent leads.
func (r *Repository) GetLeads(ctx context.Context) ([]models.Lead, error) {
	rows, err := r.queries.ListLeads(ctx)
	if err != nil {
		return nil, err
	}

	leads := make([]models.Lead, len(rows))
	for i, row := range rows {
		leads[i] = models.Lead{
			ID:          row.ID,
			Name:        row.Name,
			Phone:       row.Phone,
			Email:       row.Email,
			Source:      row.Source,
			BookingDate: row.BookingDate,
			BookingTime: row.BookingTime,
			ResourceID:  row.ResourceID,
			Status:      row.Status,
			Created:     row.CreatedAt,
		}
	}
	return leads, nil
}

// CreateLead inserts a new lead into database.
func (r *Repository) CreateLead(ctx context.Context, lead models.Lead) (models.Lead, error) {
	if lead.ID == "" {
		lead.ID = uuid.New().String()
	}
	if lead.Status == "" {
		lead.Status = "новая"
	}
	row, err := r.queries.CreateLead(ctx, db.CreateLeadParams{
		ID:          lead.ID,
		Name:        lead.Name,
		Phone:       lead.Phone,
		Email:       lead.Email,
		Source:      lead.Source,
		BookingDate: lead.BookingDate,
		BookingTime: lead.BookingTime,
		ResourceID:  lead.ResourceID,
		Status:      lead.Status,
		Payload:     "{}",
	})
	if err != nil {
		return lead, err
	}
	lead.Created = row.CreatedAt
	return lead, nil
}

// UpdateLeadStatus updates the status of a lead.
func (r *Repository) UpdateLeadStatus(ctx context.Context, id, status string) error {
	_, err := r.queries.UpdateLeadStatus(ctx, db.UpdateLeadStatusParams{
		ID:     id,
		Status: status,
	})
	return err
}
