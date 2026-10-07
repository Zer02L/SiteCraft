package exporter

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"sitecraft/internal/data"
	"sitecraft/ui/pages"

	"github.com/pocketbase/pocketbase/core"
)

// ExportStaticSite renders the landing page and copies static assets into outputDir
func ExportStaticSite(app core.App, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	settings := data.GetSiteSettings(app)
	content := data.GetSiteContent(app)
	services := data.GetServices(app)
	bookedSlots := data.GetBookedSlots(app)
	blocks := data.GetPageBlocks(app)

	// 1. Render index.html via Templ
	var buf bytes.Buffer
	component := pages.Home(settings, content, services, bookedSlots, blocks)
	if err := component.Render(context.Background(), &buf); err != nil {
		return fmt.Errorf("failed to render home page template: %w", err)
	}

	indexPath := filepath.Join(outputDir, "index.html")
	if err := os.WriteFile(indexPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write index.html: %w", err)
	}

	// 2. Render Blog and articles if enabled
	articles := data.GetArticles(app)
	if len(articles) > 0 {
		blogDir := filepath.Join(outputDir, "blog")
		_ = os.MkdirAll(blogDir, 0755)

		var blogBuf bytes.Buffer
		if err := pages.BlogList(settings, content, articles, "").Render(context.Background(), &blogBuf); err == nil {
			_ = os.WriteFile(filepath.Join(blogDir, "index.html"), blogBuf.Bytes(), 0644)
		}

		for _, a := range articles {
			artDir := filepath.Join(blogDir, a.Slug)
			_ = os.MkdirAll(artDir, 0755)
			var artBuf bytes.Buffer
			if err := pages.BlogPost(settings, content, a, articles).Render(context.Background(), &artBuf); err == nil {
				_ = os.WriteFile(filepath.Join(artDir, "index.html"), artBuf.Bytes(), 0644)
			}
		}
	}

	// 3. Copy static assets (ui/assets -> outputDir/assets)
	assetsSrc := "ui/assets"
	assetsDst := filepath.Join(outputDir, "assets")
	if _, err := os.Stat(assetsSrc); err == nil {
		if err := copyDir(assetsSrc, assetsDst); err != nil {
			return fmt.Errorf("failed to copy assets: %w", err)
		}
	}

	return nil
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}

		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		dstFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
		if err != nil {
			return err
		}
		defer dstFile.Close()

		_, err = io.Copy(dstFile, srcFile)
		return err
	})
}
