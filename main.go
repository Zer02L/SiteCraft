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

	rootCmd := &cobra.Command{
		Use:   "sitecraft",
		Short: "SiteCraft - Ultra-fast Zero-Node.js Go + Templ + SQLite/Postgres Platform",
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

	// 4. Background Cron Workers with Context-aware Termination
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

	// 5. HTTP Web Server with Graceful Shutdown
	srv, err := server.New(sqlDB, queries, mediaService, sessionManager)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	return srv.StartWithContext(ctx, port)
}
