package server

import (
	"log/slog"
	"net/http"
	"strings"

	"sitecraft/internal/auth"
	"sitecraft/ui/pages"
)

// HandleAdminLoginPage renders admin login screen.
func (s *Server) HandleAdminLoginPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if s.SessionMgr.GetBool(ctx, auth.SessionKeyAdminAuth) {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}
	settings, _ := s.Repo.GetSiteSettings(ctx)
	errMsg := s.SessionMgr.PopString(ctx, auth.SessionKeyFlashError)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pages.ClientLogin(settings, errMsg).Render(ctx, w)
}

// HandleAdminLogin authenticates administrator.
func (s *Server) HandleAdminLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_ = r.ParseForm()
	password := strings.TrimSpace(r.FormValue("password"))

	settings, _ := s.Repo.GetSiteSettings(ctx)

	authenticated := false
	if settings.ClientAdminPasswordHash != "" {
		authenticated = auth.CheckPassword(password, settings.ClientAdminPasswordHash)
	} else if settings.ClientAdminPassword != "" {
		authenticated = (password == settings.ClientAdminPassword)
	} else {
		// Default password if not configured
		authenticated = (password == "admin" || password == "sitecraft" || password == "admin123")
	}

	if !authenticated {
		s.SessionMgr.Put(ctx, auth.SessionKeyFlashError, "Неверный пароль доступа")
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
		return
	}

	_ = auth.SetAdminSession(ctx, s.SessionMgr, "admin@sitecraft.local")
	slog.Info("Admin logged in successfully", "ip", r.RemoteAddr)
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

// HandleAdminLogout clears session.
func (s *Server) HandleAdminLogout(w http.ResponseWriter, r *http.Request) {
	_ = auth.ClearAdminSession(r.Context(), s.SessionMgr)
	http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
}

// HandleAdminChangePassword updates the administrator password and returns ClientSecurityTab.
func (s *Server) HandleAdminChangePassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_ = r.ParseForm()
	currentPass := strings.TrimSpace(r.FormValue("current_password"))
	newPass := strings.TrimSpace(r.FormValue("new_password"))
	confirmPass := strings.TrimSpace(r.FormValue("confirm_password"))

	settings, _ := s.Repo.GetSiteSettings(ctx)

	var msg, errMsg string
	if currentPass != "" {
		validCurrent := false
		if settings.ClientAdminPasswordHash != "" {
			validCurrent = auth.CheckPassword(currentPass, settings.ClientAdminPasswordHash)
		} else if settings.ClientAdminPassword != "" {
			validCurrent = (currentPass == settings.ClientAdminPassword)
		} else {
			validCurrent = (currentPass == "admin" || currentPass == "sitecraft" || currentPass == "admin123")
		}
		if !validCurrent {
			errMsg = "Текущий пароль указан неверно."
		}
	}

	if errMsg == "" && len(newPass) < 8 {
		errMsg = "Новый пароль должен содержать не менее 8 символов."
	}
	if errMsg == "" && confirmPass != "" && confirmPass != newPass {
		errMsg = "Новые пароли не совпадают."
	}

	if errMsg == "" {
		hash, err := auth.HashPassword(newPass)
		if err != nil {
			errMsg = "Не удалось зашифровать пароль."
		} else {
			settings.ClientAdminPasswordHash = hash
			settings.ClientAdminPassword = ""
			if err := s.Repo.UpsertSiteSettings(ctx, settings); err != nil {
				errMsg = "Ошибка сохранения пароля."
			} else {
				msg = "Пароль доступа успешно обновлён."
				w.Header().Set("HX-Trigger-After-Settle", `{"adminToast":{"message":"Пароль обновлён","tone":"success"}}`)
			}
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pages.ClientSecurityTab(msg, errMsg).Render(ctx, w)
}
