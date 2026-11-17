package server

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"sitecraft/internal/importer"
	"sitecraft/internal/models"
	"sitecraft/ui/pages"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// HandleAdminTabImporter renders the WordPress Importer and 301-redirects management tab.
func (s *Server) HandleAdminTabImporter(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	redirects, err := s.Repo.ListRedirects(ctx)
	if err != nil {
		slog.Error("failed to list redirects", "error", err)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pages.ClientImporterTab(redirects, "").Render(ctx, w)
}

// HandleAdminImporterUpload accepts a WordPress WXR XML file and launches async background import.
func (s *Server) HandleAdminImporterUpload(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form up to 64MB
	if err := r.ParseMultipartForm(64 << 20); err != nil {
		http.Error(w, "Файл слишком велик (макс 64MB)", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("xml_file")
	if err != nil {
		http.Error(w, "Ошибка получения файла: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	createRedirects := r.FormValue("create_redirects") == "1"
	importPosts := r.FormValue("import_posts") == "1"
	importPages := r.FormValue("import_pages") == "1"
	cleanShortcodes := r.FormValue("clean_shortcodes") == "1"
	sanitizeHTML := r.FormValue("sanitize_html") == "1"

	opts := importer.Options{
		CreateRedirects: createRedirects,
		ImportPosts:     importPosts,
		ImportPages:     importPages,
		CleanShortcodes: cleanShortcodes,
		SanitizeHTML:    sanitizeHTML,
		DefaultAuthor:   "Редакция",
	}

	tempFile, err := os.CreateTemp("", "sitecraft_wxr_*.xml")
	if err != nil {
		http.Error(w, "Не удалось создать временный файл: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tempPath := tempFile.Name()

	if _, err := io.Copy(tempFile, file); err != nil {
		_ = tempFile.Close()
		_ = os.Remove(tempPath)
		http.Error(w, "Ошибка сохранения файла: "+err.Error(), http.StatusInternalServerError)
		return
	}
	_ = tempFile.Close()

	jobID := uuid.New().String()
	job := importer.DefaultJobManager.CreateJob(jobID)

	// Launch async background import goroutine
	go func() {
		defer os.Remove(tempPath)
		f, err := os.Open(tempPath)
		if err != nil {
			slog.Error("failed to open temp XML file", "error", err)
			return
		}
		defer f.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
		defer cancel()

		if s.Importer == nil {
			s.Importer = importer.New(s.Repo, s.Media)
		}

		_, _ = s.Importer.ParseAndImport(ctx, f, opts, func(p importer.Progress) {
			job.Progress = p
			select {
			case job.Updates <- p:
			default:
			}
		})
	}()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pages.ImporterSSEContainer(jobID).Render(r.Context(), w)
}

// HandleAdminImporterSSE streams live import progress updates via Server-Sent Events.
func (s *Server) HandleAdminImporterSSE(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "jobID")
	job, found := importer.DefaultJobManager.GetJob(jobID)
	if !found {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	renderProgress := func(p importer.Progress) {
		var buf bytes.Buffer
		_ = pages.ImporterProgressFragment(models.RedirectRule{}, p.CurrentItem, p.Percentage, p.IsComplete, p.PostsImported, p.PagesImported, p.RedirectsCreated, p.Errors).Render(r.Context(), &buf)
		dataPayload := strings.ReplaceAll(buf.String(), "\n", "")
		_, _ = fmt.Fprintf(w, "event: progress\ndata: %s\n\n", dataPayload)
		flusher.Flush()
	}

	// Send initial state immediately
	renderProgress(job.Progress)

	for {
		select {
		case <-r.Context().Done():
			return
		case p, ok := <-job.Updates:
			if !ok {
				return
			}
			renderProgress(p)
			if p.IsComplete {
				return
			}
		}
	}
}

// HandleAdminCreateRedirect creates a manual 301 redirect and re-renders redirects table.
func (s *Server) HandleAdminCreateRedirect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_ = r.ParseForm()
	sourcePath := strings.TrimSpace(r.FormValue("source_path"))
	targetPath := strings.TrimSpace(r.FormValue("target_path"))
	statusCode, _ := strconv.Atoi(r.FormValue("status_code"))
	if statusCode <= 0 {
		statusCode = 301
	}

	if sourcePath != "" && targetPath != "" {
		_, err := s.Repo.CreateRedirect(ctx, models.RedirectRule{
			SourcePath: sourcePath,
			TargetPath: targetPath,
			StatusCode: statusCode,
		})
		if err != nil {
			slog.Error("failed to create redirect", "error", err)
		}
	}

	redirects, _ := s.Repo.ListRedirects(ctx)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pages.RedirectsTable(redirects).Render(ctx, w)
}

// HandleAdminDeleteRedirect deletes a redirect rule by ID.
func (s *Server) HandleAdminDeleteRedirect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	if id != "" {
		_ = s.Repo.DeleteRedirect(ctx, id)
	}
	w.WriteHeader(http.StatusOK)
}
