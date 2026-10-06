package handlers

import (
	"strings"
	"testing"
)

func TestValidateLeadForm(t *testing.T) {
	validInput := LeadFormInput{
		Name:        "Иванов Иван Иванович",
		Phone:       "+7 (999) 123-45-67",
		BookingDate: "2026-10-07",
		BookingTime: "10:15",
		ActionType:  "доверенность",
		Comment:     "Пакет документов подготовлен",
		Consent:     "on",
		Honeypot:    "",
		Mode:        "slots",
	}

	t.Run("Valid input passes without errors", func(t *testing.T) {
		errs := ValidateLeadForm(validInput)
		if len(errs) != 0 {
			t.Errorf("expected 0 errors, got: %v", errs)
		}
	})

	t.Run("Honeypot catches spam bots", func(t *testing.T) {
		input := validInput
		input.Honeypot = "https://spam-bot.ru"
		errs := ValidateLeadForm(input)
		if len(errs) != 1 || !strings.Contains(errs[0], "Спам-фильтр") {
			t.Errorf("expected honeypot error, got: %v", errs)
		}
	})

	t.Run("Simple mode does not require date/time", func(t *testing.T) {
		input := LeadFormInput{
			Name:     "Петров Петр",
			Phone:    "+7 (999) 555-44-33",
			Consent:  "on",
			Honeypot: "",
			Mode:     "simple",
		}
		errs := ValidateLeadForm(input)
		if len(errs) != 0 {
			t.Errorf("expected 0 errors in simple mode, got: %v", errs)
		}
	})

	t.Run("Empty name returns error", func(t *testing.T) {
		input := validInput
		input.Name = "   "
		errs := ValidateLeadForm(input)
		if len(errs) == 0 {
			t.Errorf("expected error for empty name, got none")
		}
	})

	t.Run("Short name returns error", func(t *testing.T) {
		input := validInput
		input.Name = "А"
		errs := ValidateLeadForm(input)
		if len(errs) == 0 {
			t.Errorf("expected error for short name, got none")
		}
	})

	t.Run("Invalid phone returns error", func(t *testing.T) {
		input := validInput
		input.Phone = "12345"
		errs := ValidateLeadForm(input)
		if len(errs) == 0 {
			t.Errorf("expected error for invalid phone, got none")
		}
	})

	t.Run("Missing date or time in slots mode returns error", func(t *testing.T) {
		input := validInput
		input.BookingDate = ""
		input.BookingTime = ""
		errs := ValidateLeadForm(input)
		if len(errs) < 2 {
			t.Errorf("expected at least 2 errors for missing date and time, got: %v", errs)
		}
	})

	t.Run("Missing consent returns error", func(t *testing.T) {
		input := validInput
		input.Consent = ""
		errs := ValidateLeadForm(input)
		if len(errs) == 0 {
			t.Errorf("expected error for missing consent, got none")
		}
	})

	t.Run("Too long comment returns error", func(t *testing.T) {
		input := validInput
		input.Comment = strings.Repeat("a", 1001)
		errs := ValidateLeadForm(input)
		if len(errs) == 0 {
			t.Errorf("expected error for too long comment, got none")
		}
	})
}
