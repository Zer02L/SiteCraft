package server

import (
	"log/slog"
	"net/http"

	"sitecraft/internal/models"
	"sitecraft/internal/typography"
	"sitecraft/ui/pages"

	"github.com/go-chi/chi/v5"
)

func (s *Server) HandleAdminTabBlocks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	blocks, _ := s.Repo.GetAllPageBlocks(ctx)
	settings, _ := s.Repo.GetSiteSettings(ctx)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pages.AdminBlocksTab(blocks, settings).Render(ctx, w)
}

func (s *Server) HandleAdminCreateBlock(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_ = r.ParseForm()
	blockType := r.FormValue("type")
	title := typography.Format(r.FormValue("title"))
	content := typography.Format(r.FormValue("content"))

	blocks, _ := s.Repo.GetAllPageBlocks(ctx)
	order := len(blocks) + 1

	_, err := s.Repo.CreatePageBlock(ctx, models.PageBlock{
		Type:    blockType,
		Title:   title,
		Content: content,
		Order:   order,
		Active:  true,
		Payload: map[string]interface{}{},
	})
	if err != nil {
		slog.Error("failed to create page block", "error", err)
	} else {
		s.publishSiteMutated(ctx, "blocks")
	}

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Trigger", "refreshChecklist")
		s.HandleAdminTabBlocks(w, r)
		return
	}
	http.Redirect(w, r, "/admin?tab=blocks", http.StatusSeeOther)
}

func (s *Server) HandleAdminUpdateBlock(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_ = r.ParseForm()
	id := chi.URLParam(r, "id")
	if id == "" {
		id = r.FormValue("id")
	}
	title := r.FormValue("title")
	content := r.FormValue("content")
	blockType := r.FormValue("type")
	variant := r.FormValue("variant")
	slog.Info("HandleAdminUpdateBlock called", "id", id, "title", title, "type", blockType, "variant", variant)

	if id != "" {
		block, err := s.Repo.GetPageBlockByID(ctx, id)
		if err == nil {
			if title != "" {
				block.Title = typography.Format(title)
			}
			block.Content = typography.Format(content)
			if blockType != "" {
				block.Type = blockType
			}
			if variant != "" {
				if block.Payload == nil {
					block.Payload = make(map[string]interface{})
				}
				block.Payload["variant"] = variant
			}
			if err := s.Repo.UpdatePageBlock(ctx, block); err != nil {
				slog.Error("failed to update page block", "id", id, "error", err)
			} else {
				s.publishSiteMutated(ctx, "blocks")
			}
		} else {
			slog.Error("page block not found for update", "id", id, "error", err)
		}
	}

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Trigger", "refreshChecklist")
		s.HandleAdminTabBlocks(w, r)
		return
	}
	http.Redirect(w, r, "/admin?tab=blocks", http.StatusSeeOther)
}

func (s *Server) HandleAdminToggleBlock(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	if err := s.Repo.TogglePageBlock(ctx, id); err != nil {
		slog.Error("failed to toggle page block", "id", id, "error", err)
	} else {
		s.publishSiteMutated(ctx, "blocks")
	}
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Trigger", "refreshChecklist")
		s.HandleAdminTabBlocks(w, r)
		return
	}
	http.Redirect(w, r, "/admin?tab=blocks", http.StatusSeeOther)
}

func (s *Server) HandleAdminMoveBlockUp(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	blocks, _ := s.Repo.GetAllPageBlocks(ctx)
	for i, b := range blocks {
		if b.ID == id && i > 0 {
			prev := blocks[i-1]
			if err := s.Repo.SwapPageBlockOrder(ctx, b.ID, int64(b.Order), prev.ID, int64(prev.Order)); err != nil {
				slog.Error("failed to move block up", "id", id, "error", err)
			} else {
				s.publishSiteMutated(ctx, "blocks")
			}
			break
		}
	}
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Trigger", "refreshChecklist")
		s.HandleAdminTabBlocks(w, r)
		return
	}
	http.Redirect(w, r, "/admin?tab=blocks", http.StatusSeeOther)
}

func (s *Server) HandleAdminMoveBlockDown(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	blocks, _ := s.Repo.GetAllPageBlocks(ctx)
	for i, b := range blocks {
		if b.ID == id && i < len(blocks)-1 {
			next := blocks[i+1]
			if err := s.Repo.SwapPageBlockOrder(ctx, b.ID, int64(b.Order), next.ID, int64(next.Order)); err != nil {
				slog.Error("failed to move block down", "id", id, "error", err)
			} else {
				s.publishSiteMutated(ctx, "blocks")
			}
			break
		}
	}
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Trigger", "refreshChecklist")
		s.HandleAdminTabBlocks(w, r)
		return
	}
	http.Redirect(w, r, "/admin?tab=blocks", http.StatusSeeOther)
}

func (s *Server) HandleAdminDeleteBlock(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	if err := s.Repo.DeletePageBlock(ctx, id); err != nil {
		slog.Error("failed to delete page block", "id", id, "error", err)
	} else {
		s.publishSiteMutated(ctx, "blocks")
	}
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Trigger", "refreshChecklist")
		s.HandleAdminTabBlocks(w, r)
		return
	}
	http.Redirect(w, r, "/admin?tab=blocks", http.StatusSeeOther)
}
