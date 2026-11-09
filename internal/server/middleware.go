package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// --- A/B Testing ---
const ABCookieName = "sitecraft_ab"

type abContextKey struct{}

func WithABVariant(ctx context.Context, variant string) context.Context {
	return context.WithValue(ctx, abContextKey{}, variant)
}

func GetABVariant(ctx context.Context) string {
	if val, ok := ctx.Value(abContextKey{}).(string); ok && (val == "A" || val == "B") {
		return val
	}
	return "A"
}

func HTTPABTestingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var variant string
		if qParam := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("ab"))); qParam == "A" || qParam == "B" {
			variant = qParam
			http.SetCookie(w, &http.Cookie{
				Name:     ABCookieName,
				Value:    variant,
				Path:     "/",
				MaxAge:   30 * 24 * 3600,
				SameSite: http.SameSiteLaxMode,
			})
		} else if cookie, err := r.Cookie(ABCookieName); err == nil && cookie != nil && (cookie.Value == "A" || cookie.Value == "B") {
			variant = cookie.Value
		} else {
			if rand.IntN(2) == 0 {
				variant = "A"
			} else {
				variant = "B"
			}
			http.SetCookie(w, &http.Cookie{
				Name:     ABCookieName,
				Value:    variant,
				Path:     "/",
				MaxAge:   30 * 24 * 3600,
				SameSite: http.SameSiteLaxMode,
			})
		}
		next.ServeHTTP(w, r.WithContext(WithABVariant(r.Context(), variant)))
	})
}

// --- UTM Attribution ---
const UTMCookieName = "sitecraft_utm"

type utmContextKey struct{}

type UTMData struct {
	Source   string `json:"source,omitempty"`
	Medium   string `json:"medium,omitempty"`
	Campaign string `json:"campaign,omitempty"`
	Content  string `json:"content,omitempty"`
	Term     string `json:"term,omitempty"`
	Yclid    string `json:"yclid,omitempty"`
	Gclid    string `json:"gclid,omitempty"`
}

func (u UTMData) IsEmpty() bool {
	return u.Source == "" && u.Medium == "" && u.Campaign == "" &&
		u.Content == "" && u.Term == "" && u.Yclid == "" && u.Gclid == ""
}

func HTTPUTMMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		queryUTM := UTMData{
			Source:   strings.TrimSpace(q.Get("utm_source")),
			Medium:   strings.TrimSpace(q.Get("utm_medium")),
			Campaign: strings.TrimSpace(q.Get("utm_campaign")),
			Content:  strings.TrimSpace(q.Get("utm_content")),
			Term:     strings.TrimSpace(q.Get("utm_term")),
			Yclid:    strings.TrimSpace(q.Get("yclid")),
			Gclid:    strings.TrimSpace(q.Get("gclid")),
		}

		var utmData UTMData
		if !queryUTM.IsEmpty() {
			utmData = queryUTM
			if rawJSON, err := json.Marshal(utmData); err == nil {
				http.SetCookie(w, &http.Cookie{
					Name:     UTMCookieName,
					Value:    url.QueryEscape(string(rawJSON)),
					Path:     "/",
					MaxAge:   30 * 24 * 3600,
					SameSite: http.SameSiteLaxMode,
				})
			}
		} else if cookie, err := r.Cookie(UTMCookieName); err == nil && cookie != nil && cookie.Value != "" {
			if decoded, err := url.QueryUnescape(cookie.Value); err == nil {
				_ = json.Unmarshal([]byte(decoded), &utmData)
			}
		}

		ctx := context.WithValue(r.Context(), utmContextKey{}, utmData)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// --- Cache-Control ---
func HTTPCacheControlMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		method := r.Method

		switch {
		case strings.HasPrefix(path, "/assets/"), strings.HasPrefix(path, "/media/"):
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		case strings.HasPrefix(path, "/admin"), strings.HasPrefix(path, "/client"), strings.HasPrefix(path, "/fleet"):
			w.Header().Set("Cache-Control", "private, no-cache, no-store, must-revalidate")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
		case strings.HasSuffix(path, "/sse") || strings.Contains(path, "/sse/"):
			w.Header().Set("Cache-Control", "no-cache, no-transform")
			w.Header().Set("Connection", "keep-alive")
		case strings.HasPrefix(path, "/blog/") && method == http.MethodGet:
			if w.Header().Get("Cache-Control") == "" {
				w.Header().Set("Cache-Control", "public, max-age=300, stale-while-revalidate=3600")
			}
		case method == http.MethodGet:
			if w.Header().Get("Cache-Control") == "" {
				w.Header().Set("Cache-Control", "private, no-cache")
			}
		default:
			if w.Header().Get("Cache-Control") == "" {
				w.Header().Set("Cache-Control", "no-store")
			}
		}

		next.ServeHTTP(w, r)
	})
}

// --- Structured Request Logger ---
type statusRecorder struct {
	http.ResponseWriter
	status  int
	written int64
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	n, err := r.ResponseWriter.Write(b)
	r.written += int64(n)
	return n, err
}

func (r *statusRecorder) Flush() {
	if f, ok := r.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func HTTPRequestLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		path := r.URL.Path

		rec := &statusRecorder{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		next.ServeHTTP(rec, r)

		if strings.HasPrefix(path, "/assets/") {
			return
		}

		latency := time.Since(start)
		status := rec.status

		clientIP := r.RemoteAddr
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			clientIP = strings.TrimSpace(strings.Split(xff, ",")[0])
		} else if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
			clientIP = xrip
		}

		attrs := []slog.Attr{
			slog.String("method", r.Method),
			slog.String("path", path),
			slog.Int("status", status),
			slog.Duration("latency", latency),
			slog.String("ip", clientIP),
			slog.Int64("bytes", rec.written),
		}

		if status >= 500 {
			slog.LogAttrs(r.Context(), slog.LevelError, "http_request", attrs...)
		} else if status >= 400 {
			slog.LogAttrs(r.Context(), slog.LevelWarn, "http_request", attrs...)
		} else {
			slog.LogAttrs(r.Context(), slog.LevelInfo, "http_request", attrs...)
		}
	})
}
