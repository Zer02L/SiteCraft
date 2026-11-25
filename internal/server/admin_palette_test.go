package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"sitecraft/internal/auth"
)

func TestAdminExtractPaletteEndpoint(t *testing.T) {
	srv, conn := setupTestServer(t)
	defer conn.Close()

	ctx := context.Background()
	sessionCookie, err := srv.SessionMgr.Load(ctx, "")
	if err != nil {
		t.Fatalf("failed to init session: %v", err)
	}
	ctx = sessionCookie
	_ = auth.SetAdminSession(ctx, srv.SessionMgr, "admin@sitecraft.local")
	_, _, _ = srv.SessionMgr.Commit(ctx)

	// 1. Create a synthetic test PNG image (60x60, red and blue quadrants)
	img := image.NewRGBA(image.Rect(0, 0, 60, 60))
	for y := 0; y < 60; y++ {
		for x := 0; x < 60; x++ {
			if y < 30 {
				img.Set(x, y, color.RGBA{R: 37, G: 99, B: 235, A: 255}) // Blue-600
			} else {
				img.Set(x, y, color.RGBA{R: 255, G: 255, B: 255, A: 255}) // White
			}
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("failed to encode test png: %v", err)
	}

	// 2. Prepare multipart request
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("image", "sample.png")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	if _, err := part.Write(buf.Bytes()); err != nil {
		t.Fatalf("failed to write png bytes: %v", err)
	}
	_ = writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/admin/api/theme/extract-palette", body)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()

	srv.Router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from /admin/api/theme/extract-palette, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Success        bool     `json:"success"`
		Accent         string   `json:"accent"`
		Primary        string   `json:"primary"`
		Surface        string   `json:"surface"`
		SuggestedTheme string   `json:"suggested_theme"`
		DominantColors []struct {
			Hex string `json:"hex"`
		} `json:"dominant_colors"`
		Palettes []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"palettes"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response JSON: %v, body: %s", err, rec.Body.String())
	}

	if !resp.Success {
		t.Errorf("expected success: true, got %+v", resp)
	}

	if resp.Accent == "" || resp.Primary == "" || resp.Surface == "" {
		t.Errorf("expected non-empty accent, primary, and surface, got %+v", resp)
	}

	if len(resp.DominantColors) == 0 {
		t.Errorf("expected dominant colors to be present")
	}

	if len(resp.Palettes) < 4 {
		t.Errorf("expected at least 4 curated palettes, got %d", len(resp.Palettes))
	}
}
