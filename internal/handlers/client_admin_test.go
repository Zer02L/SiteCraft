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
	defer app.ResetBootstrapState()

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

	// 2. Доступ к странице логина до настройки пароля должен отображать экран первоначальной настройки
	recSetupPage := httptest.NewRecorder()
	reqSetupPage := httptest.NewRequest(http.MethodGet, "/client/login", nil)
	reSetupPage := &core.RequestEvent{App: app, Request: reqSetupPage, Response: recSetupPage}
	if err := handler.HandleLoginPage(reSetupPage); err != nil {
		t.Fatalf("HandleLoginPage setup error: %v", err)
	}
	if !strings.Contains(recSetupPage.Body.String(), "Настройка безопасности") {
		t.Errorf("Expected initial security setup page, got: %s", recSetupPage.Body.String())
	}

	// 3. Попытка входа до настройки должна редиректить на /client/login
	recLoginBeforeSetup := httptest.NewRecorder()
	reqLoginBeforeSetup := httptest.NewRequest(http.MethodPost, "/client/login", strings.NewReader("password=admin"))
	reqLoginBeforeSetup.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reLoginBeforeSetup := &core.RequestEvent{App: app, Request: reqLoginBeforeSetup, Response: recLoginBeforeSetup}
	if err := handler.HandleLogin(reLoginBeforeSetup); err != nil {
		t.Fatalf("HandleLogin before setup error: %v", err)
	}
	if recLoginBeforeSetup.Code != http.StatusFound || recLoginBeforeSetup.Header().Get("Location") != "/client/login" {
		t.Errorf("Expected redirect to /client/login before setup, got code %d and loc %s", recLoginBeforeSetup.Code, recLoginBeforeSetup.Header().Get("Location"))
	}

	// 4. Первоначальная настройка: валидация пароля (слишком короткий)
	recSetupShort := httptest.NewRecorder()
	formSetupShort := url.Values{"password": {"short"}, "password_confirm": {"short"}}
	reqSetupShort := httptest.NewRequest(http.MethodPost, "/client/setup", strings.NewReader(formSetupShort.Encode()))
	reqSetupShort.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reSetupShort := &core.RequestEvent{App: app, Request: reqSetupShort, Response: recSetupShort}
	if err := handler.HandleInitialSetup(reSetupShort); err != nil {
		t.Fatalf("HandleInitialSetup short error: %v", err)
	}
	if !strings.Contains(recSetupShort.Body.String(), "не менее 8 символов") {
		t.Errorf("Expected min length error in body, got: %s", recSetupShort.Body.String())
	}

	// 5. Первоначальная настройка: несовпадение паролей
	recSetupMismatch := httptest.NewRecorder()
	formSetupMismatch := url.Values{"password": {"Password123!"}, "password_confirm": {"Different123!"}}
	reqSetupMismatch := httptest.NewRequest(http.MethodPost, "/client/setup", strings.NewReader(formSetupMismatch.Encode()))
	reqSetupMismatch.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reSetupMismatch := &core.RequestEvent{App: app, Request: reqSetupMismatch, Response: recSetupMismatch}
	if err := handler.HandleInitialSetup(reSetupMismatch); err != nil {
		t.Fatalf("HandleInitialSetup mismatch error: %v", err)
	}
	if !strings.Contains(recSetupMismatch.Body.String(), "не совпадают") {
		t.Errorf("Expected mismatch error in body, got: %s", recSetupMismatch.Body.String())
	}

	// 6. Успешная первоначальная установка мастер-пароля
	masterPassword := "SuperSecurePass123!"
	recSetupOK := httptest.NewRecorder()
	formSetupOK := url.Values{"password": {masterPassword}, "password_confirm": {masterPassword}}
	reqSetupOK := httptest.NewRequest(http.MethodPost, "/client/setup", strings.NewReader(formSetupOK.Encode()))
	reqSetupOK.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reSetupOK := &core.RequestEvent{App: app, Request: reqSetupOK, Response: recSetupOK}
	if err := handler.HandleInitialSetup(reSetupOK); err != nil {
		t.Fatalf("HandleInitialSetup success error: %v", err)
	}
	if recSetupOK.Code != http.StatusFound || recSetupOK.Header().Get("Location") != "/client" {
		t.Errorf("Expected redirect to /client after setup, got code %d, loc %s", recSetupOK.Code, recSetupOK.Header().Get("Location"))
	}

	// Извлекаем сессионную куку
	var authCookie *http.Cookie
	for _, c := range recSetupOK.Result().Cookies() {
		if c.Name == ClientAuthCookieName {
			authCookie = c
			break
		}
	}
	if authCookie == nil || authCookie.Value == "" || authCookie.Value == masterPassword {
		t.Fatalf("Expected valid session token cookie (not plaintext password), got: %v", authCookie)
	}

	// 7. Попытка входа с неверным паролем
	recLoginFail := httptest.NewRecorder()
	formDataFail := url.Values{"password": {"wrong_password_here"}}
	reqLoginFail := httptest.NewRequest(http.MethodPost, "/client/login", strings.NewReader(formDataFail.Encode()))
	reqLoginFail.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reLoginFail := &core.RequestEvent{App: app, Request: reqLoginFail, Response: recLoginFail}
	if err := handler.HandleLogin(reLoginFail); err != nil {
		t.Fatalf("HandleLogin error: %v", err)
	}
	if !strings.Contains(recLoginFail.Body.String(), "Неверный пароль доступа") {
		t.Errorf("Expected invalid password message in body, got: %s", recLoginFail.Body.String())
	}

	// 8. Проверка Rate Limiter: 4 дополнительные неверные попытки должны заблокировать IP
	for i := 0; i < 4; i++ {
		recFail := httptest.NewRecorder()
		reqFail := httptest.NewRequest(http.MethodPost, "/client/login", strings.NewReader(formDataFail.Encode()))
		reqFail.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		reFail := &core.RequestEvent{App: app, Request: reqFail, Response: recFail}
		_ = handler.HandleLogin(reFail)
	}
	// 5-я попытка заблокирована
	recBlocked := httptest.NewRecorder()
	reqBlocked := httptest.NewRequest(http.MethodPost, "/client/login", strings.NewReader(formDataFail.Encode()))
	reqBlocked.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reBlocked := &core.RequestEvent{App: app, Request: reqBlocked, Response: recBlocked}
	_ = handler.HandleLogin(reBlocked)
	if recBlocked.Code != http.StatusTooManyRequests {
		t.Errorf("Expected HTTP 429 TooManyRequests on rate limit block, got %d", recBlocked.Code)
	}

	// Сбрасываем rate limiter для последующих тестов
	handler.RateLimiter().Reset(reqBlocked.RemoteAddr)
	handler.RateLimiter().Reset("192.0.2.1")
	handler.RateLimiter().Reset("")

	// 9. Доступ к дашборду с авторизационной сессионной кукой
	recDash := httptest.NewRecorder()
	reqDash := httptest.NewRequest(http.MethodGet, "/client", nil)
	reqDash.AddCookie(authCookie)
	reDash := &core.RequestEvent{App: app, Request: reqDash, Response: recDash}
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
	reqContacts.AddCookie(authCookie)
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
	reqCSV.AddCookie(authCookie)
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
	reqInvalid.AddCookie(authCookie)
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
	reqTab.AddCookie(authCookie)
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
	reqStatus.AddCookie(authCookie)
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
	reqArtSave.AddCookie(authCookie)
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
	reqArtEdit.AddCookie(authCookie)
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

	// 12. Проверка вкладки «Безопасность» и функционала смены пароля
	recSecTab := httptest.NewRecorder()
	reqSecTab := httptest.NewRequest(http.MethodGet, "/client/tab/security", nil)
	reqSecTab.AddCookie(authCookie)
	reSecTab := &core.RequestEvent{App: app, Request: reqSecTab, Response: recSecTab}
	if err := handler.HandleTabSecurity(reSecTab); err != nil {
		t.Fatalf("HandleTabSecurity error: %v", err)
	}
	if !strings.Contains(recSecTab.Body.String(), "Безопасность и доступ") {
		t.Errorf("Expected security tab content, got: %s", recSecTab.Body.String())
	}

	// 13. Смена пароля: ошибка при неверном текущем пароле
	recPassWrongOld := httptest.NewRecorder()
	formWrongOld := url.Values{
		"current_password": {"WrongCurrentPass"},
		"new_password":     {"NewSecurePassword123!"},
		"confirm_password": {"NewSecurePassword123!"},
	}
	reqPassWrongOld := httptest.NewRequest(http.MethodPost, "/client/security/password", strings.NewReader(formWrongOld.Encode()))
	reqPassWrongOld.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqPassWrongOld.AddCookie(authCookie)
	rePassWrongOld := &core.RequestEvent{App: app, Request: reqPassWrongOld, Response: recPassWrongOld}
	if err := handler.HandleChangePassword(rePassWrongOld); err != nil {
		t.Fatalf("HandleChangePassword wrong old error: %v", err)
	}
	if !strings.Contains(recPassWrongOld.Body.String(), "Неверный текущий пароль") {
		t.Errorf("Expected wrong old password error, got: %s", recPassWrongOld.Body.String())
	}

	// 14. Смена пароля: успешное обновление
	newPassword := "BrandNewMasterPass2026!"
	recPassSuccess := httptest.NewRecorder()
	formPassSuccess := url.Values{
		"current_password": {masterPassword},
		"new_password":     {newPassword},
		"confirm_password": {newPassword},
	}
	reqPassSuccess := httptest.NewRequest(http.MethodPost, "/client/security/password", strings.NewReader(formPassSuccess.Encode()))
	reqPassSuccess.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqPassSuccess.AddCookie(authCookie)
	rePassSuccess := &core.RequestEvent{App: app, Request: reqPassSuccess, Response: recPassSuccess}
	if err := handler.HandleChangePassword(rePassSuccess); err != nil {
		t.Fatalf("HandleChangePassword success error: %v", err)
	}
	if !strings.Contains(recPassSuccess.Body.String(), "успешно изменен") {
		t.Errorf("Expected password change success message, got: %s", recPassSuccess.Body.String())
	}

	// 15. Проверка входа с новым паролем
	recLoginNew := httptest.NewRecorder()
	formLoginNew := url.Values{"password": {newPassword}}
	reqLoginNew := httptest.NewRequest(http.MethodPost, "/client/login", strings.NewReader(formLoginNew.Encode()))
	reqLoginNew.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reLoginNew := &core.RequestEvent{App: app, Request: reqLoginNew, Response: recLoginNew}
	if err := handler.HandleLogin(reLoginNew); err != nil {
		t.Fatalf("HandleLogin new password error: %v", err)
	}
	if recLoginNew.Code != http.StatusFound || recLoginNew.Header().Get("Location") != "/client" {
		t.Errorf("Expected redirect to /client with new password, got code %d loc %s", recLoginNew.Code, recLoginNew.Header().Get("Location"))
	}
}

func TestArticleScheduledAndTrashWorkflow(t *testing.T) {
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
	sessionToken, err := handler.SessionStore().CreateSession()
	if err != nil {
		t.Fatalf("CreateSession error: %v", err)
	}
	authCookie := &http.Cookie{Name: ClientAuthCookieName, Value: sessionToken}

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
	reqSave.AddCookie(authCookie)
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
	reqToggle.AddCookie(authCookie)
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
	reqToggle2.AddCookie(authCookie)
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
	reqTrash.AddCookie(authCookie)
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
	reqRestore.AddCookie(authCookie)
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
	reqPerm.AddCookie(authCookie)
	rePerm := &core.RequestEvent{App: app, Request: reqPerm, Response: recPerm}

	if err := handler.HandlePermanentDeleteArticle(rePerm); err != nil {
		t.Fatalf("HandlePermanentDeleteArticle error: %v", err)
	}

	if recFinal, _ := app.FindRecordById("articles", scheduledArticle.ID); recFinal != nil {
		t.Fatalf("Статья не должна существовать в БД после permanent delete")
	}
}
