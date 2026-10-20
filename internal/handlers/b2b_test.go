package handlers

import (
	"strings"
	"testing"

	"sitecraft/internal/data"

	"github.com/pocketbase/pocketbase"
)

func TestB2BModulesAndInvoiceData(t *testing.T) {
	app := pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDataDir: t.TempDir(),
	})
	if err := app.Bootstrap(); err != nil {
		t.Fatalf("Bootstrap error: %v", err)
	}
	if err := data.InitSchema(app); err != nil {
		t.Fatalf("InitSchema error: %v", err)
	}

	// 1. Тестирование генератора счетов (Модуль 15)
	invoice, err := data.GetInvoiceData(app, "SRV-4182")
	if err != nil {
		t.Fatalf("GetInvoiceData error: %v", err)
	}
	if invoice == nil {
		t.Fatal("Expected invoice to not be nil")
	}
	if invoice.TotalAmount <= 0 {
		t.Errorf("Expected positive invoice amount, got %.2f", invoice.TotalAmount)
	}
	if !strings.Contains(invoice.PaymentCode, "ST00012") {
		t.Errorf("Expected GOST QR code string, got %s", invoice.PaymentCode)
	}
	if invoice.Supplier.INN == "" {
		t.Error("Expected supplier INN to be populated")
	}

	// 2. Тестирование сети филиалов (Модуль 16)
	branches := data.GetBranches(app)
	if len(branches) == 0 {
		t.Fatal("Expected demo branches to be initialized in DB")
	}
	foundMain := false
	for _, b := range branches {
		if b.IsMain {
			foundMain = true
		}
		if b.City == "" || b.Address == "" || b.Phone == "" {
			t.Errorf("Branch missing core fields: %+v", b)
		}
	}
	if !foundMain {
		t.Error("Expected at least one main branch")
	}

	// 3. Тестирование модуля вакансий и HR-откликов (Модуль 18)
	vacancies := data.GetVacancies(app)
	if len(vacancies) == 0 {
		t.Fatal("Expected demo vacancies to be initialized in DB")
	}
	for _, v := range vacancies {
		if v.Title == "" || v.Salary == "" {
			t.Errorf("Vacancy missing title or salary: %+v", v)
		}
	}

	// Тест создания отклика на вакансию
	err = data.CreateVacancyLead(app, "Иван Кандидатов", "+7 (999) 555-44-33", "ivan@test.local", vacancies[0].Title, "https://hh.ru/resume/123", "Большой опыт")
	if err != nil {
		t.Fatalf("Failed to create vacancy lead: %v", err)
	}

	leadsCount, err := app.CountRecords("leads")
	if err != nil || leadsCount == 0 {
		t.Errorf("Expected leads count > 0 after creating vacancy application")
	}
}
