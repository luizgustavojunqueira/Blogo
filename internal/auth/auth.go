package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Auth struct {
	Username      string
	Password      string
	SecretKey     string
	TokenValidity int64
	CookieName    string
}

func NewAuth(username, password, secretKey, cookieName string, tokenValidity int64) (*Auth, error) {
	if username == "" || password == "" || secretKey == "" || cookieName == "" || tokenValidity == 0 {
		return nil, fmt.Errorf("Invalid parameters")
	}

	if tokenValidity < 60 {
		return nil, fmt.Errorf("Token validity must be at least 60 seconds")
	}

	if len(secretKey) < 32 {
		return nil, fmt.Errorf("Secret key must be at least 32 characters")
	}

	if len(cookieName) < 8 {
		return nil, fmt.Errorf("Cookie name must be at least 8 characters")
	}

	if strings.Contains(cookieName, ":") {
		return nil, fmt.Errorf("Cookie name cannot contain ':'")
	}

	if strings.Contains(username, ":") {
		return nil, fmt.Errorf("Username cannot contain ':'")
	}

	if username == password {
		return nil, fmt.Errorf("Username and password must be different")
	}

	if len(password) < 8 {
		return nil, fmt.Errorf("Password must be at least 8 characters")
	}

	if len(username) < 4 {
		return nil, fmt.Errorf("Username must be at least 4 characters")
	}

	return &Auth{
		Username:      username,
		Password:      password,
		SecretKey:     secretKey,
		CookieName:    cookieName,
		TokenValidity: tokenValidity,
	}, nil
}

func (auth *Auth) GenerateToken(username string, expiry int64) string {
	data := fmt.Sprintf("%s:%d", username, expiry)
	h := hmac.New(sha256.New, []byte(auth.SecretKey))
	h.Write([]byte(data))
	signature := hex.EncodeToString(h.Sum(nil))
	token := fmt.Sprintf("%s:%s", data, signature)
	return token
}

func (auth *Auth) ValidateToken(token string) (bool, error) {
	if token == "" {
		return false, fmt.Errorf("Empty Token")
	}
	parts := strings.Split(token, ":")
	if len(parts) != 3 {
		return false, fmt.Errorf("Invalid Token")
	}
	username := parts[0]
	expiryStr := parts[1]
	signatureProvided := parts[2]

	expiry, err := strconv.ParseInt(expiryStr, 10, 64)
	if err != nil {
		return false, fmt.Errorf("Invalid expiry")
	}

	if time.Now().Unix() > expiry {
		return false, fmt.Errorf("Expired Token")
	}

	// Recalcula a assinatura esperada
	data := fmt.Sprintf("%s:%s", username, expiryStr)
	h := hmac.New(sha256.New, []byte(auth.SecretKey))
	h.Write([]byte(data))
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	if !hmac.Equal([]byte(signatureProvided), []byte(expectedSignature)) {
		return false, fmt.Errorf("Invalid Signature")
	}

	return true, nil
}
