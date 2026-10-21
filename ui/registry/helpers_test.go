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

func TestParsePayload_Generic(t *testing.T) {
	// 1. Nil payload
	if res := ParsePayload[struct{ Name string }](nil); res != nil {
		t.Errorf("expected nil result for nil payload, got %+v", res)
	}

	// 2. Valid payload map
	payload := map[string]interface{}{
		"title": "Super Features",
		"badge": "NEW",
	}
	type customStruct struct {
		Title string `json:"title"`
		Badge string `json:"badge"`
	}
	res := ParsePayload[customStruct](payload)
	if res == nil {
		t.Fatal("expected non-nil struct for valid payload")
	}
	if res.Title != "Super Features" || res.Badge != "NEW" {
		t.Errorf("unexpected parsed struct: %+v", res)
	}
}
