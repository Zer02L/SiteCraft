package media

import (
	"encoding/json"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/google/uuid"
	"golang.org/x/image/draw"
)

var allowedMIMEs = map[string]string{
	"image/jpeg":    ".jpg",
	"image/png":     ".png",
	"image/webp":    ".webp",
	"image/gif":     ".gif",
	"image/svg+xml": ".svg",
}

const (
	MaxUploadSize = 25 << 20 // 25 MB
)

// Service manages uploads, image conversions, resizing and cached thumbnail serving.
type Service struct {
	baseDir    string
	origDir    string
	cacheDir   string
	resizeLock sync.Mutex
}

// NewService creates a new media service with base directory (default "/data/uploads" or "pb_data/uploads").
func NewService(baseDir string) (*Service, error) {
	if baseDir == "" {
		baseDir = os.Getenv("UPLOADS_DIR")
	}
	if baseDir == "" {
		baseDir = "data/uploads"
	}

	origDir := filepath.Join(baseDir, "originals")
	cacheDir := filepath.Join(baseDir, "cache")

	if err := os.MkdirAll(origDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create originals dir: %w", err)
	}
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache dir: %w", err)
	}

	return &Service{
		baseDir:  baseDir,
		origDir:  origDir,
		cacheDir: cacheDir,
	}, nil
}

// SaveUpload validates size & MIME type and stores file into originals directory.
func (s *Service) SaveUpload(file io.Reader, header *multipart.FileHeader) (string, error) {
	if header.Size > MaxUploadSize {
		return "", fmt.Errorf("file size %d exceeds max %d bytes", header.Size, MaxUploadSize)
	}

	// Read first 512 bytes for MIME detection
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("failed to read file header: %w", err)
	}
	mimeType := http.DetectContentType(buf[:n])

	// Fallback to header content type if detection is generic octet-stream
	if mimeType == "application/octet-stream" && header.Header.Get("Content-Type") != "" {
		mimeType = header.Header.Get("Content-Type")
	}

	// Normalize svg
	if strings.Contains(strings.ToLower(header.Filename), ".svg") {
		mimeType = "image/svg+xml"
	}

	ext, ok := allowedMIMEs[mimeType]
	if !ok {
		// Also allow webp by extension if MIME detector was uncertain
		if strings.HasSuffix(strings.ToLower(header.Filename), ".webp") {
			ext = ".webp"
		} else {
			return "", fmt.Errorf("unsupported image format: %s", mimeType)
		}
	}

	rawName := strings.TrimSuffix(filepath.Base(header.Filename), filepath.Ext(header.Filename))
	cleanName := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '_'
	}, rawName)

	filename := fmt.Sprintf("%s_%s%s", uuid.New().String()[:8], cleanName, ext)
	dstPath := filepath.Join(s.origDir, filename)

	dst, err := os.Create(dstPath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	if _, err := dst.Write(buf[:n]); err != nil {
		return "", err
	}
	if _, err := io.Copy(dst, file); err != nil {
		return "", err
	}

	return filename, nil
}

// ParseDimensions parses standard size aliases or WxH string (e.g., "150x150", "thumb", "sm", "md", "lg").
func ParseDimensions(sizeStr string) (int, int) {
	switch sizeStr {
	case "thumb", "avatar":
		return 150, 150
	case "sm":
		return 320, 240
	case "md":
		return 640, 480
	case "lg":
		return 1200, 800
	case "raw", "original":
		return 0, 0
	}

	parts := strings.Split(strings.ToLower(sizeStr), "x")
	if len(parts) == 2 {
		w, _ := strconv.Atoi(parts[0])
		h, _ := strconv.Atoi(parts[1])
		if w > 0 && h > 0 {
			return w, h
		}
	}
	return 0, 0
}

// GetOrGenerateThumbnail returns path to cached thumbnail or original image.
func (s *Service) GetOrGenerateThumbnail(sizeStr, filename string) (string, error) {
	filename = filepath.Base(filename)
	origPath := filepath.Join(s.origDir, filename)

	// If original not found in originals, check public/ or ui/assets as fallback
	if _, err := os.Stat(origPath); err != nil {
		fallbackPath := filepath.Join("public", filename)
		if _, err2 := os.Stat(fallbackPath); err2 == nil {
			origPath = fallbackPath
		} else {
			return "", os.ErrNotExist
		}
	}

	width, height := ParseDimensions(sizeStr)
	if width == 0 && height == 0 {
		return origPath, nil
	}

	cachedFile := filepath.Join(s.cacheDir, sizeStr, filename)
	if _, err := os.Stat(cachedFile); err == nil {
		return cachedFile, nil
	}

	s.resizeLock.Lock()
	defer s.resizeLock.Unlock()

	// Double check under lock
	if _, err := os.Stat(cachedFile); err == nil {
		return cachedFile, nil
	}

	// Generate resized thumbnail
	f, err := os.Open(origPath)
	if err != nil {
		return origPath, nil
	}
	defer f.Close()

	srcImg, format, err := image.Decode(f)
	if err != nil {
		// If decoding failed (e.g. svg or raw webp), return original
		return origPath, nil
	}

	srcBounds := srcImg.Bounds()
	dstImg := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.BiLinear.Scale(dstImg, dstImg.Bounds(), srcImg, srcBounds, draw.Over, nil)

	if err := os.MkdirAll(filepath.Dir(cachedFile), 0755); err != nil {
		return origPath, nil
	}

	out, err := os.Create(cachedFile)
	if err != nil {
		return origPath, nil
	}
	defer out.Close()

	switch format {
	case "png":
		_ = png.Encode(out, dstImg)
	case "gif":
		_ = gif.Encode(out, dstImg, nil)
	default:
		_ = jpeg.Encode(out, dstImg, &jpeg.Options{Quality: 85})
	}

	return cachedFile, nil
}

// ServeMediaHandler returns an http.HandlerFunc that streams the requested media file with long-term cache headers.
func (s *Service) ServeMediaHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/media/"), "/")
		if len(parts) < 2 {
			http.NotFound(w, r)
			return
		}

		sizeStr := parts[0]
		filename := filepath.Base(strings.Join(parts[1:], "/"))

		filePath, err := s.GetOrGenerateThumbnail(sizeStr, filename)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		// Set long-term immutable cache headers
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		http.ServeFile(w, r, filePath)
	}
}

// UploadHandler processes multipart form upload and returns JSON response with generated paths.
func (s *Service) UploadHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(MaxUploadSize); err != nil {
			http.Error(w, "Upload too large", http.StatusBadRequest)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Missing file parameter", http.StatusBadRequest)
			return
		}
		defer file.Close()

		savedName, err := s.SaveUpload(file, header)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		resp := map[string]interface{}{
			"success":   true,
			"filename":  savedName,
			"url":       "/media/raw/" + savedName,
			"thumb_url": "/media/thumb/" + savedName,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}
}
