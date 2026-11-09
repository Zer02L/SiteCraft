package pages

import (
	"encoding/json"
	"fmt"
	"strings"

	"sitecraft/internal/models"
)

// tagsToJSON serializes string slice to JSON array string for Alpine.js.
func tagsToJSON(tags []string) string {
	if len(tags) == 0 {
		return "[]"
	}
	b, err := json.Marshal(tags)
	if err != nil {
		return "[]"
	}
	return string(b)
}

// escapeJSONString serializes string to JSON string literal for Alpine.js.
func escapeJSONString(s string) string {
	b, err := json.Marshal(s)
	if err != nil {
		return `""`
	}
	return string(b)
}

// countLeadsByStatus counts leads that are new and in-progress.
func countLeadsByStatus(leads []models.Lead) (newCount int, inProgressCount int) {
	for _, l := range leads {
		if l.Status == "новая" || l.Status == "" {
			newCount++
		} else if l.Status == "в обработке" {
			inProgressCount++
		}
	}
	return newCount, inProgressCount
}

// extractDigits returns only numeric digits from a string.
func extractDigits(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// formatPhoneMask formats an arbitrary phone number into +7 (XXX) XXX-XX-XX format for display with x-mask.
func formatPhoneMask(phone string) string {
	d := extractDigits(phone)
	if len(d) == 11 && (d[0] == '7' || d[0] == '8') {
		return fmt.Sprintf("+7 (%s) %s-%s-%s", d[1:4], d[4:7], d[7:9], d[9:11])
	}
	if len(d) == 10 {
		return fmt.Sprintf("+7 (%s) %s-%s-%s", d[0:3], d[3:6], d[6:8], d[8:10])
	}
	return phone
}

// isTabActive checks if the given tabID is present among active modules.
func isTabActive(tabID string, modules []models.AdminTabModule) bool {
	for _, m := range modules {
		if m.ID == tabID {
			return true
		}
	}
	return false
}

// formatWhatsAppPhone returns standard international digits suitable for wa.me links.
func formatWhatsAppPhone(phone string) string {
	d := extractDigits(phone)
	if len(d) == 11 && d[0] == '8' {
		return "7" + d[1:]
	}
	return d
}

// getInitials returns a 2-letter uppercase initials string from a person's name.
func getInitials(name string) string {
	fields := strings.Fields(strings.TrimSpace(name))
	if len(fields) == 0 {
		return "КЛ"
	}
	r0 := []rune(fields[0])
	if len(fields) == 1 {
		if len(r0) >= 2 {
			return strings.ToUpper(string(r0[:2]))
		}
		return strings.ToUpper(string(r0))
	}
	r1 := []rune(fields[1])
	if len(r0) > 0 && len(r1) > 0 {
		return strings.ToUpper(string(r0[:1])) + strings.ToUpper(string(r1[:1]))
	}
	return strings.ToUpper(string(r0[:1]))
}
