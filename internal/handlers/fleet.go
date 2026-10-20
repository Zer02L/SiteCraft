package handlers

import (
	"bufio"
	"fmt"
	"html"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"sitecraft/internal/data"
	"sitecraft/internal/models"
	"sitecraft/ui/pages"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// FleetHandler manages multi-tenant instances (Fleet of Binaries).
type FleetHandler struct {
	app          *pocketbase.PocketBase
	sessionStore *SessionStore
}

// NewFleetHandler creates a new FleetHandler instance.
func NewFleetHandler(app *pocketbase.PocketBase, store *SessionStore) *FleetHandler {
	return &FleetHandler{
		app:          app,
		sessionStore: store,
	}
}

// IsAuthenticated checks if the request has a valid client admin session.
func (h *FleetHandler) IsAuthenticated(e *core.RequestEvent) bool {
	cookie, err := e.Request.Cookie(ClientAuthCookieName)
	if err != nil || cookie.Value == "" {
		return false
	}
	return h.sessionStore.IsValid(cookie.Value)
}

// HandleDashboard renders the standalone Fleet Master Dashboard page.
func (h *FleetHandler) HandleDashboard(e *core.RequestEvent) error {
	if !h.IsAuthenticated(e) {
		return e.Redirect(http.StatusFound, "/client/login?redirect=/fleet")
	}

	settings := data.GetSiteSettings(h.app)
	sites, stats := h.loadFleetSites()

	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return pages.FleetDashboardPage(settings, sites, stats).Render(e.Request.Context(), e.Response)
}

// HandleTabFleet renders the Fleet tab content inside the client admin (/client).
func (h *FleetHandler) HandleTabFleet(e *core.RequestEvent) error {
	if !h.IsAuthenticated(e) {
		return e.HTML(http.StatusUnauthorized, `<div class="p-4 text-rose-400">Сессия истекла. Обновите страницу.</div>`)
	}

	sites, stats := h.loadFleetSites()
	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return pages.FleetTabContent(sites, stats).Render(e.Request.Context(), e.Response)
}

// HandleFleetAction handles instance control actions (start, stop, restart, backup).
func (h *FleetHandler) HandleFleetAction(e *core.RequestEvent) error {
	if !h.IsAuthenticated(e) {
		return e.HTML(http.StatusUnauthorized, "Unauthorized")
	}

	siteID := strings.TrimSpace(e.Request.FormValue("site_id"))
	action := strings.TrimSpace(e.Request.FormValue("action"))

	if siteID == "" {
		return e.HTML(http.StatusBadRequest, "Missing site_id")
	}

	// Validate siteID characters to prevent command injection
	for _, r := range siteID {
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_') {
			return e.HTML(http.StatusBadRequest, "Invalid site_id format")
		}
	}

	switch action {
	case "start", "stop", "restart", "backup":
		cmd := exec.Command("./fleet.sh", action, siteID)
		output, err := cmd.CombinedOutput()
		if err != nil {
			logMsg := fmt.Sprintf("Fleet command failed: %s (output: %s)", err, string(output))
			return e.HTML(http.StatusInternalServerError, fmt.Sprintf(`<span class="text-rose-400 text-xs">%s</span>`, html.EscapeString(logMsg)))
		}
	default:
		return e.HTML(http.StatusBadRequest, "Unknown action")
	}

	// Reload site status and return updated table row
	site := h.getFleetSite(siteID)
	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return pages.FleetSiteRow(site).Render(e.Request.Context(), e.Response)
}

// HandleFleetCreate creates a new site instance via fleet.sh.
func (h *FleetHandler) HandleFleetCreate(e *core.RequestEvent) error {
	if !h.IsAuthenticated(e) {
		return e.HTML(http.StatusUnauthorized, "Unauthorized")
	}

	siteID := strings.TrimSpace(e.Request.FormValue("site_id"))
	preset := strings.TrimSpace(e.Request.FormValue("preset"))
	theme := strings.TrimSpace(e.Request.FormValue("theme"))
	port := strings.TrimSpace(e.Request.FormValue("port"))

	if siteID == "" {
		return e.HTML(http.StatusBadRequest, `<div class="p-3 bg-rose-500/10 text-rose-400 rounded-xl text-xs">Укажите ID сайта</div>`)
	}
	if preset == "" {
		preset = "legal"
	}
	if theme == "" {
		theme = preset
	}
	if port == "" {
		port = "8091"
	}

	// Run ./fleet.sh create <site-id> --preset=<preset> --theme=<theme> --port=<port>
	cmd := exec.Command("./fleet.sh", "create", siteID, "--preset="+preset, "--theme="+theme, "--port="+port)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return e.HTML(http.StatusInternalServerError, fmt.Sprintf(`
			<div class="p-3 bg-rose-500/10 border border-rose-500/20 text-rose-400 rounded-xl text-xs">
				Ошибка создания сайта: %s
			</div>
		`, html.EscapeString(string(output))))
	}

	// Trigger reload of fleet list via HTMX
	e.Response.Header().Set("HX-Trigger", "fleet-updated")
	return e.HTML(http.StatusOK, fmt.Sprintf(`
		<div class="p-4 bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 rounded-2xl text-xs space-y-1">
			<div class="font-bold">✓ Инстанс '%s' успешно создан на порту %s!</div>
			<div class="text-[11px] text-slate-400">База данных SQLite инициализирована, пресет '%s' применен.</div>
		</div>
	`, html.EscapeString(siteID), html.EscapeString(port), html.EscapeString(preset)))
}

func (h *FleetHandler) loadFleetSites() ([]models.FleetSite, models.FleetStats) {
	fleetDir := "fleet_sites"
	var sites []models.FleetSite
	var stats models.FleetStats

	entries, err := os.ReadDir(fleetDir)
	if err != nil {
		return sites, stats
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		siteID := entry.Name()
		site := h.getFleetSite(siteID)
		if site.SiteID != "" {
			sites = append(sites, site)
			stats.TotalSites++
			if site.Status == "running" {
				stats.RunningSites++
			} else {
				stats.StoppedSites++
			}
		}
	}

	return sites, stats
}

func (h *FleetHandler) getFleetSite(siteID string) models.FleetSite {
	sitePath := filepath.Join("fleet_sites", siteID)
	envPath := filepath.Join(sitePath, "site.env")
	pidPath := filepath.Join(sitePath, "app.pid")

	site := models.FleetSite{
		SiteID: siteID,
		Status: "stopped",
		PID:    "-",
	}

	// Read site.env
	if f, err := os.Open(envPath); err == nil {
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.HasPrefix(line, "#") || !strings.Contains(line, "=") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])

			switch key {
			case "PRESET":
				site.Preset = val
			case "THEME":
				site.Theme = val
			case "PORT":
				site.Port = val
			case "ADMIN_EMAIL":
				site.AdminEmail = val
			case "CREATED_AT":
				site.CreatedAt = val
			}
		}
		f.Close()
	}

	if site.Port != "" {
		site.URL = fmt.Sprintf("http://127.0.0.1:%s", site.Port)
		site.AdminURL = fmt.Sprintf("http://127.0.0.1:%s/_/", site.Port)
	}

	// Check PID alive
	if pidBytes, err := os.ReadFile(pidPath); err == nil {
		pidStr := strings.TrimSpace(string(pidBytes))
		if pid, err := strconv.Atoi(pidStr); err == nil && pid > 0 {
			process, err := os.FindProcess(pid)
			if err == nil {
				// Signal 0 checks if process exists
				if err := process.Signal(syscall.Signal(0)); err == nil {
					site.Status = "running"
					site.PID = pidStr
				}
			}
		}
	}

	return site
}
