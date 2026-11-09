package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"sitecraft/internal/db"

	_ "modernc.org/sqlite"
)

func main() {
	sourceDB := flag.String("source", "pb_data/data.db", "Path to source PocketBase SQLite database")
	targetDB := flag.String("target", "pb_data/sitecraft.db", "Path to target SiteCraft SQLite database")
	sourceStorage := flag.String("storage", "pb_data/storage", "Path to source PocketBase storage directory")
	targetUploads := flag.String("uploads", "data/uploads/originals", "Path to target media uploads directory")
	flag.Parse()

	log.Printf("[Migration] Starting migration from PocketBase (%s) to Chi/SQLite (%s)...\n", *sourceDB, *targetDB)

	if _, err := os.Stat(*sourceDB); os.IsNotExist(err) {
		log.Printf("[Migration Warning] Source database %s does not exist. Skipping DB record migration.\n", *sourceDB)
	} else {
		if err := migrateDatabase(*sourceDB, *targetDB); err != nil {
			log.Fatalf("[Migration Error] Database migration failed: %v\n", err)
		}
	}

	// 2. Migrate files
	if _, err := os.Stat(*sourceStorage); os.IsNotExist(err) {
		log.Printf("[Migration Info] Storage directory %s does not exist. Skipping file migration.\n", *sourceStorage)
	} else {
		count, err := migrateFiles(*sourceStorage, *targetUploads)
		if err != nil {
			log.Printf("[Migration Warning] File migration warning: %v\n", err)
		} else {
			log.Printf("[Migration] Successfully migrated %d files to %s\n", count, *targetUploads)
		}
	}

	log.Println("[Migration] PocketBase to Chi/SQLite migration completed successfully!")
}

func migrateDatabase(sourcePath, targetPath string) error {
	src, err := sql.Open("sqlite", sourcePath)
	if err != nil {
		return fmt.Errorf("failed to open source db: %w", err)
	}
	defer src.Close()

	// Initialize target db and run migrations
	dst, queries, err := db.Open(targetPath)
	if err != nil {
		return fmt.Errorf("failed to open target db: %w", err)
	}
	defer dst.Close()
	_ = queries

	// Migrate site_settings
	migrateTable(src, dst, "site_settings")
	// Migrate site_content
	migrateTable(src, dst, "site_content")
	// Migrate page_blocks
	migrateTable(src, dst, "page_blocks")
	// Migrate services
	migrateTable(src, dst, "services")
	// Migrate leads
	migrateTable(src, dst, "leads")
	// Migrate articles
	migrateTable(src, dst, "articles")
	// Migrate service_orders
	migrateTable(src, dst, "service_orders")
	// Migrate coupons
	migrateTable(src, dst, "coupons")
	// Migrate resources
	migrateTable(src, dst, "resources")
	// Migrate branches
	migrateTable(src, dst, "branches")
	// Migrate reviews
	migrateTable(src, dst, "reviews")
	// Migrate vacancies
	migrateTable(src, dst, "vacancies")
	// Migrate chat_messages
	migrateTable(src, dst, "chat_messages")

	return nil
}

func migrateTable(src, dst *sql.DB, tableName string) {
	// Check if source table exists
	var count int
	err := src.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", tableName).Scan(&count)
	if err != nil || count == 0 {
		return
	}

	rows, err := src.Query("SELECT * FROM " + tableName)
	if err != nil {
		log.Printf("[Migration] Could not query source table %s: %v\n", tableName, err)
		return
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return
	}

	// Fetch destination table columns
	dstColsRows, err := dst.Query("PRAGMA table_info(" + tableName + ")")
	if err != nil {
		return
	}
	defer dstColsRows.Close()

	dstColsMap := make(map[string]bool)
	for dstColsRows.Next() {
		var cid int
		var name, colType string
		var notNull, pk int
		var dfltValue sql.NullString
		if err := dstColsRows.Scan(&cid, &name, &colType, &notNull, &dfltValue, &pk); err == nil {
			dstColsMap[name] = true
		}
	}

	// Common columns
	var validCols []string
	var colIndices []int
	for i, col := range cols {
		if dstColsMap[col] {
			validCols = append(validCols, col)
			colIndices = append(colIndices, i)
		}
	}

	if len(validCols) == 0 {
		return
	}

	placeholders := make([]string, len(validCols))
	for i := range placeholders {
		placeholders[i] = "?"
	}

	insertSQL := fmt.Sprintf(
		"INSERT OR REPLACE INTO %s (%s) VALUES (%s)",
		tableName,
		strings.Join(validCols, ", "),
		strings.Join(placeholders, ", "),
	)

	inserted := 0
	for rows.Next() {
		rawVals := make([]interface{}, len(cols))
		scanArgs := make([]interface{}, len(cols))
		for i := range rawVals {
			scanArgs[i] = &rawVals[i]
		}
		if err := rows.Scan(scanArgs...); err != nil {
			continue
		}

		insertVals := make([]interface{}, len(validCols))
		for i, idx := range colIndices {
			insertVals[i] = rawVals[idx]
		}

		if _, err := dst.Exec(insertSQL, insertVals...); err == nil {
			inserted++
		}
	}

	log.Printf("[Migration] Table %s: migrated %d rows\n", tableName, inserted)
}

func migrateFiles(srcDir, dstDir string) (int, error) {
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return 0, err
	}

	count := 0
	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil || info.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return nil
		}

		// Flatten directory structure or preserve file names
		dstFile := filepath.Join(dstDir, filepath.Base(rel))
		if err := copyFile(path, dstFile); err == nil {
			count++
		}
		return nil
	})

	return count, err
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
