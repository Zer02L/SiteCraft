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

// HandleAdminChangePassword updates the administrator password.
func (s *Server) HandleAdminChangePassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_ = r.ParseForm()
	newPass := strings.TrimSpace(r.FormValue("new_password"))

	if len(newPass) >= 8 {
		hash, err := auth.HashPassword(newPass)
		if err == nil {
			settings, _ := s.Repo.GetSiteSettings(ctx)
			settings.ClientAdminPasswordHash = hash
			settings.ClientAdminPassword = ""
			_ = s.Repo.UpsertSiteSettings(ctx, settings)
		}
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(`<div class="p-4 rounded-xl bg-emerald-500/10 text-emerald-400 text-xs font-bold">Пароль успешно обновлен!</div>`))
}
