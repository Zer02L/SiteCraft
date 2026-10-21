package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"sitecraft/internal/data"
	"sitecraft/internal/models"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func TestClientAdminServicesWorkflow(t *testing.T) {
	testDir := t.TempDir()
	app := pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDataDir: testDir,
	})
	defer app.ResetBootstrapState()

	if err := app.Bootstrap(); err != nil {
		t.Fatalf("Failed to bootstrap PB: %v", err)
	}
	if err := data.InitSchema(app); err != nil {
		t.Fatalf("Failed to init schema: %v", err)
	}

	handler := NewClientAdminHandler(app)

	// 1. Неавторизованный запрос к вкладке услуг -> редирект на /client/login
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/client/tab/services", nil)
	re := &core.RequestEvent{App: app}
	re.Request = req
	re.Response = rec

	if err := handler.HandleTabServices(re); err != nil {
		t.Fatalf("HandleTabServices error: %v", err)
	}
	if rec.Code != http.StatusFound || rec.Header().Get("Location") != "/client/login" {
		t.Fatalf("Expected 302 redirect to /client/login, got %d (loc: %s)", rec.Code, rec.Header().Get("Location"))
	}

	// 2. Создаем сессию и авторизуемся
	token, err := handler.sessionStore.CreateSession()
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
	authCookie := &http.Cookie{
		Name:  ClientAuthCookieName,
		Value: token,
		Path:  "/client",
	}

	// 3. Авторизованный запрос к вкладке услуг -> 200 OK
	recAuth := httptest.NewRecorder()
	reqAuth := httptest.NewRequest(http.MethodGet, "/client/tab/services", nil)
	reqAuth.AddCookie(authCookie)
	reAuth := &core.RequestEvent{App: app}
	reAuth.Request = reqAuth
	reAuth.Response = recAuth

	if err := handler.HandleTabServices(reAuth); err != nil {
		t.Fatalf("HandleTabServices error: %v", err)
	}
	if recAuth.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", recAuth.Code)
	}
	bodyTab := recAuth.Body.String()
	if !strings.Contains(bodyTab, "Каталог услуг и прайс-лист") {
		t.Errorf("Expected tab to contain catalog title, got: %s", bodyTab)
	}

	// 4. Форма создания новой услуги
	recNew := httptest.NewRecorder()
	reqNew := httptest.NewRequest(http.MethodGet, "/client/services/new", nil)
	reqNew.AddCookie(authCookie)
	reNew := &core.RequestEvent{App: app}
	reNew.Request = reqNew
	reNew.Response = recNew

	if err := handler.HandleNewServiceForm(reNew); err != nil {
		t.Fatalf("HandleNewServiceForm error: %v", err)
	}
	if recNew.Code != http.StatusOK || !strings.Contains(recNew.Body.String(), "Добавить услугу или товар") {
		t.Fatalf("Expected 200 with new service form, got %d", recNew.Code)
	}

	// 5. Валидация: попытка сохранить без названия -> 422
	recInvalid := httptest.NewRecorder()
	reqInvalid := httptest.NewRequest(http.MethodPost, "/client/services/save", strings.NewReader("price=1000"))
	reqInvalid.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqInvalid.AddCookie(authCookie)
	reInvalid := &core.RequestEvent{App: app}
	reInvalid.Request = reqInvalid
	reInvalid.Response = recInvalid

	if err := handler.HandleSaveService(reInvalid); err != nil {
		t.Fatalf("HandleSaveService error: %v", err)
	}
	if recInvalid.Code != http.StatusUnprocessableEntity {
		t.Fatalf("Expected 422 Unprocessable Entity, got %d", recInvalid.Code)
	}

	// 6. Создание двух услуг для проверки сортировки стрелками
	formData1 := url.Values{
		"title":       {"Премиальная мойка <script>alert('xss')</script>"},
		"description": {"Комплексный уход за кузовом"},
		"price":       {"3 500 ₽"},
		"order":       {"1"},
		"active":      {"true"},
		"stock_count": {"10"},
		"is_preorder": {"false"},
	}
	recSave1 := httptest.NewRecorder()
	reqSave1 := httptest.NewRequest(http.MethodPost, "/client/services/save", strings.NewReader(formData1.Encode()))
	reqSave1.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqSave1.AddCookie(authCookie)
	reSave1 := &core.RequestEvent{App: app}
	reSave1.Request = reqSave1
	reSave1.Response = recSave1

	if err := handler.HandleSaveService(reSave1); err != nil {
		t.Fatalf("HandleSaveService error: %v", err)
	}
	if recSave1.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK after save, got %d", recSave1.Code)
	}

	formData2 := url.Values{
		"title":       {"Полировка фар"},
		"description": {"Абразивное восстановление оптики"},
		"price":       {"1 800 ₽"},
		"order":       {"2"},
		"active":      {"true"},
		"stock_count": {"-1"},
		"is_preorder": {"false"},
	}
	recSave2 := httptest.NewRecorder()
	reqSave2 := httptest.NewRequest(http.MethodPost, "/client/services/save", strings.NewReader(formData2.Encode()))
	reqSave2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqSave2.AddCookie(authCookie)
	reSave2 := &core.RequestEvent{App: app}
	reSave2.Request = reqSave2
	reSave2.Response = recSave2

	if err := handler.HandleSaveService(reSave2); err != nil {
		t.Fatalf("HandleSaveService error: %v", err)
	}

	// Проверяем сохраненные услуги
	services := data.GetAllServicesAdmin(app)
	var created1, created2 *models.ServiceItem
	for i := range services {
		if strings.Contains(services[i].Title, "Премиальная мойка") {
			created1 = &services[i]
		}
		if services[i].Title == "Полировка фар" {
			created2 = &services[i]
		}
	}
	if created1 == nil || created2 == nil {
		t.Fatalf("Created services not found in database")
	}
	if strings.Contains(created1.Title, "<script>") {
		t.Fatalf("XSS was not stripped from title: %s", created1.Title)
	}

	// 7. Сортировка стрелкой (Reorder down): меняем порядок
	recReorder := httptest.NewRecorder()
	reqReorder := httptest.NewRequest(http.MethodPost, "/client/services/"+created1.ID+"/reorder?dir=down", nil)
	reqReorder.SetPathValue("id", created1.ID)
	reqReorder.AddCookie(authCookie)
	reReorder := &core.RequestEvent{App: app}
	reReorder.Request = reqReorder
	reReorder.Response = recReorder

	if err := handler.HandleReorderService(reReorder); err != nil {
		t.Fatalf("HandleReorderService error: %v", err)
	}
	if recReorder.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK after reorder, got %d", recReorder.Code)
	}

	// 8. Безопасное удаление в корзину (Soft Delete)
	recTrash := httptest.NewRecorder()
	reqTrash := httptest.NewRequest(http.MethodPost, "/client/services/"+created1.ID+"/trash", nil)
	reqTrash.SetPathValue("id", created1.ID)
	reqTrash.AddCookie(authCookie)
	reTrash := &core.RequestEvent{App: app}
	reTrash.Request = reqTrash
	reTrash.Response = recTrash

	if err := handler.HandleTrashService(reTrash); err != nil {
		t.Fatalf("HandleTrashService error: %v", err)
	}
	if recTrash.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK after trash, got %d", recTrash.Code)
	}

	// Проверяем: услуга пропала из активного каталога, но есть в корзине
	activeServices := data.GetAllServicesAdmin(app)
	for _, s := range activeServices {
		if s.ID == created1.ID {
			t.Fatalf("Service still in active catalog after moving to trash")
		}
	}
	trashServices := data.GetTrashServicesAdmin(app)
	var foundInTrash bool
	for _, s := range trashServices {
		if s.ID == created1.ID {
			foundInTrash = true
			break
		}
	}
	if !foundInTrash {
		t.Fatalf("Service not found in trash")
	}

	// 9. Восстановление из корзины в 1 клик (Restore)
	recRestore := httptest.NewRecorder()
	reqRestore := httptest.NewRequest(http.MethodPost, "/client/services/"+created1.ID+"/restore", nil)
	reqRestore.SetPathValue("id", created1.ID)
	reqRestore.AddCookie(authCookie)
	reRestore := &core.RequestEvent{App: app}
	reRestore.Request = reqRestore
	reRestore.Response = recRestore

	if err := handler.HandleRestoreService(reRestore); err != nil {
		t.Fatalf("HandleRestoreService error: %v", err)
	}
	if recRestore.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK after restore, got %d", recRestore.Code)
	}

	// Проверяем: восстановилась
	restored, err := data.GetServiceByID(app, created1.ID)
	if err != nil || restored == nil || restored.IsDeleted {
		t.Fatalf("Service was not restored properly: %+v", restored)
	}

	// 10. Окончательное удаление из корзины (Permanent delete)
	_ = data.MoveServiceToTrash(app, created1.ID)
	recPerm := httptest.NewRecorder()
	reqPerm := httptest.NewRequest(http.MethodDelete, "/client/services/"+created1.ID+"/permanent", nil)
	reqPerm.SetPathValue("id", created1.ID)
	reqPerm.AddCookie(authCookie)
	rePerm := &core.RequestEvent{App: app}
	rePerm.Request = reqPerm
	rePerm.Response = recPerm

	if err := handler.HandlePermanentDeleteService(rePerm); err != nil {
		t.Fatalf("HandlePermanentDeleteService error: %v", err)
	}
	if recPerm.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK after permanent delete, got %d", recPerm.Code)
	}

	_, errPermanentlyGone := data.GetServiceByID(app, created1.ID)
	if errPermanentlyGone == nil {
		t.Errorf("Expected service to be permanently deleted")
	}
}
