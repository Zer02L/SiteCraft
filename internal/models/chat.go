package models

import "time"

// ChatMessage represents a single message in the Live Web-to-TG Chat.
type ChatMessage struct {
	ID        string    `json:"id"`
	SessionID string    `json:"session_id"`
	Sender    string    `json:"sender"` // "visitor" or "operator"
	Message   string    `json:"message"`
	TgMsgID   int64     `json:"tg_msg_id"`
	CreatedAt time.Time `json:"created_at"`
	Created   string    `json:"created"` // Formatted time for display
}
