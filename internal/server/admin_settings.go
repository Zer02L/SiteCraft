package server

import (
	"net/http"

	"sitecraft/internal/typography"
	"sitecraft/ui/pages"
)

func (s *Server) HandleAdminTabSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	settings, _ := s.Repo.GetSiteSettings(ctx)
	content, _ := s.Repo.GetSiteContent(ctx)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pages.AdminSettingsTab(settings, content, "").Render(ctx, w)
}

func (s *Server) HandleAdminSaveSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_ = r.ParseForm()

	settings, _ := s.Repo.GetSiteSettings(ctx)
	content, _ := s.Repo.GetSiteContent(ctx)

	if val := r.FormValue("site_name"); val != "" {
		settings.SiteName = typography.Format(val)
	}
	settings.Tagline = typography.Format(r.FormValue("tagline"))
	if val := r.FormValue("theme"); val != "" {
		settings.Theme = val
	}
	if val := r.FormValue("primary_color"); val != "" {
		settings.PrimaryColor = val
	}
	if val := r.FormValue("accent_color"); val != "" {
		settings.AccentColor = val
	}
	if val := r.FormValue("font_family"); val != "" {
		settings.FontFamily = val
	}
	flavor := r.FormValue("design_flavor")
	bgColor := r.FormValue("background_color")
	if flavor != "" || bgColor != "" {
		if flavor == "" {
			flavor = settings.GetDesignFlavor()
		}
		if bgColor != "" {
			settings.BackgroundColor = flavor + "|" + bgColor
		} else {
			settings.BackgroundColor = flavor
		}
	}
	settings.HeaderVariant = r.FormValue("header_variant")
	settings.YandexMetrikaID = r.FormValue("yandex_metrika_id")
	settings.WebhookURL = r.FormValue("webhook_url")
	settings.TelegramBotToken = r.FormValue("telegram_bot_token")
	settings.TelegramChatID = r.FormValue("telegram_chat_id")
	settings.ContactEmail = r.FormValue("contact_email")
	settings.LegalEntity = r.FormValue("legal_entity")
	settings.INN = r.FormValue("inn")

	if val := r.FormValue("phone1"); val != "" {
		content.Phone1 = val
		_ = s.Repo.UpsertSiteContent(ctx, content)
	}

	_ = s.Repo.UpsertSiteSettings(ctx, settings)
	s.publishSiteMutated(ctx, "settings")

	w.Header().Set("HX-Trigger", "refreshChecklist")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pages.AdminSettingsTab(settings, content, "Настройки успешно сохранены!").Render(ctx, w)
}

func (s *Server) HandleAdminTabContacts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	settings, _ := s.Repo.GetSiteSettings(ctx)
	content, _ := s.Repo.GetSiteContent(ctx)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pages.ClientContactsTab(content, settings).Render(ctx, w)
}

func (s *Server) HandleAdminSaveContacts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_ = r.ParseForm()
	content, _ := s.Repo.GetSiteContent(ctx)

	content.Phone1 = r.FormValue("phone1")
	content.Phone2 = r.FormValue("phone2")
	content.WorkHours = r.FormValue("work_hours")
	content.Address = r.FormValue("address")
	if addressSub := r.FormValue("address_sub"); addressSub != "" {
		content.AddressSub = addressSub
	}
	if daysOff := r.FormValue("days_off"); daysOff != "" {
		content.DaysOff = daysOff
	}
	_ = s.Repo.UpsertSiteContent(ctx, content)

	settings, err := s.Repo.GetSiteSettings(ctx)
	if err == nil {
		updated := false
		if siteName := r.FormValue("site_name"); siteName != "" {
			settings.SiteName = typography.Format(siteName)
			updated = true
		}
		if tagline := r.FormValue("tagline"); tagline != "" {
			settings.Tagline = typography.Format(tagline)
			updated = true
		}
		if wa := r.FormValue("whatsapp_phone"); wa != "" {
			settings.WhatsappPhone = wa
			updated = true
		}
		if tg := r.FormValue("telegram_contact"); tg != "" {
			settings.TelegramContact = tg
			updated = true
		}
		if updated {
			_ = s.Repo.UpsertSiteSettings(ctx, settings)
		}
	}

	s.publishSiteMutated(ctx, "contacts")
	w.Header().Set("HX-Trigger", "refreshChecklist")
	s.HandleAdminTabContacts(w, r)
}
