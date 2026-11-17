package data

import (
	"context"
	"encoding/json"
	"time"

	"sitecraft/internal/models"
)

// GetResources returns active resources / specialists.
func (r *Repository) GetResources(ctx context.Context) ([]models.ResourceItem, error) {
	rows, err := r.queries.ListActiveResources(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]models.ResourceItem, len(rows))
	for i, row := range rows {
		items[i] = models.ResourceItem{
			ID:     row.ID,
			Name:   row.Name,
			Role:   row.Role,
			Avatar: row.Avatar,
			Active: row.Active == 1,
			Order:  int(row.OrderIndex),
		}
	}
	return items, nil
}

// GetBranches returns active company branches.
func (r *Repository) GetBranches(ctx context.Context) ([]models.BranchItem, error) {
	rows, err := r.queries.ListActiveBranches(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]models.BranchItem, len(rows))
	for i, row := range rows {
		items[i] = models.BranchItem{
			ID:        row.ID,
			Name:      row.Name,
			City:      row.City,
			Address:   row.Address,
			Phone:     row.Phone,
			WorkHours: row.WorkHours,
			Lat:       row.Lat,
			Lng:       row.Lng,
			Metro:     row.Metro,
			IsMain:    row.IsMain == 1,
			Active:    row.Active == 1,
		}
	}
	return items, nil
}

// GetVacancies returns active vacancies.
func (r *Repository) GetVacancies(ctx context.Context) ([]models.VacancyItem, error) {
	rows, err := r.queries.ListActiveVacancies(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]models.VacancyItem, len(rows))
	for i, row := range rows {
		var resp, req, ben []string
		_ = json.Unmarshal([]byte(row.Responsibilities), &resp)
		_ = json.Unmarshal([]byte(row.Requirements), &req)
		_ = json.Unmarshal([]byte(row.Benefits), &ben)

		items[i] = models.VacancyItem{
			ID:               row.ID,
			Title:            row.Title,
			Department:       row.Department,
			Salary:           row.Salary,
			Experience:       row.Experience,
			Schedule:         row.Schedule,
			Responsibilities: resp,
			Requirements:     req,
			Benefits:         ben,
			Active:           row.Active == 1,
			Order:            int(row.OrderIndex),
		}
	}
	return items, nil
}

// GetBookedSlots returns list of booked time slots today.
func (r *Repository) GetBookedSlots(ctx context.Context, resourceID ...string) ([]string, error) {
	today := time.Now().Format("2006-01-02")
	leads, err := r.queries.ListLeads(ctx)
	if err != nil {
		return nil, err
	}

	var resID string
	if len(resourceID) > 0 {
		resID = resourceID[0]
	}

	var slots []string
	for _, l := range leads {
		if l.BookingDate == today && l.BookingTime != "" && l.Status != "отменена" {
			if resID == "" || l.ResourceID == resID {
				slots = append(slots, l.BookingTime)
			}
		}
	}
	return slots, nil
}
