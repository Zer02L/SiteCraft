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

func TestClientAdminWorkflow(t *testing.T) {
	testDir := t.TempDir()
	app := pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDataDir: testDir,
	})

	if err := app.Bootstrap(); err != nil {
		t.Fatalf("Failed to bootstrap PB: %v", err)
	}
	if err := data.InitSchema(app); err != nil {
		t.Fatalf("Failed to init schema: %v", err)
	}

	handler := NewClientAdminHandler(app)

	// 1. Неавторизованный доступ к дашборду должен редиректить на логин
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/client", nil)
	re := &core.RequestEvent{
		App: app,
	}
	re.Request = req
	re.Response = rec
	if err := handler.HandleDashboard(re); err != nil {
		t.Fatalf("HandleDashboard error: %v", err)
	}
	if rec.Code != http.StatusFound {
		t.Errorf("Expected 302 Found redirect, got %d", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "/client/login" {
		t.Errorf("Expected redirect to /client/login, got %s", loc)
	}

	// 2. Ввод неверного пароля
	recLoginFail := httptest.NewRecorder()
	formDataFail := url.Values{"password": {"wrong_pass"}}
	reqLoginFail := httptest.NewRequest(http.MethodPost, "/client/login", strings.NewReader(formDataFail.Encode()))
	reqLoginFail.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reLoginFail := &core.RequestEvent{App: app}
	reLoginFail.Request = reqLoginFail
	reLoginFail.Response = recLoginFail
	if err := handler.HandleLogin(reLoginFail); err != nil {
		t.Fatalf("HandleLogin error: %v", err)
	}
	if !strings.Contains(recLoginFail.Body.String(), "Неверный пароль доступа") {
		t.Errorf("Expected invalid password message in body")
	}

	// 3. Успешный вход с дефолтным паролем 'admin'
	recLoginOK := httptest.NewRecorder()
	formDataOK := url.Values{"password": {"admin"}}
	reqLoginOK := httptest.NewRequest(http.MethodPost, "/client/login", strings.NewReader(formDataOK.Encode()))
	reqLoginOK.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reLoginOK := &core.RequestEvent{App: app}
	reLoginOK.Request = reqLoginOK
	reLoginOK.Response = recLoginOK
	if err := handler.HandleLogin(reLoginOK); err != nil {
		t.Fatalf("HandleLogin error: %v", err)
	}
	if recLoginOK.Code != http.StatusFound {
		t.Errorf("Expected 302 redirect after login, got %d", recLoginOK.Code)
	}
	if loc := recLoginOK.Header().Get("Location"); loc != "/client" {
		t.Errorf("Expected redirect to /client, got %s", loc)
	}

	// 4. Доступ к дашборду с авторизационной кукой
	recDash := httptest.NewRecorder()
	reqDash := httptest.NewRequest(http.MethodGet, "/client", nil)
	reqDash.AddCookie(&http.Cookie{
		Name:  ClientAuthCookieName,
		Value: "admin",
	})
	reDash := &core.RequestEvent{App: app}
	reDash.Request = reqDash
	reDash.Response = recDash
	if err := handler.HandleDashboard(reDash); err != nil {
		t.Fatalf("HandleDashboard error: %v", err)
	}
	if recDash.Code != http.StatusOK {
		t.Errorf("Expected 200 OK for authorized dashboard, got %d", recDash.Code)
	}
	if !strings.Contains(recDash.Body.String(), "Кабинет управления") {
		t.Errorf("Expected dashboard content in response")
	}

	// 5. Сохранение контактов
	recContacts := httptest.NewRecorder()
	contactForm := url.Values{
		"site_name": {"Адвокатское бюро Премиум"},
		"tagline":   {"Надежная защита ваших интересов"},
		"phone1":    {"+7 (495) 123-45-67"},
		"phone2":    {"+7 (495) 765-43-21"},
		"address":   {"г. Москва, ул. Тверская, д. 10"},
	}
	reqContacts := httptest.NewRequest(http.MethodPost, "/client/contacts", strings.NewReader(contactForm.Encode()))
	reqContacts.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqContacts.AddCookie(&http.Cookie{
		Name:  ClientAuthCookieName,
		Value: "admin",
	})
	reContacts := &core.RequestEvent{App: app}
	reContacts.Request = reqContacts
	reContacts.Response = recContacts
	if err := handler.HandleSaveContacts(reContacts); err != nil {
		t.Fatalf("HandleSaveContacts error: %v", err)
	}
	if !strings.Contains(recContacts.Body.String(), "успешно сохранены") {
		t.Errorf("Expected success confirmation on contacts save")
	}

	// 6. Проверка экспорта заявок в CSV
	recCSV := httptest.NewRecorder()
	reqCSV := httptest.NewRequest(http.MethodGet, "/client/export/leads.csv", nil)
	reqCSV.AddCookie(&http.Cookie{
		Name:  ClientAuthCookieName,
		Value: "admin",
	})
	reCSV := &core.RequestEvent{App: app}
	reCSV.Request = reqCSV
	reCSV.Response = recCSV
	if err := handler.HandleExportLeadsCSV(reCSV); err != nil {
		t.Fatalf("HandleExportLeadsCSV error: %v", err)
	}
	if recCSV.Header().Get("Content-Type") != "text/csv; charset=utf-8" {
		t.Errorf("Expected Content-Type text/csv, got %s", recCSV.Header().Get("Content-Type"))
	}

	// 7. Валидация контактов: отправка некорректных данных (мусорные поля) должна возвращать ошибку
	recInvalid := httptest.NewRecorder()
	invalidForm := url.Values{
		"site_name":        {"A"},                    // Слишком короткое
		"phone1":           {"abc-not-a-phone"},     // Некорректный телефон
		"phone2":           {"123"},                 // Неполный телефон
		"whatsapp_phone":   {"garbage"},             // Некорректный WA
		"telegram_contact": {"invalid!@#$username"}, // Недопустимые символы
		"address":          {"ул."},                 // Менее 5 символов
	}
	reqInvalid := httptest.NewRequest(http.MethodPost, "/client/contacts", strings.NewReader(invalidForm.Encode()))
	reqInvalid.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqInvalid.AddCookie(&http.Cookie{
		Name:  ClientAuthCookieName,
		Value: "admin",
	})
	reInvalid := &core.RequestEvent{App: app}
	reInvalid.Request = reqInvalid
	reInvalid.Response = recInvalid
	if err := handler.HandleSaveContacts(reInvalid); err != nil {
		t.Fatalf("HandleSaveContacts invalid test error: %v", err)
	}
	invalidResp := recInvalid.Body.String()
	if !strings.Contains(invalidResp, "Ошибка валидации формы") {
		t.Errorf("Expected validation error header, got: %s", invalidResp)
	}
	if !strings.Contains(invalidResp, "Основной телефон") {
		t.Errorf("Expected phone1 error in validation response")
	}
	if !strings.Contains(invalidResp, "Адрес офиса") {
		t.Errorf("Expected address error in validation response")
	}

	// 8. Проверка подсветки активной вкладки при переходе по ?tab=contacts
	recTab := httptest.NewRecorder()
	reqTab := httptest.NewRequest(http.MethodGet, "/client?tab=contacts", nil)
	reqTab.AddCookie(&http.Cookie{
		Name:  ClientAuthCookieName,
		Value: "admin",
	})
	reTab := &core.RequestEvent{App: app}
	reTab.Request = reqTab
	reTab.Response = recTab
	if err := handler.HandleDashboard(reTab); err != nil {
		t.Fatalf("HandleDashboard tab error: %v", err)
	}
	tabBody := recTab.Body.String()
	if !strings.Contains(tabBody, "currentTab") || !strings.Contains(tabBody, "contacts") {
		t.Errorf("Expected currentTab with contacts in dashboard response, got: %s", tabBody)
	}

	// 9. Создание заявки и проверка смены статуса с обновлением OOB счетчиков
	leadsCol, err := app.FindCollectionByNameOrId("leads")
	if err != nil {
		t.Fatalf("leads collection error: %v", err)
	}
	leadRec := core.NewRecord(leadsCol)
	leadRec.Set("name", "Иван Тестовый")
	leadRec.Set("phone", "+7 (999) 111-22-33")
	leadRec.Set("status", "новая")
	if err := app.Save(leadRec); err != nil {
		t.Fatalf("save lead error: %v", err)
	}

	recStatus := httptest.NewRecorder()
	statusForm := url.Values{"status": {"в обработке"}}
	reqStatus := httptest.NewRequest(http.MethodPost, "/client/leads/"+leadRec.Id+"/status", strings.NewReader(statusForm.Encode()))
	reqStatus.SetPathValue("id", leadRec.Id)
	reqStatus.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqStatus.AddCookie(&http.Cookie{
		Name:  ClientAuthCookieName,
		Value: "admin",
	})
	reStatus := &core.RequestEvent{App: app}
	reStatus.Request = reqStatus
	reStatus.Response = recStatus
	if err := handler.HandleUpdateLeadStatus(reStatus); err != nil {
		t.Fatalf("HandleUpdateLeadStatus error: %v", err)
	}
	statusBody := recStatus.Body.String()
	if !strings.Contains(statusBody, "hx-swap-oob=\"true\"") {
		t.Errorf("Expected OOB swap for badge counters in response")
	}
	if !strings.Contains(statusBody, "Заявки в обработке: 1") {
		t.Errorf("Expected updated in-progress count 1 in OOB badges, got: %s", statusBody)
	}

	// 10. Сохранение статьи с тегами, слагом и HTML контентом
	recArtSave := httptest.NewRecorder()
	artForm := url.Values{
		"title":        {"Как подготовить документы для сделки в 2026 году"},
		"slug":         {"kak-podgotovit-dokumenty-dlya-sdelki"},
		"excerpt":      {"Краткий лид статьи"},
		"content":      {"<h2>Заголовок</h2><p>Текст статьи с <strong>жирным</strong> начертанием</p>"},
		"author_name":  {"Иван Эксперт"},
		"tags":         {"Недвижимость, Сделки, Советы"},
		"is_published": {"true"},
	}
	reqArtSave := httptest.NewRequest(http.MethodPost, "/client/articles/save", strings.NewReader(artForm.Encode()))
	reqArtSave.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqArtSave.AddCookie(&http.Cookie{
		Name:  ClientAuthCookieName,
		Value: "admin",
	})
	reArtSave := &core.RequestEvent{App: app}
	reArtSave.Request = reqArtSave
	reArtSave.Response = recArtSave
	if err := handler.HandleSaveArticle(reArtSave); err != nil {
		t.Fatalf("HandleSaveArticle error: %v", err)
	}
	if recArtSave.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK after saving article, got %d", recArtSave.Code)
	}

	articles := data.GetAllArticlesAdmin(app)
	var createdArticle *models.ArticleItem
	for i := range articles {
		if articles[i].Slug == "kak-podgotovit-dokumenty-dlya-sdelki" {
			createdArticle = &articles[i]
			break
		}
	}
	if createdArticle == nil {
		t.Fatalf("Created article not found in GetAllArticlesAdmin")
	}
	if len(createdArticle.Tags) != 3 || createdArticle.Tags[0] != "Недвижимость" {
		t.Errorf("Expected tags [Недвижимость, Сделки, Советы], got: %v", createdArticle.Tags)
	}

	// 11. Рендеринг формы редактирования созданной статьи
	recArtEdit := httptest.NewRecorder()
	reqArtEdit := httptest.NewRequest(http.MethodGet, "/client/articles/edit/"+createdArticle.ID, nil)
	reqArtEdit.SetPathValue("id", createdArticle.ID)
	reqArtEdit.AddCookie(&http.Cookie{
		Name:  ClientAuthCookieName,
		Value: "admin",
	})
	reArtEdit := &core.RequestEvent{App: app}
	reArtEdit.Request = reqArtEdit
	reArtEdit.Response = recArtEdit
	if err := handler.HandleEditArticleForm(reArtEdit); err != nil {
		t.Fatalf("HandleEditArticleForm error: %v", err)
	}
	artEditBody := recArtEdit.Body.String()
	if !strings.Contains(artEditBody, "articleEditor") {
		t.Errorf("Expected articleEditor Alpine component in edit form")
	}
	if !strings.Contains(artEditBody, "kak-podgotovit-dokumenty-dlya-sdelki") {
		t.Errorf("Expected article slug in edit form")
	}
	if !strings.Contains(artEditBody, "Недвижимость") {
		t.Errorf("Expected tags in edit form")
	}
}

func TestArticleScheduledAndTrashWorkflow(t *testing.T) {
	testDir := t.TempDir()
	app := pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDataDir: testDir,
	})

	if err := app.Bootstrap(); err != nil {
		t.Fatalf("Failed to bootstrap PB: %v", err)
	}
	if err := data.InitSchema(app); err != nil {
		t.Fatalf("Failed to init schema: %v", err)
	}

	handler := NewClientAdminHandler(app)

	// 1. Создаем статью с отложенной публикацией (в будущем: 2099-01-01 10:00)
	futureDate := "2099-01-01 10:00"
	artForm := url.Values{
		"title":        {"Будущая публикация 2099"},
		"slug":         {"budushchaya-publikatsiya-2099"},
		"content":      {"<p>Секретный контент из будущего</p>"},
		"published_at": {futureDate},
		"is_published": {"true"},
	}
	recSave := httptest.NewRecorder()
	reqSave := httptest.NewRequest(http.MethodPost, "/client/articles/save", strings.NewReader(artForm.Encode()))
	reqSave.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqSave.AddCookie(&http.Cookie{Name: ClientAuthCookieName, Value: "admin"})
	reSave := &core.RequestEvent{App: app, Request: reqSave, Response: recSave}

	if err := handler.HandleSaveArticle(reSave); err != nil {
		t.Fatalf("HandleSaveArticle error: %v", err)
	}

	// 2. Проверяем, что статья НЕ видна на публичном сайте (отложенная)
	publicArticles := data.GetArticles(app)
	for _, a := range publicArticles {
		if a.Slug == "budushchaya-publikatsiya-2099" {
			t.Fatalf("Отложенная статья не должна отображаться в публичном блоге!")
		}
	}
	if _, err := data.GetArticleBySlug(app, "budushchaya-publikatsiya-2099"); err == nil {
		t.Fatalf("GetArticleBySlug не должен находить отложенную статью!")
	}

	// 3. Но видна в панели администратора как запланированная
	adminArticles := data.GetAllArticlesAdmin(app)
	var scheduledArticle *models.ArticleItem
	for i := range adminArticles {
		if adminArticles[i].Slug == "budushchaya-publikatsiya-2099" {
			scheduledArticle = &adminArticles[i]
			break
		}
	}
	if scheduledArticle == nil {
		t.Fatalf("Статья не найдена в админке")
	}
	if !scheduledArticle.IsScheduled() {
		t.Errorf("Ожидалось IsScheduled() = true для даты %s", scheduledArticle.PublishedAt)
	}

	// 4. Проверяем быстрый тоггл скрытия/публикации в 1 клик
	recToggle := httptest.NewRecorder()
	reqToggle := httptest.NewRequest(http.MethodPost, "/client/articles/toggle/"+scheduledArticle.ID, nil)
	reqToggle.SetPathValue("id", scheduledArticle.ID)
	reqToggle.AddCookie(&http.Cookie{Name: ClientAuthCookieName, Value: "admin"})
	reToggle := &core.RequestEvent{App: app, Request: reqToggle, Response: recToggle}

	if err := handler.HandleToggleArticlePublish(reToggle); err != nil {
		t.Fatalf("HandleToggleArticlePublish error: %v", err)
	}

	recDB, _ := app.FindRecordById("articles", scheduledArticle.ID)
	if recDB.GetBool("is_published") {
		t.Errorf("После первого toggle статья должна была стать скрытой (is_published = false)")
	}

	// Второй тоггл - возвращаем публикацию
	recToggle2 := httptest.NewRecorder()
	reqToggle2 := httptest.NewRequest(http.MethodPost, "/client/articles/toggle/"+scheduledArticle.ID, nil)
	reqToggle2.SetPathValue("id", scheduledArticle.ID)
	reqToggle2.AddCookie(&http.Cookie{Name: ClientAuthCookieName, Value: "admin"})
	reToggle2 := &core.RequestEvent{App: app, Request: reqToggle2, Response: recToggle2}
	_ = handler.HandleToggleArticlePublish(reToggle2)

	recDB2, _ := app.FindRecordById("articles", scheduledArticle.ID)
	if !recDB2.GetBool("is_published") {
		t.Errorf("После второго toggle статья должна была стать опубликованной (is_published = true)")
	}

	// 5. Перемещение в корзину (Soft delete)
	recTrash := httptest.NewRecorder()
	reqTrash := httptest.NewRequest(http.MethodPost, "/client/articles/trash/"+scheduledArticle.ID, nil)
	reqTrash.SetPathValue("id", scheduledArticle.ID)
	reqTrash.AddCookie(&http.Cookie{Name: ClientAuthCookieName, Value: "admin"})
	reTrash := &core.RequestEvent{App: app, Request: reqTrash, Response: recTrash}

	if err := handler.HandleTrashArticle(reTrash); err != nil {
		t.Fatalf("HandleTrashArticle error: %v", err)
	}

	// Проверяем, что статья исчезла из активного списка админки
	activeAfterTrash := data.GetAllArticlesAdmin(app)
	for _, a := range activeAfterTrash {
		if a.ID == scheduledArticle.ID {
			t.Fatalf("Удаленная в корзину статья не должна отображаться в активном списке!")
		}
	}

	// Проверяем, что статья появилась в списке корзины
	trashArticles := data.GetTrashArticlesAdmin(app)
	foundInTrash := false
	for _, a := range trashArticles {
		if a.ID == scheduledArticle.ID {
			foundInTrash = true
			if !a.IsDeleted || a.DeletedAt == "" {
				t.Errorf("Ожидалось IsDeleted=true и заполненный DeletedAt")
			}
			break
		}
	}
	if !foundInTrash {
		t.Fatalf("Статья не найдена в корзине GetTrashArticlesAdmin")
	}

	// 6. Восстановление из корзины
	recRestore := httptest.NewRecorder()
	reqRestore := httptest.NewRequest(http.MethodPost, "/client/articles/restore/"+scheduledArticle.ID, nil)
	reqRestore.SetPathValue("id", scheduledArticle.ID)
	reqRestore.AddCookie(&http.Cookie{Name: ClientAuthCookieName, Value: "admin"})
	reRestore := &core.RequestEvent{App: app, Request: reqRestore, Response: recRestore}

	if err := handler.HandleRestoreArticle(reRestore); err != nil {
		t.Fatalf("HandleRestoreArticle error: %v", err)
	}

	// Проверяем, что статья вернулась в активный список
	activeAfterRestore := data.GetAllArticlesAdmin(app)
	foundAfterRestore := false
	for _, a := range activeAfterRestore {
		if a.ID == scheduledArticle.ID {
			foundAfterRestore = true
			break
		}
	}
	if !foundAfterRestore {
		t.Fatalf("Восстановленная статья должна присутствовать в активном списке!")
	}

	// 7. Окончательное удаление (Permanent delete)
	_ = handler.HandleTrashArticle(reTrash) // снова в корзину
	recPerm := httptest.NewRecorder()
	reqPerm := httptest.NewRequest(http.MethodDelete, "/client/articles/permanent/"+scheduledArticle.ID, nil)
	reqPerm.SetPathValue("id", scheduledArticle.ID)
	reqPerm.AddCookie(&http.Cookie{Name: ClientAuthCookieName, Value: "admin"})
	rePerm := &core.RequestEvent{App: app, Request: reqPerm, Response: recPerm}

	if err := handler.HandlePermanentDeleteArticle(rePerm); err != nil {
		t.Fatalf("HandlePermanentDeleteArticle error: %v", err)
	}

	if recFinal, _ := app.FindRecordById("articles", scheduledArticle.ID); recFinal != nil {
		t.Fatalf("Статья не должна существовать в БД после permanent delete")
	}
}
