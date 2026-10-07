package i18n

import (
	"context"
	"strings"

	"sitecraft/internal/models"
)

// LocalizeContent adapts site content fields to English if current locale is "en".
// If the content is in Russian and English was selected, standard preset strings are translated.
func LocalizeContent(ctx context.Context, c models.SiteContent) models.SiteContent {
	if GetLocale(ctx) != LangEN {
		return c
	}

	localized := c

	// Hero Title & Badge
	if localized.HeroTitle == "Нотариус" {
		localized.HeroTitle = "Notary Public"
	}
	if localized.HeroBadge == "Нотариальная контора открыта" {
		localized.HeroBadge = "Office is open"
	}
	if strings.Contains(localized.HeroSubtitle, "Совершение всех видов нотариальных действий") {
		localized.HeroSubtitle = "Execution of all types of notarial acts in strict accordance with the law. Legality, speed, and legal protection of your interests."
	}

	// Working hours
	if strings.HasPrefix(localized.WorkHours, "Понедельник — Пятница") {
		localized.WorkHours = "Monday — Friday: 9:00 AM - 5:30 PM"
	}
	if localized.LunchBreak == "13:00 - 14:00" {
		localized.LunchBreak = "1:00 PM - 2:00 PM"
	}
	if strings.Contains(localized.DaysOff, "суббота, воскресенье") {
		localized.DaysOff = "Days off: Saturday, Sunday"
	}

	// Titles and addresses
	if localized.OwnerTitle == "Нотариус города Краснодара" {
		localized.OwnerTitle = "Notary Public of Krasnodar"
	}
	if localized.NotaryTitle == "Нотариус города Краснодара" {
		localized.NotaryTitle = "Notary Public of Krasnodar"
	}
	if localized.Address == "Нотариальная контора нотариуса Захарова В. Б." {
		localized.Address = "Notary Office of V. B. Zakharov"
	}
	if localized.AddressSub == "Краснодарский край / г. Краснодар" {
		localized.AddressSub = "Krasnodar Krai / Krasnodar City"
	}

	return localized
}

// LocalizeService adapts single service item to English if locale is "en".
func LocalizeService(ctx context.Context, s models.ServiceItem) models.ServiceItem {
	if GetLocale(ctx) != LangEN {
		return s
	}

	loc := s
	switch s.Title {
	case "Удостоверение сделок и договоров":
		loc.Title = "Certification of Transactions & Contracts"
		loc.Description = "Real estate sale and purchase, donation deeds, prenuptial agreements."
		loc.Price = "statutory tariff"
	case "Оформление доверенностей":
		loc.Title = "Powers of Attorney"
		loc.Description = "General, judicial, and vehicular powers of attorney."
		if s.Price == "от 1 500 ₽" {
			loc.Price = "from 1,500 ₽"
		}
	case "Ведение наследственных дел":
		loc.Title = "Probate & Inheritance Administration"
		loc.Description = "Opening probate estates, will registry searches, certificates of inheritance."
		loc.Price = "statutory tariff"
	}
	return loc
}

// LocalizeServices adapts a list of services to English if locale is "en".
func LocalizeServices(ctx context.Context, services []models.ServiceItem) []models.ServiceItem {
	if GetLocale(ctx) != LangEN || len(services) == 0 {
		return services
	}

	result := make([]models.ServiceItem, len(services))
	for i, s := range services {
		result[i] = LocalizeService(ctx, s)
	}
	return result
}
