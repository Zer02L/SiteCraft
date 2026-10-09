package exporter

import (
	"os"
	"path/filepath"
	"testing"

	"sitecraft/ui"
)

func TestCopyFSFromEmbedded(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "export-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	dst := filepath.Join(tempDir, "assets")
	if err := copyFS(ui.AssetsFS(), dst); err != nil {
		t.Fatalf("copyFS failed: %v", err)
	}

	// Verify styles.css exists in exported assets
	cssPath := filepath.Join(dst, "styles.css")
	info, err := os.Stat(cssPath)
	if err != nil {
		t.Fatalf("expected styles.css in %s, got err: %v", dst, err)
	}
	if info.Size() == 0 {
		t.Errorf("exported styles.css is empty")
	}
}
