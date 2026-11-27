package admin_leads

import (
	"encoding/json"
	"fmt"
	"sitecraft/internal/models"
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

// LeadDetailDTO carries sanitized lead details to the single reusable LeadDetailDialog.
type LeadDetailDTO struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Initials       string `json:"initials"`
	Phone          string `json:"phone"`
	PhoneFormatted string `json:"phoneFormatted"`
	WhatsAppPhone  string `json:"waPhone"`
	Email          string `json:"email"`
	Source         string `json:"source"`
	BookingDate    string `json:"bookingDate"`
	BookingTime    string `json:"bookingTime"`
	Created        string `json:"created"`
	Status         string `json:"status"`
	Comment        string `json:"comment"`
	QuizPrice      string `json:"quizPrice"`
	UTMSource      string `json:"utmSource"`
	UTMCampaign    string `json:"utmCampaign"`
}

// BuildLeadDetailDTO converts a Lead model into a JSON-safe DTO for the slide-over dialog.
func BuildLeadDetailDTO(lead models.Lead) LeadDetailDTO {
	status := lead.Status
	if status == "" {
		status = "новая"
	}
	dto := LeadDetailDTO{
		ID:             lead.ID,
		Name:           lead.Name,
		Initials:       GetInitials(lead.Name),
		Phone:          lead.Phone,
		PhoneFormatted: FormatPhoneMask(lead.Phone),
		WhatsAppPhone:  FormatWhatsAppPhone(lead.Phone),
		Email:          lead.Email,
		Source:         lead.Source,
		BookingDate:    lead.BookingDate,
		BookingTime:    lead.BookingTime,
		Created:        lead.Created,
		Status:         status,
	}
	if lead.Payload != nil {
		if c, ok := lead.Payload["comment"].(string); ok {
			dto.Comment = c
		}
		if utmSrc, ok := lead.Payload["utm_source"].(string); ok {
			dto.UTMSource = utmSrc
		}
		if utmCmp, ok := lead.Payload["utm_campaign"].(string); ok {
			dto.UTMCampaign = utmCmp
		}
		if qMap, ok := lead.Payload["quiz_summary"].(map[string]interface{}); ok {
			if p, ok := qMap["estimated_price"]; ok && p != nil {
				dto.QuizPrice = fmt.Sprintf("%v", p)
			}
		}
	}
	return dto
}

// BuildLeadDetailJSON serializes a LeadDetailDTO to JSON for safe data-lead attributes.
func BuildLeadDetailJSON(lead models.Lead) string {
	b, err := json.Marshal(BuildLeadDetailDTO(lead))
	if err != nil {
		return "{}"
	}
	return string(b)
}
