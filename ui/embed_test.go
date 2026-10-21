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

	// Verify images/cyber_avatar.png can be opened
	imgFile, err := fsys.Open("images/cyber_avatar.png")
	if err != nil {
		t.Fatalf("failed to open embedded images/cyber_avatar.png: %v", err)
	}
	defer imgFile.Close()
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
}
