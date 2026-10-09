package registry

import (
	"testing"
)

func TestParseQuizPayload_Defaults(t *testing.T) {
	autoQuiz := ParseQuizPayload(nil, "auto")
	if autoQuiz.Badge == "" || len(autoQuiz.Steps) == 0 {
		t.Errorf("expected auto quiz defaults, got empty")
	}
	if autoQuiz.Steps[0].Options[0].Title == "" {
		t.Errorf("expected valid quiz options")
	}

	beautyQuiz := ParseQuizPayload(nil, "beauty")
	if beautyQuiz.Badge != "Бьюти-калькулятор" {
		t.Errorf("expected beauty quiz badge, got %s", beautyQuiz.Badge)
	}

	legalQuiz := ParseQuizPayload(nil, "legal")
	if legalQuiz.Badge != "Экспресс-оценка задачи" {
		t.Errorf("expected legal quiz badge, got %s", legalQuiz.Badge)
	}
}

func TestParseBeforeAfterPayload_Defaults(t *testing.T) {
	autoBA := ParseBeforeAfterPayload(nil, "auto")
	if len(autoBA.Items) == 0 {
		t.Errorf("expected auto before/after items")
	}
	if autoBA.Items[0].BeforeImage == "" || autoBA.Items[0].AfterImage == "" {
		t.Errorf("expected valid images in before/after item")
	}

	beautyBA := ParseBeforeAfterPayload(nil, "beauty")
	if len(beautyBA.Items) == 0 {
		t.Errorf("expected beauty before/after items")
	}
}
