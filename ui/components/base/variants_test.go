package base

import (
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
)

func TestButtonClasses(t *testing.T) {
	tests := []struct {
		name     string
		props    ButtonProps
		contains []string
	}{
		{
			name: "Default / Primary Button",
			props: ButtonProps{
				Variant: BtnPrimary,
				Size:    BtnSizeDefault,
			},
			contains: []string{"bg-primary", "px-4", "py-2", "text-sm", "rounded-btn"},
		},
		{
			name: "Accent Large Button",
			props: ButtonProps{
				Variant: BtnAccent,
				Size:    BtnSizeLg,
			},
			contains: []string{"bg-accent", "px-6", "py-3", "text-base", "font-bold"},
		},
		{
			name: "Outline Small with Custom Class Override",
			props: ButtonProps{
				Variant: BtnOutline,
				Size:    BtnSizeSm,
				Class:   "w-full text-center",
			},
			contains: []string{"border", "px-3", "py-1.5", "text-xs", "w-full", "text-center"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := tc.props.Classes()
			for _, expected := range tc.contains {
				if !strings.Contains(res, expected) {
					t.Errorf("expected classes to contain %q, got %q", expected, res)
				}
			}
		})
	}
}

func TestBadgeClasses(t *testing.T) {
	props := BadgeProps{
		Variant: BadgeAccent,
		Size:    BadgeSizeSm,
		Class:   "uppercase",
	}

	res := props.Classes()
	expected := []string{"bg-accent/10", "text-accent", "text-[10px]", "uppercase"}
	for _, exp := range expected {
		if !strings.Contains(res, exp) {
			t.Errorf("expected badge classes to contain %q, got %q", exp, res)
		}
	}
}

func TestCardClasses(t *testing.T) {
	props := CardProps{
		Variant: CardInteractive,
		Class:   "hover:scale-[1.01]",
	}

	res := props.Classes()
	expected := []string{"bg-surface", "rounded-card", "hover:shadow-xl", "hover:scale-[1.01]"}
	for _, exp := range expected {
		if !strings.Contains(res, exp) {
			t.Errorf("expected card classes to contain %q, got %q", exp, res)
		}
	}
}

func TestButton_RenderOnClick(t *testing.T) {
	btn := Button(ButtonProps{
		Variant: BtnDestructive,
		Attributes: templ.Attributes{
			"onclick": "window.A11Y.toggle()",
		},
	})
	var buf strings.Builder
	if err := btn.Render(context.Background(), &buf); err != nil {
		t.Fatalf("render error: %v", err)
	}
	out := buf.String()
	if strings.Contains(out, "failed sanitization") {
		t.Errorf("button onclick failed templ sanitization: %s", out)
	}
	if !strings.Contains(out, `onclick="window.A11Y.toggle()"`) {
		t.Errorf("expected onclick=\"window.A11Y.toggle()\" in button, got: %s", out)
	}
}

func TestInputClasses(t *testing.T) {
	props := InputProps{
		Size:        InputSizeDefault,
		Placeholder: "Введите имя",
	}
	res := props.Classes()
	expected := []string{"rounded-input", "bg-surface-muted", "text-content", "border-content/20"}
	for _, exp := range expected {
		if !strings.Contains(res, exp) {
			t.Errorf("expected input classes to contain %q, got %q", exp, res)
		}
	}

	errProps := InputProps{
		ErrorText: "Поле обязательно",
	}
	errRes := errProps.Classes()
	if !strings.Contains(errRes, "border-rose-500") {
		t.Errorf("expected error input to contain border-rose-500, got %q", errRes)
	}
}

func TestModalClasses(t *testing.T) {
	props := ModalProps{
		Size: ModalSizeLg,
	}
	res := props.DialogClasses()
	expected := []string{"bg-surface", "rounded-card", "max-w-lg"}
	for _, exp := range expected {
		if !strings.Contains(res, exp) {
			t.Errorf("expected modal classes to contain %q, got %q", exp, res)
		}
	}
}

func TestInput_RenderWithA11y(t *testing.T) {
	inp := Input(InputProps{
		ID:        "user-email",
		Label:     "Email",
		ErrorText: "Неверный email",
		Required:  true,
	})
	var buf strings.Builder
	if err := inp.Render(context.Background(), &buf); err != nil {
		t.Fatalf("render error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `for="user-email"`) {
		t.Errorf("expected label with for=\"user-email\", got: %s", out)
	}
	if !strings.Contains(out, `aria-invalid="true"`) {
		t.Errorf("expected aria-invalid=\"true\", got: %s", out)
	}
	if !strings.Contains(out, `aria-describedby="user-email-error"`) {
		t.Errorf("expected aria-describedby=\"user-email-error\", got: %s", out)
	}
	if !strings.Contains(out, `role="alert"`) {
		t.Errorf("expected role=\"alert\" on error text, got: %s", out)
	}
}

func TestTextareaClasses(t *testing.T) {
	props := TextareaProps{
		Placeholder: "Комментарий",
	}
	res := props.Classes()
	expected := []string{"rounded-input", "bg-surface-muted", "text-content", "border-content/20", "resize-y"}
	for _, exp := range expected {
		if !strings.Contains(res, exp) {
			t.Errorf("expected textarea classes to contain %q, got %q", exp, res)
		}
	}
}


