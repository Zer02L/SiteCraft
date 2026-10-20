package models

// BranchItem represents a physical branch/location of the company (Блок 4: Сеть филиалов).
type BranchItem struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`       // e.g. "Филиал Центр — Невский"
	City      string   `json:"city"`       // e.g. "Санкт-Петербург"
	Address   string   `json:"address"`    // e.g. "Невский пр-т, д. 42"
	Phone     string   `json:"phone"`      // e.g. "+7 (812) 309-88-21"
	WorkHours string   `json:"work_hours"` // e.g. "Пн-Вс: 09:00 — 21:00"
	Lat       float64  `json:"lat"`
	Lng       float64  `json:"lng"`
	Metro     string   `json:"metro,omitempty"`    // e.g. "Гостиный двор"
	Services  []string `json:"services,omitempty"` // e.g. ["Диагностика", "ТО", "Кузовной цех"]
	IsMain    bool     `json:"is_main"`
	Active    bool     `json:"active"`
}

// VacancyItem represents an open job position (Блок 4: Модуль HR и Вакансий).
type VacancyItem struct {
	ID               string   `json:"id"`
	Title            string   `json:"title"`            // e.g. "Мастер-приемщик / Сервисный консультант"
	Department       string   `json:"department"`       // e.g. "Сервисный центр"
	Salary           string   `json:"salary"`           // e.g. "от 120 000 ₽"
	Experience       string   `json:"experience"`       // e.g. "от 2 лет"
	Schedule         string   `json:"schedule"`         // e.g. "Сменный график 2/2"
	Responsibilities []string `json:"responsibilities"` // Обязанности
	Requirements     []string `json:"requirements"`     // Требования
	Benefits         []string `json:"benefits"`         // Условия и бонусы
	Active           bool     `json:"active"`
	Order            int      `json:"order"`
}

// InvoiceData represents a commercial B2B payment invoice and act of services (Блок 4: B2B-счета).
type InvoiceData struct {
	InvoiceNumber string             `json:"invoice_number"` // e.g. "СЧ-SRV-4182"
	InvoiceDate   string             `json:"invoice_date"`   // e.g. "19 октября 2026 г."
	OrderCode     string             `json:"order_code"`     // e.g. "SRV-4182"
	Supplier      InvoiceEntity      `json:"supplier"`       // Поставщик (Исполнитель)
	Customer      InvoiceEntity      `json:"customer"`       // Покупатель (Заказчик)
	Items         []InvoiceLineItem  `json:"items"`          // Список услуг / запчастей
	TotalAmount   float64            `json:"total_amount"`
	VATText       string             `json:"vat_text"`       // "Без НДС (УСН)"
	TotalInWords  string             `json:"total_in_words"` // Сумма прописью
	PaymentCode   string             `json:"payment_code"`   // QR-строка по ГОСТ Р 56042-2014
	Notes         string             `json:"notes,omitempty"`
}

// InvoiceEntity represents legal or individual party in invoice.
type InvoiceEntity struct {
	Name        string `json:"name"`         // Наименование юрлица или ФИО
	INN         string `json:"inn"`          // ИНН
	KPP         string `json:"kpp,omitempty"`// КПП
	OGRN        string `json:"ogrn"`         // ОГРН / ОГРНИП
	Address     string `json:"address"`      // Юридический адрес
	Phone       string `json:"phone"`        // Телефон
	BankName    string `json:"bank_name"`    // Банк
	BankBIK     string `json:"bank_bik"`     // БИК
	BankAccount string `json:"bank_account"` // Расчетный счет
	BankCorr    string `json:"bank_corr"`    // Корр. счет
}

// InvoiceLineItem represents an individual line item in invoice.
type InvoiceLineItem struct {
	Index    int     `json:"index"`
	Title    string  `json:"title"`
	Unit     string  `json:"unit"` // "усл.", "шт."
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
	Amount   float64 `json:"amount"`
}
