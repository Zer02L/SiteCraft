package server

import (
	"net/http"
	"os"
	"strings"

	"sitecraft/internal/media"
	"sitecraft/ui/pages"
)

func (s *Server) HandleAdminTabMedia(w http.ResponseWriter, r *http.Request) {
	files := s.scanMediaFiles()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pages.AdminMediaTab(files).Render(r.Context(), w)
}

func (s *Server) scanMediaFiles() []pages.MediaFileItem {
	var items []pages.MediaFileItem
	uploadsDir := "data/uploads/originals"
	entries, err := os.ReadDir(uploadsDir)
	if err == nil {
		for _, e := range entries {
			if !e.IsDir() {
				items = append(items, pages.MediaFileItem{
					Filename: e.Name(),
					URL:      "/media/raw/" + e.Name(),
					ThumbURL: "/media/thumb/" + e.Name(),
				})
			}
		}
	}
	// Also fallback public/ images
	pubEntries, err := os.ReadDir("public")
	if err == nil {
		for _, e := range pubEntries {
			if !e.IsDir() && (strings.HasSuffix(e.Name(), ".png") || strings.HasSuffix(e.Name(), ".jpg")) {
				items = append(items, pages.MediaFileItem{
					Filename: e.Name(),
					URL:      "/media/raw/" + e.Name(),
					ThumbURL: "/media/thumb/" + e.Name(),
				})
			}
		}
	}
	return items
}

// HandleAdminUploadMedia handles media library file uploads.
func (s *Server) HandleAdminUploadMedia(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(media.MaxUploadSize); err != nil {
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err == nil {
		defer file.Close()
		_, _ = s.Media.SaveUpload(file, header)
	}

	s.HandleAdminTabMedia(w, r)
}
