package pages

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestAdminThemeToggleUsesCircularReveal(t *testing.T) {
	var buf bytes.Buffer
	if err := AdminThemeToggle().Render(context.Background(), &buf); err != nil {
		t.Fatalf("render AdminThemeToggle: %v", err)
	}
	out := buf.String()
	for _, want := range []string{
		"startViewTransition",
		"clipPath",
		"::view-transition-new(root)",
		"prefers-reduced-motion",
		"theme-freeze",
		"theme-crossfade",
		"transition.finished",
		`toggle($event)`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("AdminThemeToggle markup missing %q", want)
		}
	}
}
