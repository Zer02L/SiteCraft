package handlers

import (
	"strings"
	"testing"
	"time"

	"sitecraft/internal/models"
)

func TestChatHubBroadcast(t *testing.T) {
	hub := NewChatHub()
	sessionID := "test_sess_123"

	ch1 := hub.Subscribe(sessionID)
	ch2 := hub.Subscribe(sessionID)
	defer hub.Unsubscribe(sessionID, ch1)
	defer hub.Unsubscribe(sessionID, ch2)

	msg := models.ChatMessage{
		SessionID: sessionID,
		Sender:    "operator",
		Message:   "Здравствуйте! Чем помочь?",
		CreatedAt: time.Now(),
		Created:   "12:00",
	}

	hub.Broadcast(msg)

	select {
	case received := <-ch1:
		if received.Message != msg.Message {
			t.Errorf("Expected message %q, got %q", msg.Message, received.Message)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Timeout waiting for ch1 broadcast")
	}

	select {
	case received := <-ch2:
		if received.Message != msg.Message {
			t.Errorf("Expected message %q, got %q", msg.Message, received.Message)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Timeout waiting for ch2 broadcast")
	}
}

func TestSessionIDRegexExtraction(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"Сессия: cs_abc123\nТекст сообщения", "cs_abc123"},
		{"💬 Чат с сайта\nСессия: #cs_9988\nПосетитель: Тест", "cs_9988"},
		{"сессия: sess_42\nВопрос", "sess_42"},
		{"Просто текст без сессии", ""},
	}

	for _, tc := range cases {
		match := sessionIDRegex.FindStringSubmatch(tc.input)
		var got string
		if len(match) > 1 {
			got = match[1]
		}
		if got != tc.expected {
			t.Errorf("For input %q: expected %q, got %q", tc.input, tc.expected, got)
		}
	}
}

func TestRenderChatBubbleHTML(t *testing.T) {
	visitorMsg := models.ChatMessage{
		SessionID: "sess_1",
		Sender:    "visitor",
		Message:   "<script>alert(1)</script>Привет",
		Created:   "14:20",
	}

	htmlOut := renderChatBubbleHTML(visitorMsg)
	if strings.Contains(htmlOut, "<script>") {
		t.Error("XSS script was not escaped in visitor chat bubble")
	}
	if !strings.Contains(htmlOut, "&lt;script&gt;") {
		t.Error("Expected escaped script tags in visitor bubble")
	}
	if !strings.Contains(htmlOut, "14:20") {
		t.Error("Expected timestamp in bubble")
	}

	operatorMsg := models.ChatMessage{
		SessionID: "sess_1",
		Sender:    "operator",
		Message:   "Ответ оператора",
		Created:   "14:21",
	}

	opHTML := renderChatBubbleHTML(operatorMsg)
	if !strings.Contains(opHTML, "Оператор") {
		t.Error("Expected Operator badge in bubble")
	}
	if !strings.Contains(opHTML, "Ответ оператора") {
		t.Error("Expected text in operator bubble")
	}
}
