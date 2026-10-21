package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"sitecraft/internal/data"
	"sitecraft/internal/models"
	"sitecraft/ui/pages"

	"github.com/microcosm-cc/bluemonday"
	"github.com/pocketbase/pocketbase/core"
)

// HandleTabServices renders services tab fragment for HTMX.
func (h *ClientAdminHandler) HandleTabServices(e *core.RequestEvent) error {
	if !h.isAuthorized(e) {
		return e.Redirect(http.StatusFound, "/client/login")
	}

	filter := strings.TrimSpace(e.Request.URL.Query().Get("filter"))
	services := data.GetAllServicesAdmin(h.app)
	trashServices := data.GetTrashServicesAdmin(h.app)

	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return pages.ClientServicesTab(services, trashServices, filter).Render(e.Request.Context(), e.Response)
}

// HandleNewServiceForm renders the form to create a new service.
func (h *ClientAdminHandler) HandleNewServiceForm(e *core.RequestEvent) error {
	if !h.isAuthorized(e) {
		return e.Redirect(http.StatusFound, "/client/login")
	}

	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return pages.ClientServiceForm(models.ServiceItem{
		Active:     true,
		StockCount: -1,
		Order:      len(data.GetAllServicesAdmin(h.app)) + 1,
	}, true).Render(e.Request.Context(), e.Response)
}

// HandleEditServiceForm renders the form to edit an existing service.
func (h *ClientAdminHandler) HandleEditServiceForm(e *core.RequestEvent) error {
	if !h.isAuthorized(e) {
		return e.Redirect(http.StatusFound, "/client/login")
	}

	id := e.Request.PathValue("id")
	item, err := data.GetServiceByID(h.app, id)
	if err != nil || item == nil {
		return e.String(http.StatusNotFound, "Услуга не найдена")
	}

	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return pages.ClientServiceForm(*item, false).Render(e.Request.Context(), e.Response)
}

// HandleSaveService creates or updates a service.
func (h *ClientAdminHandler) HandleSaveService(e *core.RequestEvent) error {
	if !h.isAuthorized(e) {
		return e.String(http.StatusUnauthorized, "Неавторизован")
	}

	policy := bluemonday.UGCPolicy()

	id := strings.TrimSpace(e.Request.FormValue("id"))
	title := strings.TrimSpace(policy.Sanitize(e.Request.FormValue("title")))
	description := strings.TrimSpace(policy.Sanitize(e.Request.FormValue("description")))
	price := strings.TrimSpace(policy.Sanitize(e.Request.FormValue("price")))
	active := e.Request.FormValue("active") == "true"
	isPreorder := e.Request.FormValue("is_preorder") == "true"

	order, _ := strconv.Atoi(strings.TrimSpace(e.Request.FormValue("order")))

	stockCountStr := strings.TrimSpace(e.Request.FormValue("stock_count"))
	stockCount := -1
	if stockCountStr != "" {
		if parsedStock, err := strconv.Atoi(stockCountStr); err == nil {
			stockCount = parsedStock
		}
	}

	if title == "" {
		return e.String(http.StatusUnprocessableEntity, "Название услуги обязательно для заполнения")
	}

	item := models.ServiceItem{
		ID:          id,
		Title:       title,
		Description: description,
		Price:       price,
		Order:       order,
		Active:      active,
		StockCount:  stockCount,
		IsPreorder:  isPreorder,
	}

	if _, err := data.SaveService(h.app, item); err != nil {
		return e.String(http.StatusInternalServerError, fmt.Sprintf("Ошибка сохранения: %v", err))
	}

	services := data.GetAllServicesAdmin(h.app)
	trashServices := data.GetTrashServicesAdmin(h.app)
	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return pages.ClientServicesTab(services, trashServices, "all").Render(e.Request.Context(), e.Response)
}

// HandleToggleServiceActive toggles active status in 1 click via HTMX.
func (h *ClientAdminHandler) HandleToggleServiceActive(e *core.RequestEvent) error {
	if !h.isAuthorized(e) {
		return e.String(http.StatusUnauthorized, "Неавторизован")
	}

	id := e.Request.PathValue("id")
	if _, err := data.ToggleServiceActive(h.app, id); err != nil {
		return e.String(http.StatusBadRequest, err.Error())
	}

	services := data.GetAllServicesAdmin(h.app)
	trashServices := data.GetTrashServicesAdmin(h.app)
	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return pages.ClientServicesTab(services, trashServices, "").Render(e.Request.Context(), e.Response)
}

// HandleTrashService moves service to trash (safe soft-delete).
func (h *ClientAdminHandler) HandleTrashService(e *core.RequestEvent) error {
	if !h.isAuthorized(e) {
		return e.String(http.StatusUnauthorized, "Неавторизован")
	}

	id := e.Request.PathValue("id")
	if err := data.MoveServiceToTrash(h.app, id); err != nil {
		return e.String(http.StatusBadRequest, err.Error())
	}

	services := data.GetAllServicesAdmin(h.app)
	trashServices := data.GetTrashServicesAdmin(h.app)
	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return pages.ClientServicesTab(services, trashServices, "").Render(e.Request.Context(), e.Response)
}

// HandleRestoreService restores service from trash back to active/hidden catalog.
func (h *ClientAdminHandler) HandleRestoreService(e *core.RequestEvent) error {
	if !h.isAuthorized(e) {
		return e.String(http.StatusUnauthorized, "Неавторизован")
	}

	id := e.Request.PathValue("id")
	if err := data.RestoreServiceFromTrash(h.app, id); err != nil {
		return e.String(http.StatusBadRequest, err.Error())
	}

	services := data.GetAllServicesAdmin(h.app)
	trashServices := data.GetTrashServicesAdmin(h.app)
	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return pages.ClientServicesTab(services, trashServices, "trash").Render(e.Request.Context(), e.Response)
}

// HandlePermanentDeleteService permanently removes service from database.
func (h *ClientAdminHandler) HandlePermanentDeleteService(e *core.RequestEvent) error {
	if !h.isAuthorized(e) {
		return e.String(http.StatusUnauthorized, "Неавторизован")
	}

	id := e.Request.PathValue("id")
	if err := data.PermanentDeleteService(h.app, id); err != nil {
		return e.String(http.StatusBadRequest, err.Error())
	}

	services := data.GetAllServicesAdmin(h.app)
	trashServices := data.GetTrashServicesAdmin(h.app)
	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return pages.ClientServicesTab(services, trashServices, "trash").Render(e.Request.Context(), e.Response)
}

// HandleEmptyTrashServices permanently removes all soft-deleted services.
func (h *ClientAdminHandler) HandleEmptyTrashServices(e *core.RequestEvent) error {
	if !h.isAuthorized(e) {
		return e.String(http.StatusUnauthorized, "Неавторизован")
	}

	if err := data.EmptyTrashServices(h.app); err != nil {
		return e.String(http.StatusBadRequest, err.Error())
	}

	services := data.GetAllServicesAdmin(h.app)
	trashServices := data.GetTrashServicesAdmin(h.app)
	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return pages.ClientServicesTab(services, trashServices, "trash").Render(e.Request.Context(), e.Response)
}

// HandleReorderService shifts a service up or down in the catalog in 1 click.
func (h *ClientAdminHandler) HandleReorderService(e *core.RequestEvent) error {
	if !h.isAuthorized(e) {
		return e.String(http.StatusUnauthorized, "Неавторизован")
	}

	id := e.Request.PathValue("id")
	dir := strings.ToLower(strings.TrimSpace(e.Request.URL.Query().Get("dir")))
	if dir != "up" && dir != "down" {
		return e.String(http.StatusBadRequest, "Неверное направление сортировки (ожидается up или down)")
	}

	if err := data.ReorderService(h.app, id, dir); err != nil {
		return e.String(http.StatusBadRequest, err.Error())
	}

	services := data.GetAllServicesAdmin(h.app)
	trashServices := data.GetTrashServicesAdmin(h.app)
	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return pages.ClientServicesTab(services, trashServices, "").Render(e.Request.Context(), e.Response)
}
