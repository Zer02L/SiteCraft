package data

import (
	"errors"
	"strings"
	"time"

	"sitecraft/internal/cache"
	"sitecraft/internal/models"

	"github.com/pocketbase/pocketbase/core"
)

// GetAllServicesAdmin retrieves all active/inactive services excluding trash.
func GetAllServicesAdmin(app core.App) []models.ServiceItem {
	records, err := app.FindRecordsByFilter("services", "", "order, id", 0, 0)
	if err != nil || len(records) == 0 {
		return nil
	}

	var services []models.ServiceItem
	for _, r := range records {
		if r.GetBool("is_deleted") {
			continue
		}
		services = append(services, RecordToServiceItem(r))
	}
	return services
}

// GetTrashServicesAdmin retrieves all soft-deleted services in trash.
func GetTrashServicesAdmin(app core.App) []models.ServiceItem {
	records, err := app.FindRecordsByFilter("services", "is_deleted = true", "order, id", 0, 0)
	if err != nil || len(records) == 0 {
		return nil
	}

	var services []models.ServiceItem
	for _, r := range records {
		services = append(services, RecordToServiceItem(r))
	}
	return services
}

// GetServiceByID retrieves a single service record by ID.
func GetServiceByID(app core.App, id string) (*models.ServiceItem, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("пустой идентификатор услуги")
	}
	r, err := app.FindRecordById("services", id)
	if err != nil || r == nil {
		return nil, errors.New("услуга не найдена")
	}
	item := RecordToServiceItem(r)
	return &item, nil
}

// SaveService creates or updates a service record and invalidates the in-memory cache.
func SaveService(app core.App, item models.ServiceItem) (*models.ServiceItem, error) {
	title := strings.TrimSpace(item.Title)
	if title == "" {
		return nil, errors.New("название услуги обязательно для заполнения")
	}

	servicesCol, err := app.FindCollectionByNameOrId("services")
	if err != nil {
		return nil, errors.New("коллекция services не найдена")
	}

	var rec *core.Record
	if item.ID != "" {
		rec, _ = app.FindRecordById("services", item.ID)
	}
	if rec == nil {
		rec = core.NewRecord(servicesCol)
	}

	rec.Set("title", title)
	rec.Set("description", strings.TrimSpace(item.Description))
	rec.Set("price", strings.TrimSpace(item.Price))
	rec.Set("order", item.Order)
	rec.Set("active", item.Active)
	rec.Set("is_deleted", false)

	if rec.Collection().Fields.GetByName("stock_count") != nil {
		rec.Set("stock_count", item.StockCount)
	}
	if rec.Collection().Fields.GetByName("is_preorder") != nil {
		rec.Set("is_preorder", item.IsPreorder)
	}

	if err := app.Save(rec); err != nil {
		return nil, err
	}

	cache.GetStore().InvalidateServices()
	saved := RecordToServiceItem(rec)
	return &saved, nil
}

// ToggleServiceActive toggles the active state of a service in 1 click and invalidates the cache.
func ToggleServiceActive(app core.App, id string) (bool, error) {
	if strings.TrimSpace(id) == "" {
		return false, errors.New("пустой идентификатор услуги")
	}
	rec, err := app.FindRecordById("services", id)
	if err != nil || rec == nil {
		return false, errors.New("услуга не найдена")
	}

	newStatus := !rec.GetBool("active")
	rec.Set("active", newStatus)
	if err := app.Save(rec); err != nil {
		return false, err
	}

	cache.GetStore().InvalidateServices()
	return newStatus, nil
}

// MoveServiceToTrash moves a service to trash (soft-delete, preventing accidental data loss).
func MoveServiceToTrash(app core.App, id string) error {
	if strings.TrimSpace(id) == "" {
		return errors.New("пустой идентификатор услуги")
	}
	rec, err := app.FindRecordById("services", id)
	if err != nil || rec == nil {
		return errors.New("услуга не найдена")
	}

	rec.Set("is_deleted", true)
	rec.Set("deleted_at", time.Now().Format("2006-01-02 15:04:05"))
	if err := app.Save(rec); err != nil {
		return err
	}

	cache.GetStore().InvalidateServices()
	return nil
}

// RestoreServiceFromTrash restores a service from trash back to active/hidden catalog.
func RestoreServiceFromTrash(app core.App, id string) error {
	if strings.TrimSpace(id) == "" {
		return errors.New("пустой идентификатор услуги")
	}
	rec, err := app.FindRecordById("services", id)
	if err != nil || rec == nil {
		return errors.New("услуга не найдена")
	}

	rec.Set("is_deleted", false)
	rec.Set("deleted_at", "")
	if err := app.Save(rec); err != nil {
		return err
	}

	cache.GetStore().InvalidateServices()
	return nil
}

// PermanentDeleteService permanently deletes a service record from database.
func PermanentDeleteService(app core.App, id string) error {
	if strings.TrimSpace(id) == "" {
		return errors.New("пустой идентификатор услуги")
	}
	rec, err := app.FindRecordById("services", id)
	if err != nil || rec == nil {
		return errors.New("услуга не найдена")
	}

	if err := app.Delete(rec); err != nil {
		return err
	}

	cache.GetStore().InvalidateServices()
	return nil
}

// EmptyTrashServices permanently removes all soft-deleted services from trash.
func EmptyTrashServices(app core.App) error {
	records, err := app.FindRecordsByFilter("services", "is_deleted = true", "", 0, 0)
	if err != nil {
		return err
	}
	for _, rec := range records {
		_ = app.Delete(rec)
	}

	cache.GetStore().InvalidateServices()
	return nil
}

// ReorderService shifts a service up or down in the catalog by swapping order values.
func ReorderService(app core.App, id string, direction string) error {
	if strings.TrimSpace(id) == "" {
		return errors.New("пустой идентификатор услуги")
	}

	records, err := app.FindRecordsByFilter("services", "", "order, id", 0, 0)
	if err != nil || len(records) == 0 {
		return errors.New("нет записей для сортировки")
	}

	// Filter out deleted
	var activeRecords []*core.Record
	for _, r := range records {
		if !r.GetBool("is_deleted") {
			activeRecords = append(activeRecords, r)
		}
	}

	targetIndex := -1
	for i, r := range activeRecords {
		if r.Id == id {
			targetIndex = i
			break
		}
	}

	if targetIndex == -1 {
		return errors.New("позиция не найдена")
	}

	swapIndex := -1
	if direction == "up" && targetIndex > 0 {
		swapIndex = targetIndex - 1
	} else if direction == "down" && targetIndex < len(activeRecords)-1 {
		swapIndex = targetIndex + 1
	}

	if swapIndex == -1 {
		// Already at the boundary
		return nil
	}

	currOrder := activeRecords[targetIndex].GetInt("order")
	swapOrder := activeRecords[swapIndex].GetInt("order")

	// If orders are identical, renumber sequentially
	if currOrder == swapOrder {
		currOrder = targetIndex + 1
		swapOrder = swapIndex + 1
	}

	activeRecords[targetIndex].Set("order", swapOrder)
	activeRecords[swapIndex].Set("order", currOrder)

	if err := app.Save(activeRecords[targetIndex]); err != nil {
		return err
	}
	if err := app.Save(activeRecords[swapIndex]); err != nil {
		return err
	}

	cache.GetStore().InvalidateServices()
	return nil
}

// RecordToServiceItem converts a PocketBase record into a domain ServiceItem model.
func RecordToServiceItem(r *core.Record) models.ServiceItem {
	stockCount := -1
	if r.Collection().Fields.GetByName("stock_count") != nil {
		stockCount = r.GetInt("stock_count")
	}
	isPreorder := false
	if r.Collection().Fields.GetByName("is_preorder") != nil {
		isPreorder = r.GetBool("is_preorder")
	}
	isDeleted := false
	if r.Collection().Fields.GetByName("is_deleted") != nil {
		isDeleted = r.GetBool("is_deleted")
	}
	deletedAt := ""
	if r.Collection().Fields.GetByName("deleted_at") != nil {
		deletedAt = r.GetString("deleted_at")
	}

	return models.ServiceItem{
		ID:          r.Id,
		Title:       r.GetString("title"),
		Description: r.GetString("description"),
		Price:       r.GetString("price"),
		Order:       r.GetInt("order"),
		Active:      r.GetBool("active"),
		StockCount:  stockCount,
		IsPreorder:  isPreorder,
		IsDeleted:   isDeleted,
		DeletedAt:   deletedAt,
	}
}
