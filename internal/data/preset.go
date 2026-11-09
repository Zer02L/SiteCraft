package data

import (
	"os"
	"path/filepath"
	"strings"

	"sitecraft/internal/models"
)

type PresetData struct {
	Settings models.SiteSettings  `json:"settings"`
	Content  models.SiteContent   `json:"content"`
	Services []models.ServiceItem `json:"services"`
	Blocks   []models.PageBlock   `json:"blocks"`
}

// resolvePresetPath checks if a preset file exists directly or under b2b_ / b2c_ prefixes
func resolvePresetPath(p string) string {
	if _, err := os.Stat(p); err == nil {
		return p
	}
	dir := filepath.Dir(p)
	base := filepath.Base(p)
	baseNoExt := strings.TrimSuffix(base, ".json")
	baseUnder := strings.ReplaceAll(baseNoExt, "-", "_")

	candidates := []string{
		p,
		filepath.Join("..", "..", p),
		filepath.Join("..", p),
		filepath.Join(dir, "b2b_"+baseNoExt+".json"),
		filepath.Join(dir, "b2c_"+baseNoExt+".json"),
		filepath.Join(dir, "b2b_"+baseUnder+".json"),
		filepath.Join(dir, "b2c_"+baseUnder+".json"),
		filepath.Join("..", "..", dir, "b2b_"+baseNoExt+".json"),
		filepath.Join("..", "..", dir, "b2c_"+baseNoExt+".json"),
	}

	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return p
}
