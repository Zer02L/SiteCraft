package db

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"sitecraft/sql/migrations"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

// Config holds DB connection settings
type Config struct {
	DSN          string
	MaxOpenConns int
	MaxIdleConns int
}

// Open initializes the database connection, configures connection pool,
// executes WAL and busy_timeout pragmas (for SQLite), and runs migrations.
func Open(dsn string) (*sql.DB, *Queries, error) {
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" {
		dsn = "pb_data/sitecraft.db"
	}

	isPostgres := strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://")
	var db *sql.DB
	var err error

	if isPostgres {
		db, err = sql.Open("pgx", dsn)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to open postgres database: %w", err)
		}

		maxConns := max(10, runtime.NumCPU()*2)
		db.SetMaxOpenConns(maxConns)
		db.SetMaxIdleConns(maxConns / 2)
		db.SetConnMaxLifetime(1 * time.Hour)
		db.SetConnMaxIdleTime(15 * time.Minute)

		if err := db.Ping(); err != nil {
			return nil, nil, fmt.Errorf("failed to ping postgres: %w", err)
		}

		goose.SetBaseFS(migrations.PostgresFS)
		if err := goose.SetDialect("postgres"); err != nil {
			return nil, nil, fmt.Errorf("failed to set goose postgres dialect: %w", err)
		}
		if err := goose.Up(db, "postgres"); err != nil {
			return nil, nil, fmt.Errorf("failed to run postgres migrations: %w", err)
		}
		slog.Info("Postgres connected and migrations applied successfully")
	} else {
		// SQLite driver (modernc.org/sqlite - CGO=0)
		dbPath := dsn
		if idx := strings.Index(dbPath, "?"); idx != -1 {
			dbPath = dbPath[:idx]
		}
		if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil && filepath.Dir(dbPath) != "." {
			return nil, nil, fmt.Errorf("failed to create db directory: %w", err)
		}

		pragmas := []string{
			"_pragma=busy_timeout(5000)",
			"_pragma=journal_mode(WAL)",
			"_pragma=synchronous(NORMAL)",
			"_pragma=foreign_keys(1)",
		}
		connStr := dsn
		if !strings.Contains(connStr, "?") {
			connStr += "?" + strings.Join(pragmas, "&")
		}

		db, err = sql.Open("sqlite", connStr)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to open sqlite database: %w", err)
		}

		// Pool configuration for SQLite WAL mode:
		// WAL supports concurrent readers with serialized writers.
		maxConns := max(4, runtime.NumCPU())
		db.SetMaxOpenConns(maxConns)
		db.SetMaxIdleConns(max(2, maxConns/2))
		db.SetConnMaxLifetime(1 * time.Hour)

		if err := db.Ping(); err != nil {
			return nil, nil, fmt.Errorf("failed to ping sqlite: %w", err)
		}

		goose.SetBaseFS(migrations.SQLiteFS)
		if err := goose.SetDialect("sqlite3"); err != nil {
			return nil, nil, fmt.Errorf("failed to set goose sqlite dialect: %w", err)
		}
		if err := goose.Up(db, "sqlite"); err != nil {
			return nil, nil, fmt.Errorf("failed to run sqlite migrations: %w", err)
		}
		slog.Info("SQLite (WAL mode) connected and migrations applied successfully", "path", dbPath)
	}

	queries := New(db)
	return db, queries, nil
}
