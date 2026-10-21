package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"sitecraft/internal/data"
	"sitecraft/internal/models"
	"sitecraft/ui/pages"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/filesystem"
)

// HandleTabArticles renders articles tab fragment for HTMX.
func (h *ClientAdminHandler) HandleTabArticles(e *core.RequestEvent) error {
	if !h.isAuthorized(e) {
		return e.Redirect(http.StatusFound, "/client/login")
	}
	articles := data.GetAllArticlesAdmin(h.app)
	trashArticles := data.GetTrashArticlesAdmin(h.app)
	showTrash := e.Request.URL.Query().Get("view") == "trash"

	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return pages.ClientArticlesTab(articles, trashArticles, showTrash).Render(e.Request.Context(), e.Response)
}

// HandleNewArticleForm renders form for creating a new article.
func (h *ClientAdminHandler) HandleNewArticleForm(e *core.RequestEvent) error {
	if !h.isAuthorized(e) {
		return e.Redirect(http.StatusFound, "/client/login")
	}
	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return pages.ClientArticleForm(models.ArticleItem{
		IsPublished: true,
		PublishedAt: time.Now().Format("2006-01-02 15:04"),
	}, true).Render(e.Request.Context(), e.Response)
}

// HandleEditArticleForm renders form for editing an existing article.
func (h *ClientAdminHandler) HandleEditArticleForm(e *core.RequestEvent) error {
	if !h.isAuthorized(e) {
		return e.Redirect(http.StatusFound, "/client/login")
	}

	id := e.Request.PathValue("id")
	rec, err := h.app.FindRecordById("articles", id)
	if err != nil || rec == nil {
		return e.String(http.StatusNotFound, "Статья не найдена")
	}

	article := data.RecordToArticleItem(rec)
	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return pages.ClientArticleForm(article, false).Render(e.Request.Context(), e.Response)
}

// HandleSaveArticle creates or updates an article.
func (h *ClientAdminHandler) HandleSaveArticle(e *core.RequestEvent) error {
	if !h.isAuthorized(e) {
		return e.String(http.StatusUnauthorized, "Неавторизован")
	}

	id := strings.TrimSpace(e.Request.FormValue("id"))
	title := strings.TrimSpace(e.Request.FormValue("title"))
	slug := strings.TrimSpace(e.Request.FormValue("slug"))
	excerpt := strings.TrimSpace(e.Request.FormValue("excerpt"))
	content := strings.TrimSpace(e.Request.FormValue("content"))
	author := strings.TrimSpace(e.Request.FormValue("author_name"))
	tags := strings.TrimSpace(e.Request.FormValue("tags"))
	publishedAt := strings.TrimSpace(e.Request.FormValue("published_at"))
	removeCover := e.Request.FormValue("remove_cover") == "true"
	isPublished := e.Request.FormValue("is_published") == "true"

	articlesCol, err := h.app.FindCollectionByNameOrId("articles")
	if err != nil {
		return e.String(http.StatusInternalServerError, "Коллекция articles не найдена")
	}

	var rec *core.Record
	if id != "" {
		rec, _ = h.app.FindRecordById("articles", id)
	}
	if rec == nil {
		rec = core.NewRecord(articlesCol)
	}

	if publishedAt == "" {
		if rec.GetString("published_at") != "" {
			publishedAt = rec.GetString("published_at")
		} else {
			publishedAt = time.Now().Format("2006-01-02 15:04")
		}
	} else {
		publishedAt = strings.Replace(publishedAt, "T", " ", 1)
	}

	rec.Set("title", title)
	rec.Set("slug", slug)
	rec.Set("excerpt", excerpt)
	rec.Set("content", content)
	rec.Set("author_name", author)
	rec.Set("tags", tags)
	rec.Set("published_at", publishedAt)
	rec.Set("is_published", isPublished)
	rec.Set("is_deleted", false)

	if removeCover {
		rec.Set("cover_image", nil)
	}

	if mf, fileHeader, err := e.Request.FormFile("cover_image"); err == nil && fileHeader != nil && fileHeader.Filename != "" {
		_ = mf.Close()
		if file, err := filesystem.NewFileFromMultipart(fileHeader); err == nil {
			rec.Set("cover_image", file)
		}
	}

	if err := h.app.Save(rec); err != nil {
		return e.String(http.StatusInternalServerError, fmt.Sprintf("Ошибка сохранения: %v", err))
	}

	articles := data.GetAllArticlesAdmin(h.app)
	trashArticles := data.GetTrashArticlesAdmin(h.app)
	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return pages.ClientArticlesTab(articles, trashArticles, false).Render(e.Request.Context(), e.Response)
}

// HandleToggleArticlePublish toggles article published status.
func (h *ClientAdminHandler) HandleToggleArticlePublish(e *core.RequestEvent) error {
	if !h.isAuthorized(e) {
		return e.String(http.StatusUnauthorized, "Неавторизован")
	}
	id := e.Request.PathValue("id")
	if _, err := data.ToggleArticlePublish(h.app, id); err != nil {
		return e.String(http.StatusBadRequest, err.Error())
	}
	articles := data.GetAllArticlesAdmin(h.app)
	trashArticles := data.GetTrashArticlesAdmin(h.app)
	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return pages.ClientArticlesTab(articles, trashArticles, false).Render(e.Request.Context(), e.Response)
}

// HandleTrashArticle moves article to trash.
func (h *ClientAdminHandler) HandleTrashArticle(e *core.RequestEvent) error {
	if !h.isAuthorized(e) {
		return e.String(http.StatusUnauthorized, "Неавторизован")
	}
	id := e.Request.PathValue("id")
	if err := data.MoveArticleToTrash(h.app, id); err != nil {
		return e.String(http.StatusBadRequest, err.Error())
	}
	articles := data.GetAllArticlesAdmin(h.app)
	trashArticles := data.GetTrashArticlesAdmin(h.app)
	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return pages.ClientArticlesTab(articles, trashArticles, false).Render(e.Request.Context(), e.Response)
}

// HandleRestoreArticle restores article from trash.
func (h *ClientAdminHandler) HandleRestoreArticle(e *core.RequestEvent) error {
	if !h.isAuthorized(e) {
		return e.String(http.StatusUnauthorized, "Неавторизован")
	}
	id := e.Request.PathValue("id")
	if err := data.RestoreArticleFromTrash(h.app, id); err != nil {
		return e.String(http.StatusBadRequest, err.Error())
	}
	articles := data.GetAllArticlesAdmin(h.app)
	trashArticles := data.GetTrashArticlesAdmin(h.app)
	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return pages.ClientArticlesTab(articles, trashArticles, true).Render(e.Request.Context(), e.Response)
}

// HandlePermanentDeleteArticle permanently removes an article from database.
func (h *ClientAdminHandler) HandlePermanentDeleteArticle(e *core.RequestEvent) error {
	if !h.isAuthorized(e) {
		return e.String(http.StatusUnauthorized, "Неавторизован")
	}
	id := e.Request.PathValue("id")
	if err := data.PermanentDeleteArticle(h.app, id); err != nil {
		return e.String(http.StatusBadRequest, err.Error())
	}
	articles := data.GetAllArticlesAdmin(h.app)
	trashArticles := data.GetTrashArticlesAdmin(h.app)
	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return pages.ClientArticlesTab(articles, trashArticles, true).Render(e.Request.Context(), e.Response)
}

// HandleEmptyTrashArticles permanently removes all articles in trash.
func (h *ClientAdminHandler) HandleEmptyTrashArticles(e *core.RequestEvent) error {
	if !h.isAuthorized(e) {
		return e.String(http.StatusUnauthorized, "Неавторизован")
	}
	if err := data.EmptyTrashArticles(h.app); err != nil {
		return e.String(http.StatusBadRequest, err.Error())
	}
	articles := data.GetAllArticlesAdmin(h.app)
	trashArticles := data.GetTrashArticlesAdmin(h.app)
	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return pages.ClientArticlesTab(articles, trashArticles, true).Render(e.Request.Context(), e.Response)
}
