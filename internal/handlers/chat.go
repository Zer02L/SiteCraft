package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"sitecraft/internal/data"
	"sitecraft/internal/models"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

var (
	sessionIDRegex = regexp.MustCompile(`(?i)Сессия:\s*#?([a-zA-Z0-9_\-]+)`)
	httpClientChat = &http.Client{Timeout: 10 * time.Second}
)

// ChatHub coordinates SSE client connections grouped by session ID.
type ChatHub struct {
	mu      sync.RWMutex
	clients map[string]map[chan models.ChatMessage]bool
}

// NewChatHub initializes a ChatHub instance.
func NewChatHub() *ChatHub {
	return &ChatHub{
		clients: make(map[string]map[chan models.ChatMessage]bool),
	}
}

// Subscribe attaches a new channel listener for the session.
func (h *ChatHub) Subscribe(sessionID string) chan models.ChatMessage {
	h.mu.Lock()
	defer h.mu.Unlock()

	ch := make(chan models.ChatMessage, 16)
	if _, ok := h.clients[sessionID]; !ok {
		h.clients[sessionID] = make(map[chan models.ChatMessage]bool)
	}
	h.clients[sessionID][ch] = true
	return ch
}

// Unsubscribe removes the listener channel.
func (h *ChatHub) Unsubscribe(sessionID string, ch chan models.ChatMessage) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if subs, ok := h.clients[sessionID]; ok {
		delete(subs, ch)
		close(ch)
		if len(subs) == 0 {
			delete(h.clients, sessionID)
		}
	}
}

// Broadcast dispatches a message to all active listeners for the session.
func (h *ChatHub) Broadcast(msg models.ChatMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	subs, ok := h.clients[msg.SessionID]
	if !ok {
		return
	}
	for ch := range subs {
		select {
		case ch <- msg:
		default:
			// Buffer full, skip to prevent blocking
		}
	}
}

// ChatHandler manages the Live Web-to-Telegram Chat module.
type ChatHandler struct {
	app      *pocketbase.PocketBase
	hub      *ChatHub
	pollerMu sync.Mutex
	pollerOn bool
}

// NewChatHandler creates a new ChatHandler instance.
func NewChatHandler(app *pocketbase.PocketBase) *ChatHandler {
	return &ChatHandler{
		app: app,
		hub: NewChatHub(),
	}
}

// StartBackgroundPoller starts Telegram long-polling if a bot token is configured.
func (h *ChatHandler) StartBackgroundPoller(ctx context.Context) {
	h.pollerMu.Lock()
	if h.pollerOn {
		h.pollerMu.Unlock()
		return
	}
	h.pollerOn = true
	h.pollerMu.Unlock()

	go func() {
		var lastUpdateID int64
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				settings := data.GetSiteSettings(h.app)
				if !settings.EnableLiveChat || settings.TelegramBotToken == "" {
					continue
				}

				updates, maxID, err := fetchTelegramUpdates(settings.TelegramBotToken, lastUpdateID+1)
				if err != nil {
					continue
				}
				if maxID > lastUpdateID {
					lastUpdateID = maxID
				}

				for _, u := range updates {
					h.processTelegramUpdate(u)
				}
			}
		}
	}()
}

// HandleChatSSE streams live messages from operator to the visitor's browser.
func (h *ChatHandler) HandleChatSSE(e *core.RequestEvent) error {
	sessionID := strings.TrimSpace(e.Request.URL.Query().Get("session_id"))
	if sessionID == "" {
		return e.HTML(http.StatusBadRequest, "Missing session_id")
	}

	w := e.Response
	flusher, ok := w.(http.Flusher)
	if !ok {
		return e.HTML(http.StatusInternalServerError, "SSE unsupported")
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	ch := h.hub.Subscribe(sessionID)
	defer h.hub.Unsubscribe(sessionID, ch)

	// Send initial connection comment
	_, _ = fmt.Fprintf(w, ": connected\n\n")
	flusher.Flush()

	heartbeat := time.NewTicker(20 * time.Second)
	defer heartbeat.Stop()

	ctx := e.Request.Context()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-heartbeat.C:
			if _, err := fmt.Fprintf(w, ": ping\n\n"); err != nil {
				return nil
			}
			flusher.Flush()
		case msg, ok := <-ch:
			if !ok {
				return nil
			}
			bubbleHTML := renderChatBubbleHTML(msg)
			// Format as SSE event
			payload := fmt.Sprintf("event: chat-message\ndata: %s\n\n", strings.ReplaceAll(bubbleHTML, "\n", ""))
			if _, err := fmt.Fprint(w, payload); err != nil {
				return nil
			}
			flusher.Flush()
		}
	}
}

// HandleSendMessage processes messages submitted from the website chat widget.
func (h *ChatHandler) HandleSendMessage(e *core.RequestEvent) error {
	sessionID := strings.TrimSpace(e.Request.FormValue("session_id"))
	text := strings.TrimSpace(e.Request.FormValue("message"))
	visitorName := strings.TrimSpace(e.Request.FormValue("name"))

	if sessionID == "" {
		sessionID = fmt.Sprintf("cs_%d", time.Now().UnixNano()%1000000)
	}
	if text == "" {
		return e.HTML(http.StatusBadRequest, `<div class="text-xs text-rose-500">Сообщение не может быть пустым</div>`)
	}
	if len(text) > 2000 {
		text = text[:2000]
	}
	if visitorName == "" {
		visitorName = "Посетитель"
	}

	// 1. Save visitor message in PocketBase
	msg := models.ChatMessage{
		SessionID: sessionID,
		Sender:    "visitor",
		Message:   text,
		CreatedAt: time.Now(),
		Created:   time.Now().Format("15:04"),
	}

	var tgMsgID int64
	settings := data.GetSiteSettings(h.app)

	// 2. Dispatch to Telegram
	if settings.TelegramBotToken != "" && settings.TelegramChatID != "" {
		tgID, err := sendTelegramChatMessage(settings.TelegramBotToken, settings.TelegramChatID, sessionID, visitorName, text)
		if err == nil {
			tgMsgID = tgID
			msg.TgMsgID = tgMsgID
		}
	}

	if col, err := h.app.FindCollectionByNameOrId("chat_messages"); err == nil {
		rec := core.NewRecord(col)
		rec.Set("session_id", sessionID)
		rec.Set("sender", "visitor")
		rec.Set("message", text)
		rec.Set("tg_msg_id", tgMsgID)
		_ = h.app.Save(rec)
		msg.ID = rec.Id
	}

	// 3. Broadcast to any other listeners of this session
	h.hub.Broadcast(msg)

	// 4. Return HTML fragment for immediate insertion into the chat list
	return e.HTML(http.StatusOK, renderChatBubbleHTML(msg))
}

// HandleChatHistory returns the formatted message history for a session.
func (h *ChatHandler) HandleChatHistory(e *core.RequestEvent) error {
	sessionID := strings.TrimSpace(e.Request.URL.Query().Get("session_id"))
	if sessionID == "" {
		return e.HTML(http.StatusOK, "")
	}

	records, err := h.app.FindRecordsByFilter("chat_messages", "session_id = {:sid}", "created", 100, 0, dbx.Params{"sid": sessionID})
	if err != nil || len(records) == 0 {
		return e.HTML(http.StatusOK, `
			<div class="py-6 text-center text-xs text-content-muted">
				<p class="font-medium">Здравствуйте! Чем можем вам помочь?</p>
				<p class="text-[11px] opacity-75 mt-1">Напишите нам, оператор ответит в течение 1–2 минут.</p>
			</div>
		`)
	}

	var buf bytes.Buffer
	for _, rec := range records {
		createdTime := rec.GetDateTime("created").Time().Format("15:04")
		msg := models.ChatMessage{
			ID:        rec.Id,
			SessionID: sessionID,
			Sender:    rec.GetString("sender"),
			Message:   rec.GetString("message"),
			TgMsgID:   int64(rec.GetInt("tg_msg_id")),
			Created:   createdTime,
		}
		buf.WriteString(renderChatBubbleHTML(msg))
	}

	return e.HTML(http.StatusOK, buf.String())
}

// HandleTelegramWebhook handles inbound updates from Telegram Webhook.
func (h *ChatHandler) HandleTelegramWebhook(e *core.RequestEvent) error {
	var update tgUpdate
	if err := json.NewDecoder(e.Request.Body).Decode(&update); err != nil {
		return e.HTML(http.StatusBadRequest, "Invalid JSON")
	}

	h.processTelegramUpdate(update)
	return e.HTML(http.StatusOK, "OK")
}

func (h *ChatHandler) processTelegramUpdate(u tgUpdate) {
	if u.Message == nil || u.Message.ReplyToMessage == nil || strings.TrimSpace(u.Message.Text) == "" {
		return
	}

	replyTo := u.Message.ReplyToMessage
	replyText := replyTo.Text
	if replyText == "" {
		replyText = replyTo.Caption
	}

	// 1. Try finding session ID from message text regex
	var sessionID string
	match := sessionIDRegex.FindStringSubmatch(replyText)
	if len(match) > 1 {
		sessionID = match[1]
	}

	// 2. Fallback to searching database by replied tg_msg_id
	if sessionID == "" && replyTo.MessageID > 0 {
		rec, _ := h.app.FindFirstRecordByFilter("chat_messages", "tg_msg_id = {:msg_id}", dbx.Params{"msg_id": replyTo.MessageID})
		if rec != nil {
			sessionID = rec.GetString("session_id")
		}
	}

	if sessionID == "" {
		return
	}

	operatorText := strings.TrimSpace(u.Message.Text)

	// Save operator message to database
	msg := models.ChatMessage{
		SessionID: sessionID,
		Sender:    "operator",
		Message:   operatorText,
		CreatedAt: time.Now(),
		Created:   time.Now().Format("15:04"),
	}

	if col, err := h.app.FindCollectionByNameOrId("chat_messages"); err == nil {
		rec := core.NewRecord(col)
		rec.Set("session_id", sessionID)
		rec.Set("sender", "operator")
		rec.Set("message", operatorText)
		rec.Set("tg_msg_id", u.Message.MessageID)
		_ = h.app.Save(rec)
		msg.ID = rec.Id
	}

	// Broadcast to visitor's active browser session via SSE!
	h.hub.Broadcast(msg)
}

func renderChatBubbleHTML(msg models.ChatMessage) string {
	escapedText := html.EscapeString(msg.Message)
	// Replace linebreaks with <br/>
	formattedText := strings.ReplaceAll(escapedText, "\n", "<br/>")

	if msg.Created == "" {
		msg.Created = time.Now().Format("15:04")
	}

	if msg.Sender == "visitor" {
		return fmt.Sprintf(`
			<div class="flex justify-end mb-3">
				<div class="max-w-[85%%] rounded-2xl rounded-tr-sm bg-accent text-primary-dark px-3.5 py-2.5 shadow-sm text-xs font-medium">
					<div>%s</div>
					<div class="text-[10px] opacity-70 text-right mt-1 font-mono">%s</div>
				</div>
			</div>
		`, formattedText, msg.Created)
	}

	return fmt.Sprintf(`
		<div class="flex justify-start mb-3">
			<div class="max-w-[85%%] rounded-2xl rounded-tl-sm bg-surface-muted border border-content/10 text-content px-3.5 py-2.5 shadow-sm text-xs">
				<div class="text-[10px] font-bold text-accent mb-0.5">Оператор</div>
				<div>%s</div>
				<div class="text-[10px] text-content-muted text-right mt-1 font-mono">%s</div>
			</div>
		</div>
	`, formattedText, msg.Created)
}

// Telegram Bot API models and helpers
type tgUpdate struct {
	UpdateID int64      `json:"update_id"`
	Message  *tgMessage `json:"message"`
}

type tgMessage struct {
	MessageID      int64      `json:"message_id"`
	Text           string     `json:"text"`
	Caption        string     `json:"caption"`
	ReplyToMessage *tgMessage `json:"reply_to_message"`
}

type tgSendMessageResp struct {
	Ok     bool       `json:"ok"`
	Result *tgMessage `json:"result"`
}

type tgGetUpdatesResp struct {
	Ok     bool       `json:"ok"`
	Result []tgUpdate `json:"result"`
}

func sendTelegramChatMessage(token, chatID, sessionID, visitorName, text string) (int64, error) {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
	tgText := fmt.Sprintf(
		"💬 <b>Новое сообщение в чате сайта!</b>\n\n"+
			"🆔 <b>Сессия:</b> <code>%s</code>\n"+
			"👤 <b>%s:</b> %s\n\n"+
			"👉 <i>Чтобы ответить посетителю на сайт, сделайте Reply на это сообщение.</i>",
		sessionID, html.EscapeString(visitorName), html.EscapeString(text),
	)

	bodyData := map[string]interface{}{
		"chat_id":    chatID,
		"text":       tgText,
		"parse_mode": "HTML",
	}

	b, _ := json.Marshal(bodyData)
	resp, err := httpClientChat.Post(apiURL, "application/json", bytes.NewReader(b))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var res tgSendMessageResp
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return 0, err
	}
	if res.Ok && res.Result != nil {
		return res.Result.MessageID, nil
	}
	return 0, fmt.Errorf("telegram send error")
}

func fetchTelegramUpdates(token string, offset int64) ([]tgUpdate, int64, error) {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates?offset=%d&timeout=15", token, offset)
	resp, err := httpClientChat.Get(apiURL)
	if err != nil {
		return nil, offset, err
	}
	defer resp.Body.Close()

	var res tgGetUpdatesResp
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, offset, err
	}

	var maxID int64 = offset
	for _, u := range res.Result {
		if u.UpdateID > maxID {
			maxID = u.UpdateID
		}
	}
	return res.Result, maxID, nil
}
