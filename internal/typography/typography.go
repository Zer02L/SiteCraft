package typography

import (
	"regexp"
	"strings"
)

var (
	// Одиночные предлоги, союзы и частицы русского языка, после которых ставится неразрывный пробел
	// Используем \x{00a0} для неразрывного пробела внутри символьного класса Go regexp
	rePrepositions = regexp.MustCompile(`(?i)(^|[\s\x{00a0}\(\[\{«„])(в|во|на|с|со|по|к|ко|из|изо|о|об|обо|от|ото|до|за|для|у|и|а|но|или|да|не|ни|то|как|же)\s+`)

	// Замена дефиса, окруженного пробелами, на длинное тире с неразрывным пробелом перед ним
	reDash = regexp.MustCompile(`\s+-\s+`)

	// Число с последующим словом или единицей измерения (10 шт, 100 руб, 2026 г.)
	reNumbers = regexp.MustCompile(`(\d+)\s+([а-яА-ЯёЁa-zA-Z]+)`)

	// Символ копирайта
	reCopyright = regexp.MustCompile(`\([cCсС]\)`)
)

// Format выполняет серверную типографическую обработку русскоязычного текста:
// 1. Привязывает висячие предлоги и союзы неразрывным пробелом (\u00A0)
// 2. Нормализует парные кавычки в канонические елочки «...»
// 3. Заменяет дефис-тире на длинное тире ( — )
// 4. Привязывает числа к единицам измерения
// 5. Заменяет (c) на символ ©
func Format(text string) string {
	if text == "" {
		return ""
	}

	// 1. Замена (c)/(C) на ©
	text = reCopyright.ReplaceAllString(text, "©")

	// 2. Длинное тире с неразрывным пробелом перед ним: " — "
	text = reDash.ReplaceAllString(text, "\u00A0— ")

	// 3. Привязка чисел к первым словам (например "10 дней", "100 руб")
	text = reNumbers.ReplaceAllString(text, "$1\u00A0$2")

	// 4. Привязка висячих предлогов и союзов (прогоняем трижды для цепочек предлогов типа "и в доме")
	for i := 0; i < 3; i++ {
		text = rePrepositions.ReplaceAllString(text, "$1$2\u00A0")
	}

	// 5. Русские кавычки-елочки «...» вместо машинных "..."
	text = normalizeQuotes(text)

	return strings.TrimSpace(text)
}

// normalizeQuotes преобразует парные прямые кавычки (") в русские кавычки (« и »)
func normalizeQuotes(s string) string {
	if !strings.Contains(s, "\"") {
		return s
	}

	var sb strings.Builder
	sb.Grow(len(s))
	open := false

	runes := []rune(s)
	for i, r := range runes {
		if r == '"' {
			// Проверяем, открывающая или закрывающая кавычка
			if !open {
				// Открывающая: если в начале строки или после пробела/пунктуации
				if i == 0 || isWhitespaceOrPunct(runes[i-1]) {
					sb.WriteRune('«')
					open = true
					continue
				}
			}
			// Закрывающая кавычка
			sb.WriteRune('»')
			open = false
		} else {
			sb.WriteRune(r)
		}
	}

	return sb.String()
}

func isWhitespaceOrPunct(r rune) bool {
	switch r {
	case ' ', '\t', '\n', '\r', '(', '[', '{', '<', '\u00A0':
		return true
	default:
		return false
	}
}
