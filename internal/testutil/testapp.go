package testutil

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"testing"

	"sitecraft/internal/data"
	"sitecraft/internal/db"
)

// NewTestRepo creates an isolated SQLite DB in a temp directory with all migrations applied.
func NewTestRepo(t *testing.T) (*data.Repository, *sql.DB, *db.Queries) {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	sqlDB, queries, err := db.Open(fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)", dbPath))
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})
	repo := data.NewRepository(sqlDB, queries)
	return repo, sqlDB, queries
}
