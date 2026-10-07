package handlers

import (
	"regexp"
	"strings"
	"time"
)

// LeadFormInput represents input fields submitted from the booking / lead form.
type LeadFormInput struct {
	Name        string
	Phone       string
	BookingDate string
	BookingTime string
	ActionType  string
	Comment     string
	Consent     string
	Honeypot    string // company_website_url (must remain empty for humans)
	Mode        string // "simple" or "slots"
}

var nonDigitRegex = regexp.MustCompile(`\D`)
var timeSlotRegex = regexp.MustCompile(`^\d{2}:\d{2}$`)
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// ValidateLeadForm validates the input values using the default Russian locale.
func ValidateLeadForm(input LeadFormInput) []string {
	return ValidateLeadFormLocalized(input, "ru")
}

// ValidateLeadFormLocalized validates the input values from the appointment booking / contact form in the given locale.
func ValidateLeadFormLocalized(input LeadFormInput, lang string) []string {
	var errs []string
	isEn := strings.HasPrefix(strings.ToLower(lang), "en")

	// 1. Anti-spam Honeypot check
	if strings.TrimSpace(input.Honeypot) != "" {
		if isEn {
			errs = append(errs, "Spam filter blocked the submission")
		} else {
			errs = append(errs, "Спам-фильтр заблокировал отправку формы")
		}
		return errs
	}

	// 2. Name validation
	name := strings.TrimSpace(input.Name)
	if name == "" {
		if isEn {
			errs = append(errs, "Please provide your name")
		} else {
			errs = append(errs, "Укажите ваше имя")
		}
	} else {
		runeName := []rune(name)
		if len(runeName) < 2 {
			if isEn {
				errs = append(errs, "Name must be at least 2 characters")
			} else {
				errs = append(errs, "Имя должно содержать минимум 2 символа")
			}
		} else if len(runeName) > 100 {
			if isEn {
				errs = append(errs, "Name must not exceed 100 characters")
			} else {
				errs = append(errs, "Имя не должно превышать 100 символов")
			}
		}
	}

	// 3. Phone validation
	phone := strings.TrimSpace(input.Phone)
	digitsOnly := nonDigitRegex.ReplaceAllString(phone, "")
	if phone == "" {
		if isEn {
			errs = append(errs, "Please provide a contact phone number")
		} else {
			errs = append(errs, "Укажите контактный номер телефона")
		}
	} else if len(digitsOnly) < 10 || len(digitsOnly) > 15 {
		if isEn {
			errs = append(errs, "Invalid phone number (at least 10 digits required)")
		} else {
			errs = append(errs, "Некорректный номер телефона (требуется от 10 цифр)")
		}
	}

	// 4. Booking Date & Time (if mode is "slots" or provided)
	if input.Mode == "slots" || input.BookingDate != "" || input.BookingTime != "" {
		bookingDate := strings.TrimSpace(input.BookingDate)
		if bookingDate == "" {
			if isEn {
				errs = append(errs, "Please select an appointment date")
			} else {
				errs = append(errs, "Выберите дату приёма в календаре")
			}
		} else {
			if _, err := time.Parse("2006-01-02", bookingDate); err != nil {
				if isEn {
					errs = append(errs, "Invalid appointment date format")
				} else {
					errs = append(errs, "Некорректный формат даты приёма")
				}
			}
		}

		bookingTime := strings.TrimSpace(input.BookingTime)
		if bookingTime == "" {
			if isEn {
				errs = append(errs, "Please select an available time slot")
			} else {
				errs = append(errs, "Выберите свободное время приёма из доступных слотов")
			}
		} else {
			if !timeSlotRegex.MatchString(bookingTime) {
				if isEn {
					errs = append(errs, "Invalid appointment time format")
				} else {
					errs = append(errs, "Некорректный формат времени приёма")
				}
			}
		}
	}

	// 5. Consent 152-ФЗ / Privacy policy
	consent := strings.TrimSpace(input.Consent)
	if consent != "on" && consent != "true" && consent != "1" {
		if isEn {
			errs = append(errs, "Consent to personal data processing is required")
		} else {
			errs = append(errs, "Необходимо согласие на обработку персональных данных (152-ФЗ)")
		}
	}

	// 6. Comment length check
	comment := strings.TrimSpace(input.Comment)
	if len([]rune(comment)) > 1000 {
		if isEn {
			errs = append(errs, "Comment must not exceed 1000 characters")
		} else {
			errs = append(errs, "Длина комментария не должна превышать 1000 символов")
		}
	}

	return errs
}

// PaymentFormInput represents fields submitted when initiating an online payment.
type PaymentFormInput struct {
	OrderCode string
	Amount    float64
	Email     string
	Phone     string
	Consent   string
	Honeypot  string
}

// ValidatePaymentForm validates user input for payment forms.
func ValidatePaymentForm(input PaymentFormInput, isOrderPayment bool) []string {
	var errs []string

	// 1. Anti-spam Honeypot check
	if strings.TrimSpace(input.Honeypot) != "" {
		errs = append(errs, "Спам-фильтр заблокировал отправку формы")
		return errs
	}

	// 2. Email validation (required for 54-FZ electronic receipt)
	email := strings.TrimSpace(input.Email)
	if email == "" {
		errs = append(errs, "Укажите Email для отправки электронного чека (54-ФЗ)")
	} else if len(email) > 150 || !emailRegex.MatchString(email) {
		errs = append(errs, "Укажите корректный адрес электронной почты (например, client@mail.ru)")
	}

	// 3. Phone validation
	phone := strings.TrimSpace(input.Phone)
	if phone != "" {
		digitsOnly := nonDigitRegex.ReplaceAllString(phone, "")
		if len(digitsOnly) < 10 || len(digitsOnly) > 15 {
			errs = append(errs, "Некорректный номер телефона (требуется от 10 цифр)")
		}
	} else if isOrderPayment {
		errs = append(errs, "Укажите контактный номер телефона")
	}

	// 4. Amount validation
	if input.Amount <= 0 {
		errs = append(errs, "Сумма оплаты должна быть больше 0 ₽")
	} else if input.Amount < 10 {
		errs = append(errs, "Минимальная сумма онлайн-платежа — 10 ₽")
	} else if input.Amount > 1000000 {
		errs = append(errs, "Сумма одного платежа не может превышать 1 000 000 ₽")
	}

	// 5. Order Code validation
	if isOrderPayment {
		cleanCode := strings.TrimSpace(input.OrderCode)
		if cleanCode == "" {
			errs = append(errs, "Не указан номер заказ-наряда")
		} else if len(cleanCode) < 3 || len(cleanCode) > 30 {
			errs = append(errs, "Некорректный номер заказ-наряда")
		}
	}

	// 6. 152-FZ Consent check
	consent := strings.TrimSpace(input.Consent)
	if consent != "on" && consent != "true" && consent != "1" {
		errs = append(errs, "Необходимо согласие на обработку персональных данных и условия оферты")
	}

	return errs
}

