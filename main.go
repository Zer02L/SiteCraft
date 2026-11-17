package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"sitecraft/internal/auth"
	"sitecraft/internal/cache"
	"sitecraft/internal/data"
	"sitecraft/internal/db"
	"sitecraft/internal/exporter"
	"sitecraft/internal/media"
	"sitecraft/internal/notifications"
	"sitecraft/internal/server"

	"github.com/spf13/cobra"
)

func main() {
	// Initialize structured logger
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	slog.SetDefault(slog.New(handler))

	var dbDSN string
	var presetFlag string
	var themeFlag string
	var portFlag string
	var httpFlag string
	var dirFlag string
	var testConfigFlag bool

	rootCmd := &cobra.Command{
		Use:   "sitecraft",
		Short: "SiteCraft - Ultra-fast Zero-Node.js Go + Templ + SQLite/Postgres Platform",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if testConfigFlag {
				if err := runTestConfig(dbDSN, dirFlag); err != nil {
					return err
				}
				os.Exit(0)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServe(dbDSN, portFlag, httpFlag, dirFlag, presetFlag, themeFlag)
		},
	}

	rootCmd.PersistentFlags().StringVar(&dbDSN, "db", "", "Database DSN (e.g. pb_data/sitecraft.db or postgres://...)")
	rootCmd.PersistentFlags().StringVar(&presetFlag, "preset", "", "Preset name or JSON path (legal, medical, auto, beauty)")
	rootCmd.PersistentFlags().StringVar(&themeFlag, "theme", "", "Color theme override")
	rootCmd.PersistentFlags().StringVarP(&portFlag, "port", "p", ":8090", "HTTP server listen address")
	rootCmd.PersistentFlags().StringVar(&httpFlag, "http", "", "HTTP server listen address (alias for --port, e.g. 0.0.0.0:8080 or :8090)")
	rootCmd.PersistentFlags().StringVar(&dirFlag, "dir", "", "Data directory (e.g. ./pb_data or ./data)")
	rootCmd.PersistentFlags().BoolVar(&testConfigFlag, "test-config", false, "Test database connection, schema and configuration without starting server")

	// 1. serve command
	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Starts the high-performance HTTP web server and background workers",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServe(dbDSN, portFlag, httpFlag, dirFlag, presetFlag, themeFlag)
		},
	}
	serveCmd.Flags().StringVar(&httpFlag, "http", "", "HTTP server listen address")
	serveCmd.Flags().StringVar(&dirFlag, "dir", "", "Data directory")
	rootCmd.AddCommand(serveCmd)

	// test-config command
	testConfigCmd := &cobra.Command{
		Use:   "test-config",
		Short: "Validates database connection, schema integrity, and configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTestConfig(dbDSN, dirFlag)
		},
	}
	testConfigCmd.Flags().StringVar(&dirFlag, "dir", "", "Data directory")
	rootCmd.AddCommand(testConfigCmd)

	// 2. apply-preset command
	applyPresetCmd := &cobra.Command{
		Use:   "apply-preset [preset_name]",
		Short: "Loads and applies niche preset and demo data into database",
		RunE: func(cmd *cobra.Command, args []string) error {
			presetName := presetFlag
			if len(args) > 0 && args[0] != "" {
				presetName = args[0]
			}
			if presetName == "" {
				presetName = "b2b_legal"
			}
			presetPath := presetName
			if !strings.HasSuffix(presetPath, ".json") {
				presetPath = "presets/" + presetPath + ".json"
			}

			if dirFlag != "" && dbDSN == "" {
				cleanDir := filepath.Clean(dirFlag)
				_ = os.MkdirAll(cleanDir, 0755)
				dbDSN = filepath.Join(cleanDir, "sitecraft.db")
			}

			sqlDB, queries, err := db.Open(dbDSN)
			if err != nil {
				return fmt.Errorf("failed to open database: %w", err)
			}
			defer sqlDB.Close()

			ctx := context.Background()
			repo := data.NewRepository(sqlDB, queries)
			if err := repo.ApplyPreset(ctx, presetPath); err != nil {
				return fmt.Errorf("failed to apply preset: %w", err)
			}

			if themeFlag != "" {
				_, _ = sqlDB.ExecContext(ctx, "UPDATE site_settings SET theme = ?", themeFlag)
			}

			slog.Info("Successfully applied preset", "path", presetPath, "theme", themeFlag)
			return nil
		},
	}
	applyPresetCmd.Flags().StringVar(&dirFlag, "dir", "", "Data directory")
	rootCmd.AddCommand(applyPresetCmd)

	// 3. export command
	var outDirFlag string
	exportCmd := &cobra.Command{
		Use:   "export",
		Short: "Exports the website into flat static HTML/CSS files for GitHub Pages, S3 or Cloudflare",
		RunE: func(cmd *cobra.Command, args []string) error {
			if outDirFlag == "" {
				outDirFlag = "./dist"
			}

			if dirFlag != "" && dbDSN == "" {
				cleanDir := filepath.Clean(dirFlag)
				_ = os.MkdirAll(cleanDir, 0755)
				dbDSN = filepath.Join(cleanDir, "sitecraft.db")
			}

			sqlDB, queries, err := db.Open(dbDSN)
			if err != nil {
				return fmt.Errorf("failed to open database: %w", err)
			}
			defer sqlDB.Close()

			ctx := context.Background()
			repo := data.NewRepository(sqlDB, queries)

			if presetFlag != "" {
				presetPath := presetFlag
				if !strings.HasSuffix(presetPath, ".json") {
					presetPath = "presets/" + presetPath + ".json"
				}
				_ = repo.ApplyPreset(ctx, presetPath)
			}

			if err := exporter.ExportStaticSite(repo, outDirFlag); err != nil {
				return fmt.Errorf("export failed: %w", err)
			}

			slog.Info("Successfully exported static site", "outDir", outDirFlag)
			return nil
		},
	}
	exportCmd.Flags().StringVar(&outDirFlag, "out", "./dist", "Output directory for exported files")
	exportCmd.Flags().StringVar(&dirFlag, "dir", "", "Data directory")
	rootCmd.AddCommand(exportCmd)

	if err := rootCmd.Execute(); err != nil {
		slog.Error("CLI command failed", "error", err)
		os.Exit(1)
	}
}

func runServe(dbDSN, port, httpFlag, dirFlag, presetFlag, themeFlag string) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Prioritize --http if provided (e.g. 0.0.0.0:8080, 127.0.0.1:8090, :8090, 8080)
	if httpFlag != "" {
		port = httpFlag
	}
	if !strings.Contains(port, ":") {
		port = ":" + port
	}

	// Environment variable overrides
	if envPort := os.Getenv("PORT"); envPort != "" {
		if !strings.HasPrefix(envPort, ":") {
			port = ":" + envPort
		} else {
			port = envPort
		}
	}

	// If --dir is specified and dbDSN is empty, use dir
	if dirFlag != "" && dbDSN == "" {
		cleanDir := filepath.Clean(dirFlag)
		_ = os.MkdirAll(cleanDir, 0755)
		dbDSN = filepath.Join(cleanDir, "sitecraft.db")
	}

	// 1. Database & Migrations
	sqlDB, queries, err := db.Open(dbDSN)
	if err != nil {
		return fmt.Errorf("database initialization error: %w", err)
	}
	defer sqlDB.Close()

	repo := data.NewRepository(sqlDB, queries)

	// Apply preset if requested
	if presetFlag != "" {
		presetPath := presetFlag
		if !strings.HasSuffix(presetPath, ".json") {
			presetPath = "presets/" + presetPath + ".json"
		}
		if err := repo.ApplyPreset(ctx, presetPath); err != nil {
			slog.Warn("Failed to apply preset", "error", err)
		} else if themeFlag != "" {
			_, _ = sqlDB.ExecContext(ctx, "UPDATE site_settings SET theme = ?", themeFlag)
		}
	}

	// 2. Media Service
	uploadsDir := "data/uploads"
	if dirFlag != "" {
		uploadsDir = filepath.Join(dirFlag, "uploads")
	}
	mediaService, err := media.NewService(uploadsDir)
	if err != nil {
		slog.Warn("Media service init warning", "error", err)
	}

	// 3. Session Manager
	isPostgres := strings.HasPrefix(dbDSN, "postgres://") || strings.HasPrefix(dbDSN, "postgresql://")
	sessionManager := auth.NewSessionManager(sqlDB, isPostgres)

	// 4. Background SIGHUP Listener for zero-downtime hot-reload (systemctl reload / kill -HUP)
	hupChan := make(chan os.Signal, 1)
	signal.Notify(hupChan, syscall.SIGHUP)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-hupChan:
				slog.Info("Received SIGHUP signal: reloading configuration, clearing in-memory caches...")
				cache.GetStore().InvalidateAll()
				if refreshedSettings, err := repo.GetSiteSettings(context.Background()); err == nil {
					cache.GetStore().SetSettings(refreshedSettings)
					slog.Info("Configuration hot-reloaded successfully without downtime", "theme", refreshedSettings.Theme)
				}
			}
		}
	}()

	// 5. Background Cron Workers with Context-aware Termination
	go func() {
		reminderTicker := time.NewTicker(10 * time.Minute)
		scheduleTicker := time.NewTicker(1 * time.Minute)
		defer reminderTicker.Stop()
		defer scheduleTicker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-reminderTicker.C:
				notifications.CheckBookingReminders(ctx, repo)
			case <-scheduleTicker.C:
				notifications.CheckArticlesSchedule(repo)
			}
		}
	}()

	// 6. HTTP Web Server with Graceful Shutdown
	srv, err := server.New(sqlDB, queries, mediaService, sessionManager)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	return srv.StartWithContext(ctx, port)
}

// runTestConfig verifies database connectivity, schema migrations, and settings without running the web server.
func runTestConfig(dbDSN, dirFlag string) error {
	if dirFlag != "" && dbDSN == "" {
		cleanDir := filepath.Clean(dirFlag)
		dbDSN = filepath.Join(cleanDir, "sitecraft.db")
	}
	if dbDSN == "" {
		dbDSN = os.Getenv("DATABASE_URL")
	}
	if dbDSN == "" {
		dbDSN = "pb_data/sitecraft.db"
	}

	fmt.Println("==================================================")
	fmt.Println("🔍 SiteCraft Pre-Flight & Configuration Test")
	fmt.Println("==================================================")
	fmt.Printf("• Target Database DSN: %s\n", dbDSN)

	// 1. Open DB & verify migrations
	sqlDB, queries, err := db.Open(dbDSN)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ [FAIL] Database connection or migrations failed: %v\n", err)
		return fmt.Errorf("database initialization failed: %w", err)
	}
	defer sqlDB.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 2. Ping
	if err := sqlDB.PingContext(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "❌ [FAIL] Database ping failed: %v\n", err)
		return fmt.Errorf("ping failed: %w", err)
	}
	fmt.Println("✓ [OK] Database connection and Ping successful")
	fmt.Println("✓ [OK] Goose migrations verified and up to date")

	// 3. SQLite-specific integrity checks
	isPostgres := strings.HasPrefix(dbDSN, "postgres://") || strings.HasPrefix(dbDSN, "postgresql://")
	if !isPostgres {
		var checkResult string
		if err := sqlDB.QueryRowContext(ctx, "PRAGMA quick_check;").Scan(&checkResult); err != nil {
			fmt.Fprintf(os.Stderr, "❌ [FAIL] SQLite quick_check failed: %v\n", err)
			return fmt.Errorf("sqlite quick_check failed: %w", err)
		}
		if checkResult != "ok" {
			fmt.Fprintf(os.Stderr, "❌ [FAIL] SQLite integrity check error: %s\n", checkResult)
			return fmt.Errorf("sqlite integrity corrupted: %s", checkResult)
		}

		var journalMode string
		_ = sqlDB.QueryRowContext(ctx, "PRAGMA journal_mode;").Scan(&journalMode)
		fmt.Printf("✓ [OK] SQLite schema integrity verified (journal_mode=%s)\n", journalMode)
	}

	// 4. Site Settings & Repository validation
	repo := data.NewRepository(sqlDB, queries)
	settings, err := repo.GetSiteSettings(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ [FAIL] Querying site_settings failed: %v\n", err)
		return fmt.Errorf("site_settings query failed: %w", err)
	}
	title := settings.SiteName
	if title == "" {
		title = "(Default / Empty)"
	}
	fmt.Printf("✓ [OK] Core site settings verified (site_name: %s, theme: %s)\n", title, settings.Theme)

	fmt.Println("==================================================")
	fmt.Println("✅ [PASS] All configuration and pre-flight checks passed!")
	fmt.Println("==================================================")
	return nil
}
