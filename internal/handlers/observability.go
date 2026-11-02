package handlers

import (
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"ganigeorgiev.com/pocketbase/pocketbase/core"
)

// InitLogger configures the default slog structured logger.
// In production (ENV=production or PROD=1) it outputs structured JSON to stdout;
// in development it outputs human-readable text.
func InitLogger() *slog.Logger {
	isProd := os.Getenv("ENV") == "production" || os.Getenv("PROD") == "1"

	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	var handler slog.Handler
	if isProd {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger
}

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

// Flush ensures SSE streams (Chat and Order Tracker) work without buffer stalls.
func (r *statusRecorder) Flush() {
	if f, ok := r.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Unwrap provides access to the underlying ResponseWriter for standard library compatibility.
func (r *statusRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}

// RequestLoggerMiddleware provides structured HTTP access logging via log/slog.
func RequestLoggerMiddleware() func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		start := time.Now()
		path := e.Request.URL.Path

		rec := &statusRecorder{
			ResponseWriter: e.Response,
			status:         http.StatusOK,
		}
		e.Response = rec

		err := e.Next()

		latency := time.Since(start)
		status := rec.status

		// Пропускаем логирование статических ассетов, чтобы не засорять продакшен-логи
		if strings.HasPrefix(path, "/assets/") {
			return err
		}

		clientIP := e.RealIP()
		if clientIP == "" {
			clientIP = e.Request.RemoteAddr
		}

		attrs := []slog.Attr{
			slog.String("method", e.Request.Method),
			slog.String("path", path),
			slog.Int("status", status),
			slog.Duration("latency", latency),
			slog.String("ip", clientIP),
			slog.Int64("bytes", rec.written),
		}

		if err != nil {
			attrs = append(attrs, slog.String("error", err.Error()))
			slog.LogAttrs(e.Request.Context(), slog.LevelError, "http_request", attrs...)
		} else if status >= 500 {
			slog.LogAttrs(e.Request.Context(), slog.LevelError, "http_request", attrs...)
		} else if status >= 400 {
			slog.LogAttrs(e.Request.Context(), slog.LevelWarn, "http_request", attrs...)
		} else {
			slog.LogAttrs(e.Request.Context(), slog.LevelInfo, "http_request", attrs...)
		}

		return err
	}
}
