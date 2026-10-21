package testutil

import (
	"testing"

	"sitecraft/internal/data"

	"github.com/pocketbase/pocketbase"
)

// NewTestApp creates an isolated PocketBase instance in a temporary directory
// with initialized schemas, and automatically registers cleanup on test completion.
func NewTestApp(t *testing.T) *pocketbase.PocketBase {
	t.Helper()
	testDir := t.TempDir()
	app := pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDataDir: testDir,
	})
	if err := app.Bootstrap(); err != nil {
		t.Fatalf("Failed to bootstrap test PB: %v", err)
	}
	if err := data.InitSchema(app); err != nil {
		t.Fatalf("Failed to init schema for test PB: %v", err)
	}
	t.Cleanup(func() {
		app.ResetBootstrapState()
	})
	return app
}
