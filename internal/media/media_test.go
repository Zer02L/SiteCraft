package media_test

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"sitecraft/internal/media"
)

func TestMediaUploadAndThumbnail(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "sitecraft_media_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	svc, err := media.NewService(tempDir)
	if err != nil {
		t.Fatalf("NewService failed: %v", err)
	}

	// Create test in-memory PNG
	img := image.NewRGBA(image.Rect(0, 0, 400, 300))
	for x := 0; x < 400; x++ {
		for y := 0; y < 300; y++ {
			img.Set(x, y, color.RGBA{R: 200, G: 100, B: 50, A: 255})
		}
	}
	var imgBuf bytes.Buffer
	if err := png.Encode(&imgBuf, img); err != nil {
		t.Fatalf("png encode failed: %v", err)
	}

	// Prepare multipart request
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "sample_image.png")
	if err != nil {
		t.Fatalf("CreateFormFile failed: %v", err)
	}
	_, _ = part.Write(imgBuf.Bytes())
	_ = writer.Close()

	req := httptest.NewRequest("POST", "/media/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rr := httptest.NewRecorder()

	uploadHandler := svc.UploadHandler()
	uploadHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("upload handler failed with status %d: %s", rr.Code, rr.Body.String())
	}

	// Test serving thumbnail
	parts := bytes.Split(rr.Body.Bytes(), []byte(`"filename":"`))
	if len(parts) < 2 {
		t.Fatalf("unexpected JSON: %s", rr.Body.String())
	}
	filename := string(bytes.Split(parts[1], []byte(`"`))[0])

	thumbReq := httptest.NewRequest("GET", "/media/thumb/"+filename, nil)
	thumbRR := httptest.NewRecorder()
	serveHandler := svc.ServeMediaHandler()
	serveHandler.ServeHTTP(thumbRR, thumbReq)

	if thumbRR.Code != http.StatusOK {
		t.Fatalf("serve media failed with status %d", thumbRR.Code)
	}
	if cacheCtrl := thumbRR.Header().Get("Cache-Control"); cacheCtrl == "" {
		t.Errorf("expected Cache-Control header, got empty")
	}

	// Test serving with WebP content negotiation
	webpPath := tempDir + "/originals/" + filename[:len(filename)-4] + ".webp"
	_ = os.WriteFile(webpPath, []byte("fake-webp-content"), 0644)

	rawReq := httptest.NewRequest("GET", "/media/raw/"+filename, nil)
	rawReq.Header.Set("Accept", "image/webp,image/apng,image/*,*/*;q=0.8")
	rawRR := httptest.NewRecorder()
	serveHandler.ServeHTTP(rawRR, rawReq)

	if rawRR.Code != http.StatusOK {
		t.Fatalf("serve media with webp accept failed with status %d", rawRR.Code)
	}
	if ct := rawRR.Header().Get("Content-Type"); ct != "image/webp" {
		t.Errorf("expected Content-Type image/webp, got %s", ct)
	}
	if vary := rawRR.Header().Get("Vary"); vary != "Accept" {
		t.Errorf("expected Vary: Accept, got %s", vary)
	}
}
