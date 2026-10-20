package handlers

import (
	"strings"
	"testing"
	"time"
)

func TestGenerateICS(t *testing.T) {
	title := "Запись: Технический аудит и диагностика — SiteCraft Auto"
	description := "Клиент: Алексей Смирнов\nТелефон: +7 (999) 000-11-22"
	location := "г. Москва, ул. Автозаводская, 23"
	startTime := time.Date(2026, 11, 5, 14, 30, 0, 0, time.UTC)
	endTime := startTime.Add(45 * time.Minute)
	uid := "lead_12345@sitecraft"

	icsBytes := GenerateICS(title, description, location, startTime, endTime, uid)
	icsContent := string(icsBytes)

	// Проверяем обязательные заголовки RFC 5545
	if !strings.Contains(icsContent, "BEGIN:VCALENDAR") {
		t.Errorf("Expected BEGIN:VCALENDAR in ICS output")
	}
	if !strings.Contains(icsContent, "VERSION:2.0") {
		t.Errorf("Expected VERSION:2.0 in ICS output")
	}
	if !strings.Contains(icsContent, "PRODID:-//SiteCraft Engine//Booking Calendar 1.0//RU") {
		t.Errorf("Expected PRODID in ICS output")
	}
	if !strings.Contains(icsContent, "BEGIN:VEVENT") {
		t.Errorf("Expected BEGIN:VEVENT in ICS output")
	}
	if !strings.Contains(icsContent, "SUMMARY:Запись: Технический аудит и диагностика — SiteCraft Auto") {
		t.Errorf("Expected SUMMARY with service name and site name")
	}
	if !strings.Contains(icsContent, "LOCATION:г. Москва\\, ул. Автозаводская\\, 23") {
		t.Errorf("Expected escaped LOCATION in ICS output, got:\n%s", icsContent)
	}
	if !strings.Contains(icsContent, "STATUS:CONFIRMED") {
		t.Errorf("Expected STATUS:CONFIRMED in ICS output")
	}
	if !strings.Contains(icsContent, "UID:lead_12345@sitecraft") {
		t.Errorf("Expected UID in ICS output")
	}
	if !strings.Contains(icsContent, "DTSTART:20261105T143000Z") {
		t.Errorf("Expected DTSTART:20261105T143000Z in ICS output")
	}
	if !strings.Contains(icsContent, "DTEND:20261105T151500Z") {
		t.Errorf("Expected DTEND:20261105T151500Z in ICS output")
	}
	if !strings.Contains(icsContent, "END:VCALENDAR") {
		t.Errorf("Expected END:VCALENDAR in ICS output")
	}
}
