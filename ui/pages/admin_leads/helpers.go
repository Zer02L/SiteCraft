package admin_leads

import (
	"strings"
	"unicode"
)

// FormatPhoneMask formats a phone number into readable format.
func FormatPhoneMask(phone string) string {
	var digits strings.Builder
	for _, r := range phone {
		if unicode.IsDigit(r) {
			digits.WriteRune(r)
		}
	}
	s := digits.String()
	if len(s) == 11 && (s[0] == '7' || s[0] == '8') {
		return "+7 (" + s[1:4] + ") " + s[4:7] + "-" + s[7:9] + "-" + s[9:11]
	}
	return phone
}

// FormatWhatsAppPhone returns digits for wa.me link.
func FormatWhatsAppPhone(phone string) string {
	var digits strings.Builder
	for _, r := range phone {
		if unicode.IsDigit(r) {
			digits.WriteRune(r)
		}
	}
	s := digits.String()
	if len(s) == 11 && s[0] == '8' {
		return "7" + s[1:]
	}
	return s
}

// GetInitials returns initials for a person's name.
func GetInitials(name string) string {
	parts := strings.Fields(name)
	if len(parts) >= 2 {
		r1 := []rune(parts[0])
		r2 := []rune(parts[1])
		if len(r1) > 0 && len(r2) > 0 {
			return strings.ToUpper(string(r1[0]) + string(r2[0]))
		}
	}
	r := []rune(name)
	if len(r) >= 2 {
		return strings.ToUpper(string(r[:2]))
	} else if len(r) == 1 {
		return strings.ToUpper(string(r[0]))
	}
	return "ЛД"
}
