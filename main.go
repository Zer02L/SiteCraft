package main

import (
	"log"
	"os"
	"strings"

	"sitecraft/internal/cache"
	"sitecraft/internal/data"
	"sitecraft/internal/exporter"
	"sitecraft/internal/handlers"
	"sitecraft/internal/notifications"

	"ganigeorgiev.com/pocketbase/pocketbase"
	"ganigeorgiev.com/pocketbase/pocketbase/core"
	"github.com/spf13/cobra"
)

func main() {
	// Инициализация структурированного логирования slog (JSON в проде, текст в dev)
	handlers.InitLogger()

	app := pocketbase.New()

	var presetFlag string
	var themeFlag string

	// Флаги на корневую команду
	app.RootCmd.PersistentFlags().StringVar(&presetFlag, "preset", "", "Preset name or JSON path (legal, medical, auto, beauty)")
	app.RootCmd.PersistentFlags().StringVar(&themeFlag, "theme", "", "Color theme override (legal, medical, auto, beauty, tech, dark)")

	// Команда для явного применения пресета в базу данных: ./app apply-preset <name> [--theme=...]
	applyPresetCmd := &cobra.Command{
		Use:   "apply-preset [preset_name]",
		Short: "Loads and applies niche preset and demo data into PocketBase",
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
			if err := app.Bootstrap(); err != nil {
				return err
			}
			if err := data.InitSchema(app); err != nil {
				return err
			}
			if err := data.ApplyPreset(app, presetPath); err != nil {
				return err
			}
			if themeFlag != "" {
				if settingsCol, err := app.FindCollectionByNameOrId("site_settings"); err == nil {
					rec, _ := app.FindFirstRecordByFilter("site_settings", "")
					if rec == nil {
						rec = core.NewRecord(settingsCol)
					}
					rec.Set("theme", themeFlag)
					_ = app.Save(rec)
				}
			}
			log.Printf("[Preset] Successfully applied %s (theme: %s)\n", presetPath, themeFlag)
			return nil
		},
	}
	applyPresetCmd.Flags().StringVar(&themeFlag, "theme", "", "Color theme override")
	app.RootCmd.AddCommand(applyPresetCmd)

	// Команда статического экспорта: ./app export [--preset=...] [--out=./dist]
	var outDirFlag string
	exportCmd := &cobra.Command{
		Use:   "export",
		Short: "Exports the website into flat static HTML/CSS files for GitHub Pages, S3 or Cloudflare",
		RunE: func(cmd *cobra.Command, args []string) error {
			if outDirFlag == "" {
				outDirFlag = "./dist"
			}
			if err := app.Bootstrap(); err != nil {
				return err
			}
			if err := data.InitSchema(app); err != nil {
				return err
			}
			if presetFlag != "" {
				presetPath := presetFlag
				if !strings.HasSuffix(presetPath, ".json") {
					presetPath = "presets/" + presetPath + ".json"
				}
				_ = data.ApplyPreset(app, presetPath)
			}
			if err := exporter.ExportStaticSite(app, outDirFlag); err != nil {
				return err
			}
			log.Printf("[Export] Сайт успешно экспортирован в статический вид: %s\n", outDirFlag)
			return nil
		},
	}
	exportCmd.Flags().StringVar(&outDirFlag, "out", "./dist", "Output directory for exported files")
	app.RootCmd.AddCommand(exportCmd)

	// Инициализация структуры коллекций и начальных данных
	app.OnBootstrap().BindFunc(func(e *core.BootstrapEvent) error {
		if err := e.Next(); err != nil {
			return err
		}
		if err := data.InitSchema(e.App); err != nil {
			log.Printf("[Bootstrap Warning] Не удалось инициализировать схему БД: %v\n", err)
		}
		return nil
	})

	// Настройка веб-сервера и применение пресета при старте serve (если передан флаг)
	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		// Извлекаем флаги --preset и --theme даже если они переданы после `serve`
		presetVal := presetFlag
		themeVal := themeFlag
		for i, arg := range os.Args {
			if strings.HasPrefix(arg, "--preset=") {
				presetVal = strings.TrimPrefix(arg, "--preset=")
			} else if arg == "--preset" && i+1 < len(os.Args) {
				presetVal = os.Args[i+1]
			}
			if strings.HasPrefix(arg, "--theme=") {
				themeVal = strings.TrimPrefix(arg, "--theme=")
			} else if arg == "--theme" && i+1 < len(os.Args) {
				themeVal = os.Args[i+1]
			}
		}

		if presetVal != "" {
			presetPath := presetVal
			if !strings.HasSuffix(presetPath, ".json") {
				presetPath = "presets/" + presetPath + ".json"
			}
			if err := data.ApplyPreset(app, presetPath); err != nil {
				log.Printf("[Preset Warning] Не удалось применить пресет: %v\n", err)
			} else if themeVal != "" {
				if _, err := app.FindCollectionByNameOrId("site_settings"); err == nil {
					rec, _ := app.FindFirstRecordByFilter("site_settings", "")
					if rec != nil {
						rec.Set("theme", themeVal)
						_ = app.Save(rec)
					}
				}
			}
		}

		// Регулярная проверка отложенных публикаций блога (инвалидация кэша статей)
		app.Cron().MustAdd("check_scheduled_articles", "* * * * *", func() {
			cache.GetStore().InvalidateArticles()
		})

		// Модуль #12: Регулярная проверка напоминаний о визитах за 2 часа (раз в 10 минут)
		app.Cron().MustAdd("check_booking_reminders", "*/10 * * * *", func() {
			notifications.CheckBookingReminders(app)
		})

		handlers.RegisterRoutes(app, se.Router)
		return se.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
