package typography

import (
	"strings"
	"testing"
)

func TestFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Prepositions with non-breaking spaces",
			input:    "Юридические услуги в Москве и по всей России",
			expected: "Юридические услуги в\u00A0Москве и\u00A0по\u00A0всей России",
		},
		{
			name:     "Quotes normalization",
			input:    `Компания "Вектор" гарантирует качество`,
			expected: "Компания «Вектор» гарантирует качество",
		},
		{
			name:     "Dash normalization",
			input:    "Сайт - это визитная карточка бизнеса",
			expected: "Сайт\u00A0— это визитная карточка бизнеса",
		},
		{
			name:     "Numbers with units",
			input:    "Гарантия 10 лет на 500 часов",
			expected: "Гарантия 10\u00A0лет на\u00A0500\u00A0часов",
		},
		{
			name:     "Copyright symbol",
			input:    "Все права защищены (c) 2026",
			expected: "Все права защищены © 2026",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Format(tt.input)
			if got != tt.expected {
				t.Errorf("Format(%q) =\n%q, want:\n%q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestFormat_Empty(t *testing.T) {
	if got := Format(""); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestFormat_ChainPrepositions(t *testing.T) {
	got := Format("Приходите к нам и в офис")
	if !strings.Contains(got, "к\u00A0нам") || !strings.Contains(got, "в\u00A0офис") {
		t.Errorf("expected chained prepositions with nbsp, got %q", got)
	}
}
