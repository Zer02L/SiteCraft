package data

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"sitecraft/internal/db"
	"sitecraft/internal/models"

	"github.com/google/uuid"
)

// NormalizeRedirectPath normalizes URL paths and extracts paths from full URLs.
func NormalizeRedirectPath(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return ""
	}
	if strings.HasPrefix(p, "http://") || strings.HasPrefix(p, "https://") {
		if u, err := url.Parse(p); err == nil {
			p = u.RequestURI()
		}
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return p
}

// GetRedirectsMap loads all redirects into an in-memory map keyed by normalized source path.
// Serves from in-memory cache when available for sub-millisecond lookups.
func (r *Repository) GetRedirectsMap(ctx context.Context) (map[string]models.RedirectRule, error) {
	if cached, ok := r.Cache().GetRedirects(); ok {
		return cached, nil
	}
	dbList, err := r.queries.ListRedirects(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list redirects: %w", err)
	}
	rules := make(map[string]models.RedirectRule, len(dbList))
	for _, item := range dbList {
		norm := NormalizeRedirectPath(item.SourcePath)
		rule := models.RedirectRule{
			ID:         item.ID,
			SourcePath: item.SourcePath,
			TargetPath: item.TargetPath,
			StatusCode: int(item.StatusCode),
			CreatedAt:  item.CreatedAt,
		}
		rules[norm] = rule
		// Also register trimmed trailing slash variant if path has no query params
		if !strings.Contains(norm, "?") {
			trimmed := strings.TrimSuffix(norm, "/")
			if trimmed != "" && trimmed != norm {
				rules[trimmed] = rule
			}
			withSlash := trimmed + "/"
			if withSlash != norm {
				rules[withSlash] = rule
			}
		}
	}
	r.Cache().SetRedirects(rules)
	return rules, nil
}

// FindRedirect performs in-memory lookup for an incoming path/URI in < 0.1 ms.
func (r *Repository) FindRedirect(ctx context.Context, path, requestURI string) (*models.RedirectRule, bool) {
	rules, err := r.GetRedirectsMap(ctx)
	if err != nil || len(rules) == 0 {
		return nil, false
	}

	normReq := NormalizeRedirectPath(requestURI)
	if rule, ok := rules[normReq]; ok {
		return &rule, true
	}

	normPath := NormalizeRedirectPath(path)
	if rule, ok := rules[normPath]; ok {
		return &rule, true
	}

	trimmed := strings.TrimSuffix(normPath, "/")
	if rule, ok := rules[trimmed]; ok {
		return &rule, true
	}

	withSlash := trimmed + "/"
	if rule, ok := rules[withSlash]; ok {
		return &rule, true
	}

	return nil, false
}

// ListRedirects returns all redirects ordered by creation date descending.
func (r *Repository) ListRedirects(ctx context.Context) ([]models.RedirectRule, error) {
	dbList, err := r.queries.ListRedirects(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list redirects: %w", err)
	}
	out := make([]models.RedirectRule, len(dbList))
	for i, item := range dbList {
		out[i] = models.RedirectRule{
			ID:         item.ID,
			SourcePath: item.SourcePath,
			TargetPath: item.TargetPath,
			StatusCode: int(item.StatusCode),
			CreatedAt:  item.CreatedAt,
		}
	}
	return out, nil
}

// GetRedirectByID retrieves a redirect by its UUID.
func (r *Repository) GetRedirectByID(ctx context.Context, id string) (models.RedirectRule, error) {
	item, err := r.queries.GetRedirectByID(ctx, id)
	if err != nil {
		return models.RedirectRule{}, err
	}
	return models.RedirectRule{
		ID:         item.ID,
		SourcePath: item.SourcePath,
		TargetPath: item.TargetPath,
		StatusCode: int(item.StatusCode),
		CreatedAt:  item.CreatedAt,
	}, nil
}

// CreateRedirect creates a new redirect rule and invalidates in-memory cache.
func (r *Repository) CreateRedirect(ctx context.Context, rule models.RedirectRule) (models.RedirectRule, error) {
	if rule.ID == "" {
		rule.ID = uuid.New().String()
	}
	if rule.StatusCode <= 0 {
		rule.StatusCode = 301
	}
	rule.SourcePath = NormalizeRedirectPath(rule.SourcePath)
	rule.TargetPath = strings.TrimSpace(rule.TargetPath)
	if rule.SourcePath == "" || rule.TargetPath == "" {
		return models.RedirectRule{}, fmt.Errorf("source and target paths cannot be empty")
	}

	res, err := r.queries.CreateRedirect(ctx, db.CreateRedirectParams{
		ID:         rule.ID,
		SourcePath: rule.SourcePath,
		TargetPath: rule.TargetPath,
		StatusCode: int64(rule.StatusCode),
	})
	if err != nil {
		return models.RedirectRule{}, err
	}
	r.Cache().InvalidateRedirects()
	return models.RedirectRule{
		ID:         res.ID,
		SourcePath: res.SourcePath,
		TargetPath: res.TargetPath,
		StatusCode: int(res.StatusCode),
		CreatedAt:  res.CreatedAt,
	}, nil
}

// UpsertRedirect inserts or updates a redirect rule by source_path.
func (r *Repository) UpsertRedirect(ctx context.Context, rule models.RedirectRule) (models.RedirectRule, error) {
	if rule.ID == "" {
		rule.ID = uuid.New().String()
	}
	if rule.StatusCode <= 0 {
		rule.StatusCode = 301
	}
	rule.SourcePath = NormalizeRedirectPath(rule.SourcePath)
	rule.TargetPath = strings.TrimSpace(rule.TargetPath)
	if rule.SourcePath == "" || rule.TargetPath == "" {
		return models.RedirectRule{}, fmt.Errorf("source and target paths cannot be empty")
	}

	res, err := r.queries.UpsertRedirect(ctx, db.UpsertRedirectParams{
		ID:         rule.ID,
		SourcePath: rule.SourcePath,
		TargetPath: rule.TargetPath,
		StatusCode: int64(rule.StatusCode),
	})
	if err != nil {
		return models.RedirectRule{}, err
	}
	r.Cache().InvalidateRedirects()
	return models.RedirectRule{
		ID:         res.ID,
		SourcePath: res.SourcePath,
		TargetPath: res.TargetPath,
		StatusCode: int(res.StatusCode),
		CreatedAt:  res.CreatedAt,
	}, nil
}

// DeleteRedirect removes a redirect rule and invalidates in-memory cache.
func (r *Repository) DeleteRedirect(ctx context.Context, id string) error {
	err := r.queries.DeleteRedirect(ctx, id)
	if err != nil {
		return err
	}
	r.Cache().InvalidateRedirects()
	return nil
}

// CountRedirects returns total number of redirects configured.
func (r *Repository) CountRedirects(ctx context.Context) (int64, error) {
	return r.queries.CountRedirects(ctx)
}
