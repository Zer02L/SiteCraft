package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"sitecraft/internal/data"
	"sitecraft/internal/models"
	"sitecraft/internal/notifications"
	"sitecraft/ui/pages"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// TMAHandler manages Telegram Mini App authentication, Paywall (Ниша В) and E-commerce (Ниша А).
type TMAHandler struct {
	app *pocketbase.PocketBase
}

// NewTMAHandler creates a new TMAHandler instance.
func NewTMAHandler(app *pocketbase.PocketBase) *TMAHandler {
	return &TMAHandler{app: app}
}

// ValidateTelegramInitData verifies the cryptographic authenticity of Telegram WebApp initData.
func ValidateTelegramInitData(initDataStr string, botToken string) (*models.TMAUser, error) {
	if initDataStr == "" {
		return nil, fmt.Errorf("empty initData")
	}

	values, err := url.ParseQuery(initDataStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse initData: %w", err)
	}

	receivedHash := values.Get("hash")
	if receivedHash == "" {
		return nil, fmt.Errorf("missing hash in initData")
	}

	// Remove hash before creating data_check_string
	values.Del("hash")

	var keys []string
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var checkParts []string
	for _, k := range keys {
		checkParts = append(checkParts, fmt.Sprintf("%s=%s", k, values.Get(k)))
	}
	dataCheckString := strings.Join(checkParts, "\n")

	// If no bot token configured, allow mock/sandbox verification for development
	if botToken == "" {
		// Mock parse user if present
		if userJSON := values.Get("user"); userJSON != "" {
			var u models.TMAUser
			_ = json.Unmarshal([]byte(userJSON), &u)
			return &u, nil
		}
		return &models.TMAUser{
			ID:        999001,
			FirstName: "Telegram",
			LastName:  "User",
			Username:  "tma_user",
		}, nil
	}

	// secret_key = HMAC_SHA256("WebAppData", bot_token)
	hKey := hmac.New(sha256.New, []byte("WebAppData"))
	hKey.Write([]byte(botToken))
	secretKey := hKey.Sum(nil)

	// expected_hash = HMAC_SHA256(data_check_string, secret_key)
	hData := hmac.New(sha256.New, secretKey)
	hData.Write([]byte(dataCheckString))
	expectedHash := hex.EncodeToString(hData.Sum(nil))

	if !hmac.Equal([]byte(receivedHash), []byte(expectedHash)) {
		return nil, fmt.Errorf("invalid hash signature")
	}

	// Parse user data from payload
	userJSON := values.Get("user")
	if userJSON == "" {
		return nil, fmt.Errorf("missing user in initData")
	}

	var user models.TMAUser
	if err := json.Unmarshal([]byte(userJSON), &user); err != nil {
		return nil, fmt.Errorf("failed to decode user json: %w", err)
	}

	return &user, nil
}

// HandleTMAAuth verifies initData and returns user session data.
func (h *TMAHandler) HandleTMAAuth(e *core.RequestEvent) error {
	settings := data.GetSiteSettings(h.app)

	var req struct {
		InitData string `json:"initData"`
	}
	if err := json.NewDecoder(e.Request.Body).Decode(&req); err != nil {
		return e.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid JSON body"})
	}

	user, err := ValidateTelegramInitData(req.InitData, settings.TelegramBotToken)
	if err != nil {
		return e.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}

	// Check if user has VIP status (Ниша В)
	// Check leads/subscribers for this telegram_id
	user.IsVIP = h.checkVIPStatus(user.ID)

	// Set session cookie
	http.SetCookie(e.Response, &http.Cookie{
		Name:     "tma_session",
		Value:    fmt.Sprintf("%d:%s", user.ID, user.Username),
		Path:     "/",
		MaxAge:   86400 * 30,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	return e.JSON(http.StatusOK, user)
}

func (h *TMAHandler) checkVIPStatus(telegramID int64) bool {
	if telegramID <= 0 {
		return false
	}
	// Check if lead/payment exists with this telegramID and status "оплачена"
	filter := fmt.Sprintf("payload.telegram_id = '%d' && status = 'оплачена'", telegramID)
	rec, err := h.app.FindFirstRecordByFilter("leads", filter)
	return err == nil && rec != nil
}

// HandleVIPClub serves Ниша В (Paywall / Закрытый клуб / TMA видео-контент).
func (h *TMAHandler) HandleVIPClub(e *core.RequestEvent) error {
	settings := data.GetSiteSettings(h.app)
	content := data.GetSiteContent(h.app)

	// Check session cookie
	var user *models.TMAUser
	cookie, err := e.Request.Cookie("tma_session")
	if err == nil && cookie.Value != "" {
		parts := strings.Split(cookie.Value, ":")
		if len(parts) >= 2 {
			var tid int64
			_, _ = fmt.Sscanf(parts[0], "%d", &tid)
			user = &models.TMAUser{
				ID:        tid,
				Username:  parts[1],
				FirstName: parts[1],
				IsVIP:     h.checkVIPStatus(tid),
			}
		}
	}

	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return pages.VIPClub(settings, content, user).Render(e.Request.Context(), e.Response)
}

// HandleShop serves Ниша А (Микро-E-commerce и Доставка для TMA).
func (h *TMAHandler) HandleShop(e *core.RequestEvent) error {
	settings := data.GetSiteSettings(h.app)
	content := data.GetSiteContent(h.app)
	services := data.GetServices(h.app)

	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return pages.ShopCatalog(settings, content, services).Render(e.Request.Context(), e.Response)
}

// HandleShopCheckout processes orders from TMA Cart (Ниша А).
func (h *TMAHandler) HandleShopCheckout(e *core.RequestEvent) error {
	settings := data.GetSiteSettings(h.app)

	var req struct {
		Name         string   `json:"name"`
		Phone        string   `json:"phone"`
		Address      string   `json:"address"`
		DeliveryType string   `json:"delivery_type"` // delivery, pickup
		TelegramUser string   `json:"telegram_user"`
		TelegramID   int64    `json:"telegram_id"`
		Items        []string `json:"items"`
		TotalAmount  float64  `json:"total_amount"`
		Comment      string   `json:"comment"`
	}

	if err := json.NewDecoder(e.Request.Body).Decode(&req); err != nil {
		return e.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	if req.Phone == "" {
		return e.JSON(http.StatusBadRequest, map[string]string{"error": "Номер телефона обязателен"})
	}

	// Create record in leads collection
	leadsCol, err := h.app.FindCollectionByNameOrId("leads")
	if err != nil {
		return e.JSON(http.StatusInternalServerError, map[string]string{"error": "Database error"})
	}

	rec := core.NewRecord(leadsCol)
	rec.Set("name", req.Name)
	rec.Set("phone", req.Phone)
	rec.Set("source", "TMA E-commerce / Shop")
	rec.Set("status", "новая")

	payload := map[string]interface{}{
		"action_type":   "shop_order",
		"address":       req.Address,
		"delivery_type": req.DeliveryType,
		"telegram_user": req.TelegramUser,
		"telegram_id":   fmt.Sprintf("%d", req.TelegramID),
		"items":         req.Items,
		"total_amount":  req.TotalAmount,
		"comment":       req.Comment,
		"created_at":    time.Now().Format("2006-01-02 15:04:05"),
	}
	rec.Set("payload", payload)

	if err := h.app.Save(rec); err != nil {
		return e.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save order"})
	}

	leadModel := models.Lead{
		ID:      rec.Id,
		Name:    req.Name,
		Phone:   req.Phone,
		Source:  "TMA E-commerce / Shop",
		Status:  "новая",
		Payload: payload,
	}

	notifications.SendLeadNotification(settings, leadModel)

	return e.JSON(http.StatusOK, map[string]interface{}{
		"success":  true,
		"order_id": rec.Id,
		"message":  "Заказ успешно оформлен",
	})
}
