package data

import (
	"strings"

	"sitecraft/internal/models"

	"ganigeorgiev.com/pocketbase/dbx"
	"ganigeorgiev.com/pocketbase/pocketbase/core"
)

// GetBookedSlots returns a map of date -> slice of booked times for active leads (status != 'отменена')
// Optionally filters by resource_id (specialist or workstation).
func GetBookedSlots(app core.App, resourceID ...string) map[string][]string {
	result := make(map[string][]string)

	filter := "status != 'отменена'"
	params := dbx.Params{}
	if len(resourceID) > 0 && resourceID[0] != "" {
		filter += " && resource_id = {:resID}"
		params["resID"] = resourceID[0]
	}

	records, err := app.FindRecordsByFilter("leads", filter, "booking_date,booking_time", 0, 0, params)
	if err != nil {
		return result
	}

	for _, r := range records {
		date := r.GetString("booking_date")
		time := r.GetString("booking_time")
		if date != "" && time != "" {
			result[date] = append(result[date], time)
		}
	}
	return result
}

// GetResources retrieves active specialists, workstations or cabinets sorted by order
func GetResources(app core.App) []models.ResourceItem {
	records, err := app.FindRecordsByFilter("resources", "active = true", "order", 0, 0)
	if err != nil || len(records) == 0 {
		return nil
	}

	var items []models.ResourceItem
	for _, r := range records {
		var services []string
		if sStr := r.GetString("service_ids"); sStr != "" {
			parts := strings.Split(sStr, ",")
			for _, p := range parts {
				if trimmed := strings.TrimSpace(p); trimmed != "" {
					services = append(services, trimmed)
				}
			}
		}

		items = append(items, models.ResourceItem{
			ID:       r.Id,
			Name:     r.GetString("name"),
			Role:     r.GetString("role"),
			Avatar:   r.GetString("avatar"),
			Active:   r.GetBool("active"),
			Order:    r.GetInt("order"),
			Services: services,
		})
	}
	return items
}
