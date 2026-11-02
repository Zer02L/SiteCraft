package data

import (
	"log"

	"ganigeorgiev.com/pocketbase/pocketbase/core"
)

func initLeadsCollection(app core.App) error {
	col, err := app.FindCollectionByNameOrId("leads")
	if err != nil {
		col = core.NewBaseCollection("leads")
		col.Fields.Add(
			&core.TextField{Name: "name", Required: true},
			&core.TextField{Name: "phone", Required: true},
			&core.TextField{Name: "email"},
			&core.TextField{Name: "source"},
			&core.TextField{Name: "booking_date"},
			&core.TextField{Name: "booking_time"},
			&core.TextField{Name: "resource_id"},
			&core.TextField{Name: "status"}, // например, "новая", "обработана", "отменена"
			&core.JSONField{Name: "payload"}, // динамические поля: action_type, comment, vin, etc.
			&core.AutodateField{Name: "created", OnCreate: true},
			&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true},
		)

		if err := app.Save(col); err != nil {
			return err
		}
		log.Println("[Migration] Коллекция leads для онлайн-заявок успешно создана")
	} else {
		changed := false
		if col.Fields.GetByName("resource_id") == nil {
			col.Fields.Add(&core.TextField{Name: "resource_id"})
			changed = true
		}
		if col.Fields.GetByName("created") == nil {
			col.Fields.Add(&core.AutodateField{Name: "created", OnCreate: true})
			changed = true
		}
		if col.Fields.GetByName("updated") == nil {
			col.Fields.Add(&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true})
			changed = true
		}
		if changed {
			_ = app.Save(col)
		}
	}
	return nil
}

func initChatMessagesCollection(app core.App) error {
	col, err := app.FindCollectionByNameOrId("chat_messages")
	if err != nil {
		col = core.NewBaseCollection("chat_messages")
		col.Fields.Add(
			&core.TextField{Name: "session_id", Required: true},
			&core.TextField{Name: "sender", Required: true}, // "visitor" or "operator"
			&core.TextField{Name: "message", Required: true},
			&core.NumberField{Name: "tg_msg_id"},
		)

		if err := app.Save(col); err != nil {
			return err
		}
		log.Println("[Migration] Коллекция chat_messages для Live Web-to-TG чата успешно создана")
	}
	return nil
}
