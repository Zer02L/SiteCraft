package i18n

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

const (
	// LangRU represents Russian locale.
	LangRU = "ru"
	// LangEN represents English locale.
	LangEN = "en"
	// DefaultLang is the fallback locale.
	DefaultLang = LangRU
	// CookieName is the HTTP cookie name used for user's language preference.
	CookieName = "lang"
)

// SupportedLocales contains all available language codes.
var SupportedLocales = []string{LangRU, LangEN}

type contextKey struct{}

var localeKey = contextKey{}

// Normalize normalizes any language code string into a supported locale ("ru" or "en").
// Defaults to DefaultLang ("ru").
func Normalize(lang string) string {
	l := strings.ToLower(strings.TrimSpace(lang))
	if l == "" {
		return DefaultLang
	}
	if strings.HasPrefix(l, "en") {
		return LangEN
	}
	if strings.HasPrefix(l, "ru") {
		return LangRU
	}
	return DefaultLang
}

// IsSupported returns true if the specified language is supported.
func IsSupported(lang string) bool {
	l := strings.ToLower(strings.TrimSpace(lang))
	return l == LangRU || l == LangEN
}

// WithLocale returns a context containing the given locale.
func WithLocale(ctx context.Context, lang string) context.Context {
	return context.WithValue(ctx, localeKey, Normalize(lang))
}

// GetLocale extracts the current locale from context, defaulting to "ru".
func GetLocale(ctx context.Context) string {
	if ctx == nil {
		return DefaultLang
	}
	if val, ok := ctx.Value(localeKey).(string); ok && val != "" {
		return val
	}
	return DefaultLang
}

// DetectLang extracts the preferred locale from an HTTP request.
// Resolution order:
// 1. Query parameter ?lang=...
// 2. Cookie "lang"
// 3. Accept-Language HTTP header
// 4. Fallback: DefaultLang ("ru")
func DetectLang(r *http.Request) string {
	if r == nil {
		return DefaultLang
	}

	// 1. Query parameter
	if q := strings.TrimSpace(r.URL.Query().Get("lang")); q != "" {
		return Normalize(q)
	}

	// 2. Cookie
	if c, err := r.Cookie(CookieName); err == nil && c.Value != "" {
		return Normalize(c.Value)
	}

	// 3. Accept-Language header
	if al := r.Header.Get("Accept-Language"); al != "" {
		if parsed := parseAcceptLanguage(al); parsed != "" {
			return parsed
		}
	}

	return DefaultLang
}

// parseAcceptLanguage parses an Accept-Language header (e.g., "en-US,en;q=0.9,ru;q=0.8").
func parseAcceptLanguage(header string) string {
	parts := strings.Split(header, ",")
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if semi := strings.Index(item, ";"); semi != -1 {
			item = item[:semi]
		}
		item = strings.TrimSpace(item)
		if strings.HasPrefix(strings.ToLower(item), "en") {
			return LangEN
		}
		if strings.HasPrefix(strings.ToLower(item), "ru") {
			return LangRU
		}
	}
	return ""
}

// T translates a key using the locale found in context.
func T(ctx context.Context, key string, args ...any) string {
	return Translate(GetLocale(ctx), key, args...)
}

// Translate translates a key for a given locale string.
func Translate(lang string, key string, args ...any) string {
	lang = Normalize(lang)
	catalog, ok := catalogs[lang]
	if !ok {
		catalog = catalogs[DefaultLang]
	}

	val, found := catalog[key]
	if !found {
		// Fallback to default catalog
		if defCat, defOk := catalogs[DefaultLang]; defOk {
			val, found = defCat[key]
		}
	}
	if !found {
		val = key
	}

	if len(args) > 0 {
		return fmt.Sprintf(val, args...)
	}
	return val
}

// Plural returns the appropriate plural string based on count and context locale.
func Plural(ctx context.Context, count int, one, few, many string) string {
	return PluralByLang(GetLocale(ctx), count, one, few, many)
}

// PluralByLang selects plural forms for Russian (1, 2-4, 5+) or English (1, other).
func PluralByLang(lang string, count int, one, few, many string) string {
	lang = Normalize(lang)
	if count < 0 {
		count = -count
	}

	if lang == LangRU {
		n := count % 100
		n10 := count % 10
		if n >= 11 && n <= 19 {
			return many
		}
		if n10 == 1 {
			return one
		}
		if n10 >= 2 && n10 <= 4 {
			return few
		}
		return many
	}

	// English & fallback
	if count == 1 {
		return one
	}
	return many
}

// OGLocale returns OpenGraph locale string (e.g. "ru_RU" or "en_US").
func OGLocale(ctx context.Context) string {
	if GetLocale(ctx) == LangEN {
		return "en_US"
	}
	return "ru_RU"
}

// OtherLang returns the alternate language for quick switching.
func OtherLang(current string) string {
	if Normalize(current) == LangRU {
		return LangEN
	}
	return LangRU
}
