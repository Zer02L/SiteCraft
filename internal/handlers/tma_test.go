package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"testing"
)

func generateValidInitData(botToken string, userID int64, username string) string {
	values := url.Values{}
	values.Set("auth_date", "1710000000")
	values.Set("query_id", "AAG_test")
	values.Set("user", fmt.Sprintf(`{"id":%d,"first_name":"Test","last_name":"User","username":"%s"}`, userID, username))

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

	// secret_key = HMAC_SHA256("WebAppData", bot_token)
	hKey := hmac.New(sha256.New, []byte("WebAppData"))
	hKey.Write([]byte(botToken))
	secretKey := hKey.Sum(nil)

	// hash = HMAC_SHA256(dataCheckString, secretKey)
	hData := hmac.New(sha256.New, secretKey)
	hData.Write([]byte(dataCheckString))
	hash := hex.EncodeToString(hData.Sum(nil))

	values.Set("hash", hash)
	return values.Encode()
}

func TestValidateTelegramInitData_Success(t *testing.T) {
	botToken := "123456789:ABCdefGHIjklMNOpqrSTUvwxYZ"
	initData := generateValidInitData(botToken, 1234567, "testuser")

	user, err := ValidateTelegramInitData(initData, botToken)
	if err != nil {
		t.Fatalf("Expected valid initData, got error: %v", err)
	}

	if user.ID != 1234567 {
		t.Errorf("Expected user ID 1234567, got %d", user.ID)
	}
	if user.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got %s", user.Username)
	}
}

func TestValidateTelegramInitData_TamperedHash(t *testing.T) {
	botToken := "123456789:ABCdefGHIjklMNOpqrSTUvwxYZ"
	initData := generateValidInitData(botToken, 1234567, "testuser")

	// Tamper with data
	tampered := initData + "&tampered=1"

	_, err := ValidateTelegramInitData(tampered, botToken)
	if err == nil {
		t.Fatal("Expected error for tampered initData, but got nil")
	}
}

func TestValidateTelegramInitData_MockMode(t *testing.T) {
	// When botToken is empty, sandbox mock mode should succeed
	mockData := "auth_date=1710000000&hash=mockhash&user=%7B%22id%22%3A888%2C%22username%22%3A%22mockuser%22%7D"
	user, err := ValidateTelegramInitData(mockData, "")
	if err != nil {
		t.Fatalf("Expected mock mode to succeed, got %v", err)
	}
	if user.ID != 888 {
		t.Errorf("Expected user ID 888, got %d", user.ID)
	}
}
