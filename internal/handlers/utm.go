package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"ganigeorgiev.com/pocketbase/pocketbase/core"
)

const UTMCookieName = "sitecraft_utm"

// UTMData stores captured marketing campaign attribution parameters.
type UTMData struct {
	Source   string `json:"source,omitempty"`
	Medium   string `json:"medium,omitempty"`
	Campaign string `json:"campaign,omitempty"`
	Content  string `json:"content,omitempty"`
	Term     string `json:"term,omitempty"`
	Yclid    string `json:"yclid,omitempty"`
	Gclid    string `json:"gclid,omitempty"`
}

// IsEmpty returns true if no attribution tags are present.
func (u UTMData) IsEmpty() bool {
	return u.Source == "" && u.Medium == "" && u.Campaign == "" &&
		u.Content == "" && u.Term == "" && u.Yclid == "" && u.Gclid == ""
}

type utmContextKey struct{}

// WithUTM returns a child context carrying UTM parameters.
func WithUTM(ctx context.Context, utm UTMData) context.Context {
	return context.WithValue(ctx, utmContextKey{}, utm)
}

// GetUTM extracts UTM parameters from the context.
func GetUTM(ctx context.Context) UTMData {
	if val, ok := ctx.Value(utmContextKey{}).(UTMData); ok {
		return val
	}
	return UTMData{}
}

// ExtractUTMFromQuery parses attribution parameters from URL query string.
func ExtractUTMFromQuery(q url.Values) UTMData {
	return UTMData{
		Source:   strings.TrimSpace(q.Get("utm_source")),
		Medium:   strings.TrimSpace(q.Get("utm_medium")),
		Campaign: strings.TrimSpace(q.Get("utm_campaign")),
		Content:  strings.TrimSpace(q.Get("utm_content")),
		Term:     strings.TrimSpace(q.Get("utm_term")),
		Yclid:    strings.TrimSpace(q.Get("yclid")),
		Gclid:    strings.TrimSpace(q.Get("gclid")),
	}
}

// UTMMiddleware extracts UTM tags from query params, persists them to a 30-day cookie,
// and injects UTMData into request context.
func UTMMiddleware() func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		queryUTM := ExtractUTMFromQuery(e.Request.URL.Query())

		var utmData UTMData

		if !queryUTM.IsEmpty() {
			utmData = queryUTM
			if rawJSON, err := json.Marshal(utmData); err == nil {
				http.SetCookie(e.Response, &http.Cookie{
					Name:     UTMCookieName,
					Value:    url.QueryEscape(string(rawJSON)),
					Path:     "/",
					MaxAge:   30 * 24 * 3600, // 30 дней сохранения атрибуции
					SameSite: http.SameSiteLaxMode,
					HttpOnly: false, // доступно скриптам при необходимости
				})
			}
		} else if cookie, err := e.Request.Cookie(UTMCookieName); err == nil && cookie != nil && cookie.Value != "" {
			if decoded, err := url.QueryUnescape(cookie.Value); err == nil {
				_ = json.Unmarshal([]byte(decoded), &utmData)
			}
		}

		ctx := WithUTM(e.Request.Context(), utmData)
		e.Request = e.Request.WithContext(ctx)

		return e.Next()
	}
}
