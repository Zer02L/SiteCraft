package main

import (
	"fmt"
	"html"
	"log"
	"net/http"
	"os"
	"strings"

	"lawyer/internal/data"
	"lawyer/internal/handlers"
	"lawyer/ui/pages"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

func main() {
	app := pocketbase.New()

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

	// Настройка маршрутов веб-сайта
	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		// Раздача скомпилированных стилей и ассетов
		se.Router.GET("/assets/{path...}", apis.Static(os.DirFS("./ui/assets"), false))

		// Главная страница (SSR с помощью Templ)
		se.Router.GET("/", func(e *core.RequestEvent) error {
			content := data.GetSiteContent(app)
			services := data.GetServices(app)
			bookedSlots := data.GetBookedSlots(app)

			e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
			return pages.Home(content, services, bookedSlots).Render(e.Request.Context(), e.Response)
		})

		// API для получения занятых слотов (JSON)
		se.Router.GET("/api/slots", func(e *core.RequestEvent) error {
			bookedSlots := data.GetBookedSlots(app)
			return e.JSON(http.StatusOK, bookedSlots)
		})

		// Приём заявок с формы с полной валидацией (HTMX endpoint)
		se.Router.POST("/api/leads", func(e *core.RequestEvent) error {
			name := strings.TrimSpace(e.Request.FormValue("name"))
			phone := strings.TrimSpace(e.Request.FormValue("phone"))
			bookingDate := strings.TrimSpace(e.Request.FormValue("booking_date"))
			bookingTime := strings.TrimSpace(e.Request.FormValue("booking_time"))
			actionType := strings.TrimSpace(e.Request.FormValue("action_type"))
			comment := strings.TrimSpace(e.Request.FormValue("comment"))
			consent := e.Request.FormValue("consent")

			input := handlers.LeadFormInput{
				Name:        name,
				Phone:       phone,
				BookingDate: bookingDate,
				BookingTime: bookingTime,
				ActionType:  actionType,
				Comment:     comment,
				Consent:     consent,
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

			// Проверка на повторное бронирование (если слот уже занят кем-то другим)
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
						<p class="text-xs text-amber-700">К сожалению, этот слот только что забронировали. Пожалуйста, выберите другое свободное время в сетке.</p>
					</div>
				`)
			}

			// Сохраняем в коллекцию leads в PocketBase
			if leadsCol, err := app.FindCollectionByNameOrId("leads"); err == nil {
				leadRecord := core.NewRecord(leadsCol)
				leadRecord.Set("name", name)
				leadRecord.Set("phone", phone)
				leadRecord.Set("booking_date", bookingDate)
				leadRecord.Set("booking_time", bookingTime)
				leadRecord.Set("action_type", actionType)
				leadRecord.Set("comment", comment)
				leadRecord.Set("status", "новая")
				if err := app.Save(leadRecord); err != nil {
					log.Printf("[Leads Error] Не удалось сохранить заявку в базу: %v\n", err)
					return e.HTML(http.StatusInternalServerError, `
						<div class="p-4 bg-rose-50 border border-rose-300 rounded-xl text-rose-800 text-sm">
							<div class="font-bold">Ошибка сервера</div>
							<p class="mt-1 text-xs">Не удалось сохранить запись. Пожалуйста, повторите попытку или позвоните нам по телефону.</p>
						</div>
					`)
				}
			}

			log.Printf("Новая запись: ФИО=%s, Тел=%s, Дата=%s %s, Действие=%s, Коммент=%s\n",
				name, phone, bookingDate, bookingTime, actionType, comment)

			// Возвращаем аккуратное HTML-сообщение об успехе для HTMX
			successHTML := fmt.Sprintf(`
				<div class="p-4 bg-emerald-50 border border-emerald-300 rounded-xl text-emerald-800 text-sm shadow-sm transition-all duration-200">
					<div class="font-bold flex items-center gap-2">
						<svg class="w-5 h-5 text-emerald-600 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path></svg>
						Запись успешно подтверждена!
					</div>
					<p class="mt-2 text-xs text-emerald-700">Мы зарезервировали для вас <strong>%s в %s</strong>. Наш специалист свяжется с вами по номеру <strong>%s</strong> для предварительной консультации и сверки пакета документов.</p>
				</div>
			`, html.EscapeString(bookingDate), html.EscapeString(bookingTime), html.EscapeString(phone))
			return e.HTML(http.StatusOK, successHTML)
		})

		return se.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
