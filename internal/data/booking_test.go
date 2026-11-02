package data

import (
	"testing"

	"ganigeorgiev.com/pocketbase/pocketbase"
	"ganigeorgiev.com/pocketbase/pocketbase/core"
)

func setupTestApp(t *testing.T) *pocketbase.PocketBase {
	t.Helper()
	testDir := t.TempDir()
	app := pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDataDir: testDir,
	})
	if err := app.Bootstrap(); err != nil {
		t.Fatalf("Failed to bootstrap test PB: %v", err)
	}
	if err := InitSchema(app); err != nil {
		t.Fatalf("Failed to init test schema: %v", err)
	}
	return app
}

func TestGetBookedSlotsLogic(t *testing.T) {
	app := setupTestApp(t)
	defer app.ResetBootstrapState()

	col, err := app.FindCollectionByNameOrId("leads")
	if err != nil {
		t.Fatalf("leads collection not found: %v", err)
	}

	// 1. Проверяем пустую базу — должен вернуть пустую map
	initialSlots := GetBookedSlots(app)
	if len(initialSlots) != 0 {
		t.Errorf("Expected 0 initial booked slots, got %d", len(initialSlots))
	}

	// 2. Создаем тестовые заявки: активные, отмененные, привязанные к специалистам
	testLeads := []struct {
		name       string
		phone      string
		date       string
		time       string
		status     string
		resourceID string
	}{
		{"Клиент 1", "+79991112233", "2026-11-10", "10:00", "новая", "spec_1"},
		{"Клиент 2", "+79991112234", "2026-11-10", "11:30", "подтверждена", "spec_2"},
		{"Клиент 3 (отмена)", "+79991112235", "2026-11-10", "14:00", "отменена", "spec_1"}, // Должен игнорироваться
		{"Клиент 4", "+79991112236", "2026-11-11", "09:00", "в работе", "spec_1"},
	}

	for _, tl := range testLeads {
		rec := core.NewRecord(col)
		rec.Set("name", tl.name)
		rec.Set("phone", tl.phone)
		rec.Set("booking_date", tl.date)
		rec.Set("booking_time", tl.time)
		rec.Set("status", tl.status)
		rec.Set("resource_id", tl.resourceID)
		if err := app.Save(rec); err != nil {
			t.Fatalf("Failed to save test lead %s: %v", tl.name, err)
		}
	}

	// 3. Выборка всех занятых слотов без фильтрации по специалисту
	allSlots := GetBookedSlots(app)
	day1Slots, exists := allSlots["2026-11-10"]
	if !exists {
		t.Fatalf("Expected slots for 2026-11-10")
	}
	if len(day1Slots) != 2 {
		t.Errorf("Expected exactly 2 active slots on 2026-11-10 (excluding cancelled), got %d: %v", len(day1Slots), day1Slots)
	}

	// Проверяем, что отмененный слот 14:00 не попал в результат
	for _, sTime := range day1Slots {
		if sTime == "14:00" {
			t.Errorf("Cancelled slot 14:00 should not be returned as booked")
		}
	}

	day2Slots, existsDay2 := allSlots["2026-11-11"]
	if !existsDay2 || len(day2Slots) != 1 || day2Slots[0] != "09:00" {
		t.Errorf("Expected [09:00] for 2026-11-11, got %v", day2Slots)
	}

	// 4. Фильтрация по специалисту (spec_1)
	spec1Slots := GetBookedSlots(app, "spec_1")
	if len(spec1Slots["2026-11-10"]) != 1 || spec1Slots["2026-11-10"][0] != "10:00" {
		t.Errorf("Expected only [10:00] for spec_1 on 2026-11-10, got %v", spec1Slots["2026-11-10"])
	}
	if len(spec1Slots["2026-11-11"]) != 1 || spec1Slots["2026-11-11"][0] != "09:00" {
		t.Errorf("Expected [09:00] for spec_1 on 2026-11-11, got %v", spec1Slots["2026-11-11"])
	}

	// 5. Фильтрация по специалисту (spec_2)
	spec2Slots := GetBookedSlots(app, "spec_2")
	if len(spec2Slots["2026-11-10"]) != 1 || spec2Slots["2026-11-10"][0] != "11:30" {
		t.Errorf("Expected [11:30] for spec_2 on 2026-11-10, got %v", spec2Slots["2026-11-10"])
	}
	if len(spec2Slots["2026-11-11"]) != 0 {
		t.Errorf("Expected 0 slots for spec_2 on 2026-11-11, got %v", spec2Slots["2026-11-11"])
	}

	// 6. Несуществующий специалист
	unknownSlots := GetBookedSlots(app, "spec_nonexistent")
	if len(unknownSlots) != 0 {
		t.Errorf("Expected 0 slots for unknown specialist, got %d", len(unknownSlots))
	}
}

func TestGetResourcesLogic(t *testing.T) {
	app := setupTestApp(t)
	defer app.ResetBootstrapState()

	col, err := app.FindCollectionByNameOrId("resources")
	if err != nil {
		t.Fatalf("resources collection not found: %v", err)
	}

	// Добавляем неактивного специалиста и специалиста с услугами через запятую
	inactiveRec := core.NewRecord(col)
	inactiveRec.Set("name", "Неактивный сотрудник")
	inactiveRec.Set("role", "Архивариус")
	inactiveRec.Set("active", false)
	inactiveRec.Set("order", 99)
	if err := app.Save(inactiveRec); err != nil {
		t.Fatalf("Failed to save inactive resource: %v", err)
	}

	activeCustomRec := core.NewRecord(col)
	activeCustomRec.Set("name", "Анна Смирнова")
	activeCustomRec.Set("role", "Юрист-консультант")
	activeCustomRec.Set("active", true)
	activeCustomRec.Set("order", 0) // Должен быть первым по сортировке
	activeCustomRec.Set("service_ids", "consultation, audit , verification")
	if err := app.Save(activeCustomRec); err != nil {
		t.Fatalf("Failed to save active custom resource: %v", err)
	}

	resources := GetResources(app)
	if len(resources) == 0 {
		t.Fatalf("Expected non-empty resources list")
	}

	// 1. Первый элемент должен быть с order=0 (Анна Смирнова)
	if resources[0].Name != "Анна Смирнова" {
		t.Errorf("Expected first resource to be Anna Smirnova (order 0), got: %s", resources[0].Name)
	}

	// 2. Проверяем парсинг и тримминг comma-separated service_ids
	expectedServices := []string{"consultation", "audit", "verification"}
	if len(resources[0].Services) != len(expectedServices) {
		t.Fatalf("Expected %d parsed services, got %d: %v", len(expectedServices), len(resources[0].Services), resources[0].Services)
	}
	for i, s := range expectedServices {
		if resources[0].Services[i] != s {
			t.Errorf("Service [%d] mismatch: expected %q, got %q", i, s, resources[0].Services[i])
		}
	}

	// 3. Проверяем, что неактивный специалист отфильтрован
	for _, res := range resources {
		if res.Name == "Неактивный сотрудник" {
			t.Errorf("Inactive resource should not be returned by GetResources")
		}
	}
}
