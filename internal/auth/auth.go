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
	secretKey     string
	TokenValidity int64
	CookieName    string
}

func NewAuth(username, password, secretKey, cookieName string, tokenValidity int64) *Auth {
	return &Auth{
		Username:      username,
		Password:      password,
		secretKey:     secretKey,
		CookieName:    cookieName,
		TokenValidity: tokenValidity,
	}
}

func (auth *Auth) GenerateToken(username string, expiry int64) string {
	data := fmt.Sprintf("%s:%d", username, expiry)
	h := hmac.New(sha256.New, []byte(auth.secretKey))
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
	h := hmac.New(sha256.New, []byte(auth.secretKey))
	h.Write([]byte(data))
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	if !hmac.Equal([]byte(signatureProvided), []byte(expectedSignature)) {
		return false, fmt.Errorf("Invalid Signature")
	}

	return true, nil
}
