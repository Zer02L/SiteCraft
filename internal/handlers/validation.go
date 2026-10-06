package handlers

import (
	"regexp"
	"strings"
	"time"
)

// LeadFormInput represents input fields submitted from the booking form.
type LeadFormInput struct {
	Name        string
	Phone       string
	BookingDate string
	BookingTime string
	ActionType  string
	Comment     string
	Consent     string
}

var nonDigitRegex = regexp.MustCompile(`\D`)
var timeSlotRegex = regexp.MustCompile(`^\d{2}:\d{2}$`)

// ValidateLeadForm validates the input values from the appointment booking form.
func ValidateLeadForm(input LeadFormInput) []string {
	var errs []string

	name := strings.TrimSpace(input.Name)
	if name == "" {
		errs = append(errs, "Укажите ФИО заявителя")
	} else {
		runeName := []rune(name)
		if len(runeName) < 2 {
			errs = append(errs, "ФИО должно содержать минимум 2 символа")
		} else if len(runeName) > 100 {
			errs = append(errs, "ФИО не должно превышать 100 символов")
		}
	}

	phone := strings.TrimSpace(input.Phone)
	digitsOnly := nonDigitRegex.ReplaceAllString(phone, "")
	if phone == "" {
		errs = append(errs, "Укажите контактный номер телефона")
	} else if len(digitsOnly) < 10 || len(digitsOnly) > 15 {
		errs = append(errs, "Некорректный номер телефона (требуется от 10 цифр)")
	}

	bookingDate := strings.TrimSpace(input.BookingDate)
	if bookingDate == "" {
		errs = append(errs, "Выберите дату приёма в календаре")
	} else {
		if _, err := time.Parse("2006-01-02", bookingDate); err != nil {
			errs = append(errs, "Некорректный формат даты приёма")
		}
	}

	bookingTime := strings.TrimSpace(input.BookingTime)
	if bookingTime == "" {
		errs = append(errs, "Выберите свободное время приёма из доступных слотов")
	} else {
		if !timeSlotRegex.MatchString(bookingTime) {
			errs = append(errs, "Некорректный формат времени приёма")
		}
	}

	consent := strings.TrimSpace(input.Consent)
	if consent != "on" && consent != "true" && consent != "1" {
		errs = append(errs, "Необходимо согласие на обработку персональных данных (152-ФЗ)")
	}

	comment := strings.TrimSpace(input.Comment)
	if len([]rune(comment)) > 1000 {
		errs = append(errs, "Длина комментария не должна превышать 1000 символов")
	}

	return errs
}
