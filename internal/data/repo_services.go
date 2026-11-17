package data

import (
	"context"

	"sitecraft/internal/db"
	"sitecraft/internal/models"

	"github.com/google/uuid"
)

// GetServices returns active services.
func (r *Repository) GetServices(ctx context.Context) ([]models.ServiceItem, error) {
	c := r.Cache()
	if cached, ok := c.GetServices(); ok {
		return cached, nil
	}

	rows, err := r.queries.ListActiveServices(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]models.ServiceItem, len(rows))
	for i, row := range rows {
		items[i] = models.ServiceItem{
			ID:          row.ID,
			Title:       row.Title,
			Description: row.Description,
			Price:       row.Price,
			Order:       int(row.OrderIndex),
			Active:      row.Active == 1,
			StockCount:  int(row.StockCount),
			IsPreorder:  row.IsPreorder == 1,
			IsDeleted:   row.IsDeleted == 1,
		}
	}
	c.SetServices(items)
	return items, nil
}

// GetServiceByID returns a single service by ID.
func (r *Repository) GetServiceByID(ctx context.Context, id string) (models.ServiceItem, error) {
	row, err := r.queries.GetServiceByID(ctx, id)
	if err != nil {
		return models.ServiceItem{}, err
	}
	return models.ServiceItem{
		ID:          row.ID,
		Title:       row.Title,
		Description: row.Description,
		Price:       row.Price,
		Order:       int(row.OrderIndex),
		Active:      row.Active == 1,
		StockCount:  int(row.StockCount),
		IsPreorder:  row.IsPreorder == 1,
		IsDeleted:   row.IsDeleted == 1,
	}, nil
}

// SaveService creates or updates a service and invalidates services cache.
func (r *Repository) SaveService(ctx context.Context, item models.ServiceItem) error {
	actInt := int64(0)
	if item.Active {
		actInt = 1
	}
	stock := int64(item.StockCount)
	if stock == 0 {
		stock = -1
	}
	preorder := int64(0)
	if item.IsPreorder {
		preorder = 1
	}

	var err error
	if item.ID != "" {
		_, err = r.queries.UpdateService(ctx, db.UpdateServiceParams{
			ID:          item.ID,
			Title:       item.Title,
			Description: item.Description,
			Price:       item.Price,
			OrderIndex:  int64(item.Order),
			Active:      actInt,
			StockCount:  stock,
			IsPreorder:  preorder,
		})
	} else {
		_, err = r.queries.CreateService(ctx, db.CreateServiceParams{
			ID:          uuid.New().String(),
			Title:       item.Title,
			Description: item.Description,
			Price:       item.Price,
			OrderIndex:  int64(item.Order),
			Active:      actInt,
			StockCount:  stock,
			IsPreorder:  preorder,
		})
	}
	if err != nil {
		return err
	}
	r.Cache().InvalidateServices()
	return nil
}

// ToggleService toggles the active state of a service and invalidates services cache.
func (r *Repository) ToggleService(ctx context.Context, id string) error {
	svc, err := r.queries.GetServiceByID(ctx, id)
	if err != nil {
		return err
	}
	newActive := int64(1)
	if svc.Active == 1 {
		newActive = 0
	}
	_, err = r.queries.UpdateService(ctx, db.UpdateServiceParams{
		ID:          svc.ID,
		Title:       svc.Title,
		Description: svc.Description,
		Price:       svc.Price,
		OrderIndex:  svc.OrderIndex,
		Active:      newActive,
		StockCount:  svc.StockCount,
		IsPreorder:  svc.IsPreorder,
	})
	if err != nil {
		return err
	}
	r.Cache().InvalidateServices()
	return nil
}

// SoftDeleteService marks a service as deleted and invalidates services cache.
func (r *Repository) SoftDeleteService(ctx context.Context, id string) error {
	if err := r.queries.SoftDeleteService(ctx, id); err != nil {
		return err
	}
	r.Cache().InvalidateServices()
	return nil
}
