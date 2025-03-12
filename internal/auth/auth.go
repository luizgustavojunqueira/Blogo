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
	username      string
	password      string
	secretKey     string
	tokenValidity int64
	cookieName    string
}

type AuthConfig struct {
	Username      string // Username for authentication, at least 4 characters
	Password      string // Password for authentication, at least 8 characters
	SecretKey     string // Secret key for token generation, at least 32 characters
	TokenValidity int64  // Token validity in seconds, at least 60 seconds
	CookieName    string // Name of the cookie, at least 8 characters
}

// NewAuth creates a new Auth instance from the provided configuration.
// It returns an error if the configuration is invalid.
func NewAuth(config AuthConfig) (*Auth, error) {
	if config.Username == "" || config.Password == "" || config.SecretKey == "" || config.CookieName == "" || config.TokenValidity == 0 {
		return nil, fmt.Errorf("Invalid parameters")
	}

	if config.TokenValidity < 60 {
		return nil, fmt.Errorf("Token validity must be at least 60 seconds")
	}

	if len(config.SecretKey) < 32 {
		return nil, fmt.Errorf("Secret key must be at least 32 characters")
	}

	if len(config.CookieName) < 8 {
		return nil, fmt.Errorf("Cookie name must be at least 8 characters")
	}

	if strings.Contains(config.CookieName, ":") {
		return nil, fmt.Errorf("Cookie name cannot contain ':'")
	}

	if strings.Contains(config.Username, ":") {
		return nil, fmt.Errorf("Username cannot contain ':'")
	}

	if config.Username == config.Password {
		return nil, fmt.Errorf("Username and password must be different")
	}

	if len(config.Password) < 8 {
		return nil, fmt.Errorf("Password must be at least 8 characters")
	}

	if len(config.Username) < 4 {
		return nil, fmt.Errorf("Username must be at least 4 characters")
	}

	return &Auth{
		username:      config.Username,
		password:      config.Password,
		secretKey:     config.SecretKey,
		tokenValidity: config.TokenValidity,
		cookieName:    config.CookieName,
	}, nil
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

func (auth *Auth) ValidateCredentials(username, password string) bool {
	return username == auth.username && password == auth.password
}

func (auth *Auth) GetCookieName() string {
	return auth.cookieName
}

func (auth *Auth) GetTokenValidity() int64 {
	return auth.tokenValidity
}
