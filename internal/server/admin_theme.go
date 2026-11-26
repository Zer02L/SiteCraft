package server

import (
	"encoding/json"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	_ "golang.org/x/image/webp"

	"sitecraft/internal/theme"
)

const (
	SessionKeyPrevTheme   = "prev_theme"
	SessionKeyPrevPrimary = "prev_primary_color"
	SessionKeyPrevAccent  = "prev_accent_color"
	SessionKeyPrevBg      = "prev_background_color"
	SessionKeyPrevFont    = "prev_font_family"
)

func (s *Server) HandleAdminSaveThemeTokens(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_ = r.ParseForm()

	settings, _ := s.Repo.GetSiteSettings(ctx)

	if theme := r.FormValue("theme"); theme != "" {
		settings.Theme = theme
	}
	if primary := r.FormValue("primary_color"); primary != "" {
		settings.PrimaryColor = primary
	}
	if accent := r.FormValue("accent_color"); accent != "" {
		settings.AccentColor = accent
	}
	if font := r.FormValue("font_family"); font != "" {
		settings.FontFamily = font
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

	if err := s.Repo.UpsertSiteSettings(ctx, settings); err != nil {
		slog.Error("failed to save theme tokens", "error", err)
		http.Error(w, "Ошибка сохранения токенов", http.StatusInternalServerError)
		return
	}

	s.publishSiteMutated(ctx, "theme")
	w.Header().Set("HX-Trigger", "refreshChecklist")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(`<span class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-xl bg-emerald-50 border border-emerald-300 text-emerald-800 text-xs font-semibold animate-fade-in"><span class="w-2 h-2 rounded-full bg-emerald-500 animate-pulse"></span>Тема и токены сохранены в БД!</span>`))
}

// HandleAdminApplyThemePreset applies a theme preset in 1 click, saves previous scheme to session for rollback, and updates site_settings atomically.
func (s *Server) HandleAdminApplyThemePreset(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_ = r.ParseForm()

	presetID := r.FormValue("preset_id")
	p, ok := theme.GetPreset(presetID)
	if !ok {
		http.Error(w, "Пресет не найден", http.StatusBadRequest)
		return
	}

	settings, err := s.Repo.GetSiteSettings(ctx)
	if err != nil {
		http.Error(w, "Ошибка получения настроек сайта", http.StatusInternalServerError)
		return
	}

	// Backup current settings into session for 1-click rollback
	if s.SessionMgr != nil {
		s.SessionMgr.Put(ctx, SessionKeyPrevTheme, settings.Theme)
		s.SessionMgr.Put(ctx, SessionKeyPrevPrimary, settings.PrimaryColor)
		s.SessionMgr.Put(ctx, SessionKeyPrevAccent, settings.AccentColor)
		s.SessionMgr.Put(ctx, SessionKeyPrevBg, settings.BackgroundColor)
		s.SessionMgr.Put(ctx, SessionKeyPrevFont, settings.FontFamily)
	}

	// Apply preset parameters
	settings.Theme = p.Theme
	settings.PrimaryColor = p.PrimaryColor
	settings.AccentColor = p.AccentColor
	settings.BackgroundColor = p.BackgroundColor
	settings.FontFamily = p.FontFamily

	if err := s.Repo.UpsertSiteSettings(ctx, settings); err != nil {
		slog.Error("failed to apply theme preset", "preset_id", presetID, "error", err)
		http.Error(w, "Ошибка сохранения темы", http.StatusInternalServerError)
		return
	}

	s.publishSiteMutated(ctx, "theme")
	w.Header().Set("HX-Trigger", "refreshChecklist")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// Return status badge with rollback button
	resp := fmt.Sprintf(`<div id="theme-preset-status" class="flex items-center gap-2 px-3 py-2 rounded-xl bg-emerald-50 border border-emerald-300 text-emerald-900 text-xs font-medium animate-fade-in">
		<span class="w-2 h-2 rounded-full bg-emerald-500 animate-pulse shrink-0"></span>
		<span>Стиль <strong>«%s»</strong> успешно применен!</span>
		<button type="button" hx-post="/admin/api/theme/rollback" hx-target="#theme-preset-status" hx-swap="outerHTML" class="ml-auto underline hover:text-emerald-950 font-bold cursor-pointer">Откатить назад</button>
	</div>`, p.Name)
	_, _ = w.Write([]byte(resp))
}

// HandleAdminRollbackThemePreset reverts theme settings from session in 1 click.
func (s *Server) HandleAdminRollbackThemePreset(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if s.SessionMgr == nil || !s.SessionMgr.Exists(ctx, SessionKeyPrevTheme) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<div class="text-xs text-slate-500 italic">Нет сохраненной темы для отката</div>`))
		return
	}

	settings, err := s.Repo.GetSiteSettings(ctx)
	if err != nil {
		http.Error(w, "Ошибка получения настроек", http.StatusInternalServerError)
		return
	}

	settings.Theme = s.SessionMgr.GetString(ctx, SessionKeyPrevTheme)
	settings.PrimaryColor = s.SessionMgr.GetString(ctx, SessionKeyPrevPrimary)
	settings.AccentColor = s.SessionMgr.GetString(ctx, SessionKeyPrevAccent)
	settings.BackgroundColor = s.SessionMgr.GetString(ctx, SessionKeyPrevBg)
	settings.FontFamily = s.SessionMgr.GetString(ctx, SessionKeyPrevFont)

	if err := s.Repo.UpsertSiteSettings(ctx, settings); err != nil {
		slog.Error("failed to rollback theme preset", "error", err)
		http.Error(w, "Ошибка отката темы", http.StatusInternalServerError)
		return
	}

	// Remove session backup
	s.SessionMgr.Remove(ctx, SessionKeyPrevTheme)
	s.SessionMgr.Remove(ctx, SessionKeyPrevPrimary)
	s.SessionMgr.Remove(ctx, SessionKeyPrevAccent)
	s.SessionMgr.Remove(ctx, SessionKeyPrevBg)
	s.SessionMgr.Remove(ctx, SessionKeyPrevFont)

	s.publishSiteMutated(ctx, "theme")
	w.Header().Set("HX-Trigger", "refreshChecklist")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(`<div id="theme-preset-status" class="flex items-center gap-1.5 px-3 py-2 rounded-xl bg-blue-50 border border-blue-300 text-blue-900 text-xs font-medium animate-fade-in">
		<span class="w-2 h-2 rounded-full bg-blue-500 shrink-0"></span>
		<span>Предыдущая тема успешно восстановлена!</span>
	</div>`))
}

// HandleAdminExtractPalette extracts dominant colors and curated 60-30-10 palettes from an uploaded image.
func (s *Server) HandleAdminExtractPalette(w http.ResponseWriter, r *http.Request) {
	// Support max 10MB image uploads
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		_ = r.ParseForm()
	}

	var imgReader io.Reader

	// Check if a file was uploaded directly
	file, _, err := r.FormFile("image")
	if err != nil {
		file, _, err = r.FormFile("file")
	}

	if err == nil && file != nil {
		defer file.Close()
		imgReader = file
	} else {
		// Alternatively, check for image_url or path in form values
		imgPath := r.FormValue("image_url")
		if imgPath == "" {
			imgPath = r.FormValue("path")
		}

		if imgPath == "" {
			http.Error(w, "Файл изображения не передан", http.StatusBadRequest)
			return
		}

		// Security: clean path and prevent path traversal
		cleanPath := filepath.Clean(strings.TrimPrefix(imgPath, "/"))
		if strings.HasPrefix(cleanPath, "..") {
			http.Error(w, "Недопустимый путь к изображению", http.StatusBadRequest)
			return
		}

		// Look up in media directory or public directory
		var f *os.File
		if s.Media != nil {
			var localFile string
			localFile, err = s.Media.GetOrGenerateThumbnail("raw", filepath.Base(cleanPath))
			if err == nil {
				f, err = os.Open(localFile)
			}
		}
		if f == nil {
			f, err = os.Open(cleanPath)
		}
		if err != nil {
			http.Error(w, "Изображение не найдено", http.StatusNotFound)
			return
		}
		defer f.Close()
		imgReader = f
	}

	img, _, err := image.Decode(imgReader)
	if err != nil {
		slog.Warn("failed to decode image for palette extraction", "error", err)
		http.Error(w, "Не удалось декодировать изображение. Поддерживаются JPEG, PNG, WebP и GIF.", http.StatusBadRequest)
		return
	}

	palette, err := theme.ExtractPaletteFromImage(img)
	if err != nil {
		slog.Error("failed to extract palette from image", "error", err)
		http.Error(w, "Ошибка извлечения палитры: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"success":          true,
		"accent":           palette.Accent,
		"primary":          palette.Primary,
		"surface":          palette.Surface,
		"suggested_theme":  palette.SuggestedTheme,
		"suggested_flavor": palette.SuggestedFlavor,
		"dominant_colors":  palette.DominantColors,
		"palettes":         palette.Palettes,
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(resp)
}
