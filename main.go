package main

import (
	"fmt"
	"html"
	"log"
	"net/http"
	"os"
	"strings"

	"sitecraft/internal/data"
	"sitecraft/internal/handlers"
	"sitecraft/internal/models"
	"sitecraft/internal/modules"
	"sitecraft/internal/notifications"
	"sitecraft/ui/components"
	"sitecraft/ui/pages"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

func main() {
	app := pocketbase.New()

	var presetFlag string
	app.RootCmd.Flags().StringVar(&presetFlag, "preset", "", "Path to preset JSON file or preset name (legal, medical, auto, beauty)")

	// Инициализация структуры коллекций, сидов и инвалидации кэша
	app.OnBootstrap().BindFunc(func(e *core.BootstrapEvent) error {
		if err := e.Next(); err != nil {
			return err
		}
		if err := data.InitSchema(e.App); err != nil {
			log.Printf("[Bootstrap Warning] Не удалось инициализировать схему БД: %v\n", err)
		}
		if presetFlag != "" {
			presetPath := presetFlag
			if !strings.HasSuffix(presetPath, ".json") {
				presetPath = "presets/" + presetPath + ".json"
			}
			if err := data.ApplyPreset(e.App, presetPath); err != nil {
				log.Printf("[Preset Warning] Не удалось применить пресет: %v\n", err)
			}
		}
		return nil
	})

	// Настройка маршрутов веб-сайта
	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		// Раздача скомпилированных стилей и ассетов
		se.Router.GET("/assets/{path...}", apis.Static(os.DirFS("./ui/assets"), false))

		// Главная страница (SSR с помощью Templ + In-Memory Cache)
		se.Router.GET("/", func(e *core.RequestEvent) error {
			settings := data.GetSiteSettings(app)
			content := data.GetSiteContent(app)
			services := data.GetServices(app)
			bookedSlots := data.GetBookedSlots(app)
			blocks := data.GetPageBlocks(app)

			e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
			return pages.Home(settings, content, services, bookedSlots, blocks).Render(e.Request.Context(), e.Response)
		})

		// API: Ленивая подгрузка карты через HTMX (Zero-JS Lazy Map)
		se.Router.GET("/api/map-frame", func(e *core.RequestEvent) error {
			content := data.GetSiteContent(app)
			e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
			return components.MapFrame(content.MapIframe, "Интерактивная карта — "+content.HeroTitle).Render(e.Request.Context(), e.Response)
		})

		// API: Модальное окно онлайн-оплаты
		se.Router.GET("/api/payment/modal", func(e *core.RequestEvent) error {
			settings := data.GetSiteSettings(app)
			serviceID := e.Request.URL.Query().Get("service_id")
			var selectedService models.ServiceItem
			services := data.GetServices(app)
			for _, s := range services {
				if s.ID == serviceID {
					selectedService = s
					break
				}
			}
			if selectedService.ID == "" && len(services) > 0 {
				selectedService = services[0]
			}
			e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
			return components.PaymentModal(settings, selectedService).Render(e.Request.Context(), e.Response)
		})

		// API: Инициация онлайн-оплаты (ЮKassa / Mock)
		se.Router.POST("/api/payment/create", func(e *core.RequestEvent) error {
			settings := data.GetSiteSettings(app)
			description := e.Request.FormValue("description")
			email := e.Request.FormValue("email")
			phone := e.Request.FormValue("phone")
			amount := 1000.0
			_, _ = fmt.Sscanf(e.Request.FormValue("amount"), "%f", &amount)
			if amount <= 0 {
				amount = 1000.0
			}

			provider := modules.NewPaymentProvider(settings)
			res, err := provider.CreatePayment(models.PaymentRequest{
				Amount:      amount,
				Currency:    "RUB",
				Description: description,
				ReturnURL:   "/payment/success",
				Email:       email,
				Phone:       phone,
			})
			if err != nil {
				return e.HTML(http.StatusInternalServerError, fmt.Sprintf(`<div class="p-3 bg-rose-50 text-rose-700 text-xs rounded-lg">Ошибка создания платежа: %s</div>`, html.EscapeString(err.Error())))
			}

			e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
			return components.PaymentSuccessMessage(res).Render(e.Request.Context(), e.Response)
		})

		// API: Получение занятых слотов (JSON)
		se.Router.GET("/api/slots", func(e *core.RequestEvent) error {
			bookedSlots := data.GetBookedSlots(app)
			return e.JSON(http.StatusOK, bookedSlots)
		})

		// Приём заявок с формы с защитой honeypot и уведомлениями (HTMX endpoint)
		se.Router.POST("/api/leads", func(e *core.RequestEvent) error {
			settings := data.GetSiteSettings(app)

			name := strings.TrimSpace(e.Request.FormValue("name"))
			phone := strings.TrimSpace(e.Request.FormValue("phone"))
			bookingDate := strings.TrimSpace(e.Request.FormValue("booking_date"))
			bookingTime := strings.TrimSpace(e.Request.FormValue("booking_time"))
			actionType := strings.TrimSpace(e.Request.FormValue("action_type"))
			comment := strings.TrimSpace(e.Request.FormValue("comment"))
			consent := e.Request.FormValue("consent")
			honeypot := e.Request.FormValue("company_website_url")

			input := handlers.LeadFormInput{
				Name:        name,
				Phone:       phone,
				BookingDate: bookingDate,
				BookingTime: bookingTime,
				ActionType:  actionType,
				Comment:     comment,
				Consent:     consent,
				Honeypot:    honeypot,
				Mode:        settings.BookingMode,
			}

			validationErrors := handlers.ValidateLeadForm(input)

			// Если есть ошибки валидации — возвращаем структурированное сообщение
			if len(validationErrors) > 0 {
				var errList strings.Builder
				for _, msg := range validationErrors {
					errList.WriteString(fmt.Sprintf(
						`<li class="flex items-start gap-2"><svg class="w-4 h-4 text-rose-500 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path></svg><span>%s</span></li>`,
						html.EscapeString(msg),
					))
				}

				errorHTML := fmt.Sprintf(`
					<div class="p-4 bg-rose-50 border border-rose-300 rounded-xl text-rose-900 text-sm shadow-sm transition-all duration-200">
						<div class="font-bold flex items-center gap-2 mb-2 text-rose-800">
							<svg class="w-5 h-5 text-rose-600 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path></svg>
							Пожалуйста, проверьте данные формы:
						</div>
						<ul class="text-xs text-rose-700 space-y-1.5 ml-1">
							%s
						</ul>
					</div>
				`, errList.String())
				return e.HTML(http.StatusUnprocessableEntity, errorHTML)
			}

			// Проверка на повторное бронирование слота (только если указаны дата и время)
			if bookingDate != "" && bookingTime != "" {
				existingRecords, _ := app.FindRecordsByFilter(
					"leads",
					"booking_date = {:date} && booking_time = {:time} && status != 'отменена'",
					"",
					1,
					0,
					dbx.Params{"date": bookingDate, "time": bookingTime},
				)
				if len(existingRecords) > 0 {
					return e.HTML(http.StatusConflict, `
						<div class="p-4 bg-amber-50 border border-amber-300 rounded-xl text-amber-900 text-sm shadow-sm">
							<div class="font-bold flex items-center gap-2 mb-1 text-amber-800">
								<svg class="w-5 h-5 text-amber-600 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path></svg>
								Выбранное время уже занято
							</div>
							<p class="text-xs text-amber-700">К сожалению, этот слот только что забронировали. Пожалуйста, выберите другое свободное время.</p>
						</div>
					`)
				}
			}

			// Подготовка JSON payload для гибких полей
			payloadMap := map[string]interface{}{}
			if actionType != "" {
				payloadMap["action_type"] = actionType
			}
			if comment != "" {
				payloadMap["comment"] = comment
			}

			// Сохраняем в коллекцию leads в PocketBase
			leadModel := models.Lead{
				Name:        name,
				Phone:       phone,
				Status:      "новая",
				Source:      "website",
				BookingDate: bookingDate,
				BookingTime: bookingTime,
				Payload:     payloadMap,
			}

			if leadsCol, err := app.FindCollectionByNameOrId("leads"); err == nil {
				leadRecord := core.NewRecord(leadsCol)
				leadRecord.Set("name", name)
				leadRecord.Set("phone", phone)
				leadRecord.Set("booking_date", bookingDate)
				leadRecord.Set("booking_time", bookingTime)
				leadRecord.Set("status", "новая")
				leadRecord.Set("source", "website")
				leadRecord.Set("payload", payloadMap)

				if err := app.Save(leadRecord); err != nil {
					log.Printf("[Leads Error] Не удалось сохранить заявку в базу: %v\n", err)
					return e.HTML(http.StatusInternalServerError, `
						<div class="p-4 bg-rose-50 border border-rose-300 rounded-xl text-rose-800 text-sm">
							<div class="font-bold">Ошибка сервера</div>
							<p class="mt-1 text-xs">Не удалось сохранить запись. Пожалуйста, повторите попытку или позвоните нам по телефону.</p>
						</div>
					`)
				}
				leadModel.ID = leadRecord.Id
			}

			// Отправка уведомлений (Telegram Bot, CRM Webhook)
			notifications.SendLeadNotification(settings, leadModel)

			log.Printf("Новая заявка: ФИО=%s, Тел=%s, Дата=%s %s, Инфо=%+v\n",
				name, phone, bookingDate, bookingTime, payloadMap)

			// Возвращаем сообщение об успехе
			var successMsg string
			if bookingDate != "" && bookingTime != "" {
				successMsg = fmt.Sprintf(`Мы зарезервировали для вас <strong>%s в %s</strong>. Наш специалист свяжется с вами по номеру <strong>%s</strong> для подтверждения.`,
					html.EscapeString(bookingDate), html.EscapeString(bookingTime), html.EscapeString(phone))
			} else {
				successMsg = fmt.Sprintf(`Спасибо за обращение! Наш менеджер свяжется с вами по номеру <strong>%s</strong> в ближайшее рабочее время.`,
					html.EscapeString(phone))
			}

			successHTML := fmt.Sprintf(`
				<div class="p-4 bg-emerald-50 border border-emerald-300 rounded-xl text-emerald-800 text-sm shadow-sm transition-all duration-200">
					<div class="font-bold flex items-center gap-2">
						<svg class="w-5 h-5 text-emerald-600 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path></svg>
						Заявка успешно принята!
					</div>
					<p class="mt-2 text-xs text-emerald-700">%s</p>
				</div>
			`, successMsg)
			return e.HTML(http.StatusOK, successHTML)
		})

		return se.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
