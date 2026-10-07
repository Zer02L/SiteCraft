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

func TestValidatePaymentForm(t *testing.T) {
	validOrderPayment := PaymentFormInput{
		OrderCode: "ORD-2026-101",
		Amount:    3500.0,
		Email:     "client@example.com",
		Phone:     "+7 (999) 000-11-22",
		Consent:   "on",
		Honeypot:  "",
	}

	t.Run("Valid order payment passes", func(t *testing.T) {
		errs := ValidatePaymentForm(validOrderPayment, true)
		if len(errs) != 0 {
			t.Errorf("expected 0 errors, got: %v", errs)
		}
	})

	t.Run("Honeypot catches bots immediately", func(t *testing.T) {
		input := validOrderPayment
		input.Honeypot = "https://bot-site.com"
		errs := ValidatePaymentForm(input, true)
		if len(errs) != 1 || !strings.Contains(errs[0], "Спам-фильтр") {
			t.Errorf("expected honeypot error, got: %v", errs)
		}
	})

	t.Run("Invalid email fails 54-FZ validation", func(t *testing.T) {
		input := validOrderPayment
		input.Email = "invalid-email"
		errs := ValidatePaymentForm(input, true)
		if len(errs) == 0 {
			t.Errorf("expected error for invalid email, got none")
		}

		input.Email = ""
		errs = ValidatePaymentForm(input, true)
		if len(errs) == 0 {
			t.Errorf("expected error for empty email, got none")
		}
	})

	t.Run("Invalid amount fails", func(t *testing.T) {
		input := validOrderPayment
		input.Amount = 0
		errs := ValidatePaymentForm(input, true)
		if len(errs) == 0 {
			t.Errorf("expected error for 0 amount, got none")
		}

		input.Amount = 5.0
		errs = ValidatePaymentForm(input, true)
		if len(errs) == 0 {
			t.Errorf("expected error for amount < 10, got none")
		}

		input.Amount = 1500000.0
		errs = ValidatePaymentForm(input, true)
		if len(errs) == 0 {
			t.Errorf("expected error for amount > 1M, got none")
		}
	})

	t.Run("Missing order code fails when isOrderPayment is true", func(t *testing.T) {
		input := validOrderPayment
		input.OrderCode = ""
		errs := ValidatePaymentForm(input, true)
		if len(errs) == 0 {
			t.Errorf("expected error for missing order code, got none")
		}
	})

	t.Run("Missing phone fails for order payment", func(t *testing.T) {
		input := validOrderPayment
		input.Phone = ""
		errs := ValidatePaymentForm(input, true)
		if len(errs) == 0 {
			t.Errorf("expected error for missing phone on order payment, got none")
		}
	})

	t.Run("Missing consent fails 152-FZ", func(t *testing.T) {
		input := validOrderPayment
		input.Consent = ""
		errs := ValidatePaymentForm(input, true)
		if len(errs) == 0 {
			t.Errorf("expected error for missing consent, got none")
		}
	})
}

func TestValidateLeadFormLocalized_EN(t *testing.T) {
	t.Run("Returns English error messages when lang is en", func(t *testing.T) {
		input := LeadFormInput{
			Name:     "",
			Phone:    "123",
			Consent:  "",
			Honeypot: "http://bot.com",
			Mode:     "simple",
		}
		errs := ValidateLeadFormLocalized(input, "en")
		if len(errs) != 1 || errs[0] != "Spam filter blocked the submission" {
			t.Errorf("expected English spam error, got: %v", errs)
		}

		input.Honeypot = ""
		errs = ValidateLeadFormLocalized(input, "en")
		if len(errs) < 3 {
			t.Errorf("expected multiple English errors, got: %v", errs)
		}
		foundName := false
		foundPhone := false
		for _, e := range errs {
			if strings.Contains(e, "Please provide your name") {
				foundName = true
			}
			if strings.Contains(e, "Invalid phone number") {
				foundPhone = true
			}
		}
		if !foundName || !foundPhone {
			t.Errorf("expected English validation messages, got: %v", errs)
		}
	})
}
