package ai

import (
	"regexp"
	"strings"
)

var (
	// emailRegex matches standard email patterns.
	emailRegex = regexp.MustCompile(`(?i)\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}\b`)

	// phoneRegex matches common Russian and international phone formats:
	// +7 (999) 123-45-67, 8 (999) 123-45-67, +79991234567, 89991234567, etc.
	phoneRegex = regexp.MustCompile(`(?:\+7|8)[\s\-\(]*(\d{3})[\s\-\)]*(\d{3})[\s\-]*(\d{2})[\s\-]*(\d{2})`)
)

// MaskPII masks sensitive personal data (phones and emails) in accordance with Russian 152-FZ.
// Example:
//
//	"+7 (999) 123-45-67" -> "+7 (999) ***-45-67"
//	"ivan.petrov@example.com" -> "i***v@example.com"
func MaskPII(text string) string {
	if text == "" {
		return ""
	}

	// 1. Mask Phone Numbers
	text = phoneRegex.ReplaceAllStringFunc(text, func(match string) string {
		sub := phoneRegex.FindStringSubmatch(match)
		if len(sub) == 5 {
			// sub[1] = area code (999), sub[2] = 123, sub[3] = 45, sub[4] = 67
			// Format as: +7 (999) ***-xx-xx
			return "+7 (" + sub[1] + ") ***-" + sub[3] + "-" + sub[4]
		}
		return "+7 (***) ***-**-**"
	})

	// 2. Mask Email Addresses
	text = emailRegex.ReplaceAllStringFunc(text, func(email string) string {
		parts := strings.SplitN(email, "@", 2)
		if len(parts) != 2 {
			return email
		}
		user := parts[0]
		domain := parts[1]

		if len(user) <= 2 {
			return strings.Repeat("*", len(user)) + "@" + domain
		}
		// e.g. ivan -> i***n
		firstChar := string(user[0])
		lastChar := string(user[len(user)-1])
		return firstChar + "***" + lastChar + "@" + domain
	})

	return text
}
