package ui

import (
	"io/fs"
	"testing"
)

func TestEmbeddedAssetsFS(t *testing.T) {
	fsys := EmbeddedAssetsFS()
	if fsys == nil {
		t.Fatal("expected non-nil EmbeddedAssetsFS")
	}

	// Verify styles.css can be opened
	f, err := fsys.Open("styles.css")
	if err != nil {
		t.Fatalf("failed to open embedded styles.css: %v", err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		t.Fatalf("failed to stat embedded styles.css: %v", err)
	}
	if info.Size() == 0 {
		t.Errorf("embedded styles.css is empty")
	}

	// Verify admin.css can be opened
	adminFile, err := fsys.Open("admin.css")
	if err != nil {
		t.Fatalf("failed to open embedded admin.css: %v", err)
	}
	defer adminFile.Close()

	adminInfo, err := adminFile.Stat()
	if err != nil {
		t.Fatalf("failed to stat embedded admin.css: %v", err)
	}
	if adminInfo.Size() == 0 {
		t.Errorf("embedded admin.css is empty")
	}

	// Verify images/cyber_avatar.png and cyber_avatar.webp can be opened
	imgFile, err := fsys.Open("images/cyber_avatar.png")
	if err != nil {
		t.Fatalf("failed to open embedded images/cyber_avatar.png: %v", err)
	}
	defer imgFile.Close()

	webpFile, err := fsys.Open("images/cyber_avatar.webp")
	if err != nil {
		t.Fatalf("failed to open embedded images/cyber_avatar.webp: %v", err)
	}
	defer webpFile.Close()
}

func TestAssetsFS(t *testing.T) {
	fsys := AssetsFS()
	if fsys == nil {
		t.Fatal("expected non-nil AssetsFS")
	}

	_, err := fs.Stat(fsys, "styles.css")
	if err != nil {
		t.Fatalf("failed to stat styles.css via AssetsFS: %v", err)
	}

	_, err = fs.Stat(fsys, "admin.css")
	if err != nil {
		t.Fatalf("failed to stat admin.css via AssetsFS: %v", err)
	}
}
