package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"

	"sitecraft/internal/cache"
	"sitecraft/internal/data"
	"sitecraft/ui/pages"

	"ganigeorgiev.com/pocketbase/pocketbase/core"
	"golang.org/x/crypto/bcrypt"
)

// HandleTabContacts renders contacts tab fragment for HTMX.
func (h *ClientAdminHandler) HandleTabContacts(e *core.RequestEvent) error {
	if !h.isAuthorized(e) {
		return e.Redirect(http.StatusFound, "/client/login")
	}
	settings := data.GetSiteSettings(h.app)
	content := data.GetSiteContent(h.app)
	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return pages.ClientContactsTab(content, settings).Render(e.Request.Context(), e.Response)
}

// HandleTabSecurity renders security tab fragment for HTMX.
func (h *ClientAdminHandler) HandleTabSecurity(e *core.RequestEvent) error {
	if !h.isAuthorized(e) {
		return e.Redirect(http.StatusFound, "/client/login")
	}
	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return pages.ClientSecurityTab("", "").Render(e.Request.Context(), e.Response)
}

// HandleChangePassword processes password change in client cabinet.
func (h *ClientAdminHandler) HandleChangePassword(e *core.RequestEvent) error {
	if !h.isAuthorized(e) {
		return e.String(http.StatusUnauthorized, "Неавторизован")
	}

	currentPassword := strings.TrimSpace(e.Request.FormValue("current_password"))
	newPassword := strings.TrimSpace(e.Request.FormValue("new_password"))
	confirmPassword := strings.TrimSpace(e.Request.FormValue("confirm_password"))

	settings := data.GetSiteSettings(h.app)

	var currentValid bool
	if settings.ClientAdminPasswordHash != "" {
		currentValid = bcrypt.CompareHashAndPassword([]byte(settings.ClientAdminPasswordHash), []byte(currentPassword)) == nil
	} else if settings.ClientAdminPassword != "" {
		currentValid = currentPassword == settings.ClientAdminPassword
	}

	if !currentValid {
		e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
		return pages.ClientSecurityTab("", "Неверный текущий пароль").Render(e.Request.Context(), e.Response)
	}

	if utf8.RuneCountInString(newPassword) < 8 {
		e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
		return pages.ClientSecurityTab("", "Новый пароль должен содержать не менее 8 символов").Render(e.Request.Context(), e.Response)
	}

	if newPassword != confirmPassword {
		e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
		return pages.ClientSecurityTab("", "Новый пароль и подтверждение не совпадают").Render(e.Request.Context(), e.Response)
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
		return pages.ClientSecurityTab("", "Ошибка шифрования нового пароля").Render(e.Request.Context(), e.Response)
	}

	if setRec, err := h.app.FindFirstRecordByFilter("site_settings", ""); err == nil && setRec != nil {
		setRec.Set("client_admin_password_hash", string(newHash))
		setRec.Set("client_admin_password", "")
		_ = h.app.Save(setRec)
	}
	cache.GetStore().InvalidateSettings()

	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return pages.ClientSecurityTab("Пароль доступа успешно изменен!", "").Render(e.Request.Context(), e.Response)
}

// HandleSaveContacts updates site_settings and site_content with validation.
func (h *ClientAdminHandler) HandleSaveContacts(e *core.RequestEvent) error {
	if !h.isAuthorized(e) {
		return e.String(http.StatusUnauthorized, "Неавторизован")
	}

	siteName := strings.TrimSpace(e.Request.FormValue("site_name"))
	tagline := strings.TrimSpace(e.Request.FormValue("tagline"))
	whatsapp := strings.TrimSpace(e.Request.FormValue("whatsapp_phone"))
	tg := strings.TrimSpace(e.Request.FormValue("telegram_contact"))

	phone1 := strings.TrimSpace(e.Request.FormValue("phone1"))
	phone2 := strings.TrimSpace(e.Request.FormValue("phone2"))
	address := strings.TrimSpace(e.Request.FormValue("address"))
	addressSub := strings.TrimSpace(e.Request.FormValue("address_sub"))
	workHours := strings.TrimSpace(e.Request.FormValue("work_hours"))
	daysOff := strings.TrimSpace(e.Request.FormValue("days_off"))

	// Валидация полей формы
	var validationErrors []string

	if utf8.RuneCountInString(siteName) < 2 {
		validationErrors = append(validationErrors, "Название компании / сайта должно содержать от 2 до 100 символов")
	} else if utf8.RuneCountInString(siteName) > 100 {
		validationErrors = append(validationErrors, "Название компании / сайта не должно превышать 100 символов")
	}

	if utf8.RuneCountInString(address) < 5 {
		validationErrors = append(validationErrors, "Адрес офиса должен содержать не менее 5 символов")
	} else if utf8.RuneCountInString(address) > 200 {
		validationErrors = append(validationErrors, "Адрес офиса не должно превышать 200 символов")
	}

	if !isValidPhone(phone1) {
		validationErrors = append(validationErrors, "Основной телефон должен быть в формате +7 (999) 999-99-99")
	}

	if phone2 != "" && !isValidPhone(phone2) {
		validationErrors = append(validationErrors, "Дополнительный телефон должен быть в формате +7 (999) 999-99-99")
	}

	if whatsapp != "" && !isValidPhone(whatsapp) {
		validationErrors = append(validationErrors, "WhatsApp должен быть в формате +7 (999) 999-99-99")
	}

	if tg != "" && !isValidTelegram(tg) {
		validationErrors = append(validationErrors, "Telegram должен быть в формате @username или t.me/username")
	}

	if len(validationErrors) > 0 {
		e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
		var errItems strings.Builder
		for _, errText := range validationErrors {
			errItems.WriteString(fmt.Sprintf("<li>%s</li>", errText))
		}
		return e.String(http.StatusOK, fmt.Sprintf(`<div class="text-xs text-rose-400 font-medium p-3 rounded-xl bg-rose-500/10 border border-rose-500/20"><div class="flex items-center gap-1.5 font-bold mb-1 text-rose-300"><svg class="w-4 h-4 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg> <span>Ошибка валидации формы:</span></div><ul class="list-disc list-inside space-y-0.5">%s</ul></div>`, errItems.String()))
	}

	// Форматирование телефонов и телеграм
	formattedPhone1 := formatPhoneMask(phone1)
	formattedPhone2 := ""
	if phone2 != "" {
		formattedPhone2 = formatPhoneMask(phone2)
	}
	formattedWA := ""
	if whatsapp != "" {
		formattedWA = formatPhoneMask(whatsapp)
	}
	formattedTG := tg
	if formattedTG != "" && !strings.HasPrefix(formattedTG, "@") && !strings.HasPrefix(formattedTG, "http") && !strings.HasPrefix(formattedTG, "t.me") {
		formattedTG = "@" + formattedTG
	}

	// Обновление site_settings
	if setRec, err := h.app.FindFirstRecordByFilter("site_settings", ""); err == nil && setRec != nil {
		if siteName != "" {
			setRec.Set("site_name", siteName)
		}
		setRec.Set("tagline", tagline)
		setRec.Set("whatsapp_phone", formattedWA)
		setRec.Set("telegram_contact", formattedTG)
		_ = h.app.Save(setRec)
	}

	// Обновление site_content
	if cntRec, err := h.app.FindFirstRecordByFilter("site_content", ""); err == nil && cntRec != nil {
		cntRec.Set("phone1", formattedPhone1)
		cntRec.Set("phone2", formattedPhone2)
		cntRec.Set("address", address)
		cntRec.Set("address_sub", addressSub)
		cntRec.Set("work_hours", workHours)
		cntRec.Set("days_off", daysOff)
		_ = h.app.Save(cntRec)
	}

	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return e.String(http.StatusOK, `<div class="text-xs text-emerald-400 font-medium flex items-center gap-1.5 p-2.5 rounded-xl bg-emerald-500/10 border border-emerald-500/20"><svg class="w-4 h-4 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path></svg> <span>Контакты и реквизиты успешно сохранены!</span></div>`)
}

func isValidPhone(phone string) bool {
	digits := extractDigits(phone)
	if len(digits) == 11 && (digits[0] == '7' || digits[0] == '8') {
		return true
	}
	if len(digits) == 10 {
		return true
	}
	return false
}

func isValidTelegram(tg string) bool {
	tg = strings.TrimSpace(tg)
	tg = strings.TrimPrefix(tg, "https://")
	tg = strings.TrimPrefix(tg, "http://")
	tg = strings.TrimPrefix(tg, "t.me/")
	tg = strings.TrimPrefix(tg, "@")
	if len(tg) < 3 || len(tg) > 32 {
		return false
	}
	for _, r := range tg {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_') {
			return false
		}
	}
	return true
}

func extractDigits(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func formatPhoneMask(phone string) string {
	d := extractDigits(phone)
	if len(d) == 11 && (d[0] == '7' || d[0] == '8') {
		return fmt.Sprintf("+7 (%s) %s-%s-%s", d[1:4], d[4:7], d[7:9], d[9:11])
	}
	if len(d) == 10 {
		return fmt.Sprintf("+7 (%s) %s-%s-%s", d[0:3], d[3:6], d[6:8], d[8:10])
	}
	return phone
}
