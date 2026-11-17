package models

// RedirectRule represents a URL redirection rule (e.g. 301 Moved Permanently)
// used to preserve SEO rankings during migrations from WordPress or legacy platforms.
type RedirectRule struct {
	ID         string `json:"id"`
	SourcePath string `json:"source_path"`
	TargetPath string `json:"target_path"`
	StatusCode int    `json:"status_code"`
	CreatedAt  string `json:"created_at"`
}
