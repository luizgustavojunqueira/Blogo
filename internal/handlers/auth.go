package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/luizgustavojunqueira/Blogo/internal/auth"
	"github.com/luizgustavojunqueira/Blogo/internal/templates"
)

type AuthHandler struct {
	auth     *auth.Auth
	logger   *log.Logger
	blogName string
}

func NewAuthHandler(auth *auth.Auth, logger *log.Logger, blogName string) *AuthHandler {
	return &AuthHandler{
		auth:     auth,
		logger:   logger,
		blogName: blogName,
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {

		ctx := r.Context()

		page := templates.LoginPage(h.blogName)
		page.Render(ctx, w)

		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	if username != h.auth.Username || password != h.auth.Password {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	expiry := time.Now().Unix() + h.auth.TokenValidity

	token := h.auth.GenerateToken(username, expiry)

	cookie := http.Cookie{
		Name:     h.auth.CookieName,
		Value:    token,
		Expires:  time.Unix(expiry, 0),
		HttpOnly: true,
	}

	http.SetCookie(w, &cookie)
	w.Header().Set("HX-Location", "/")
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie := http.Cookie{
		Name:   "session",
		MaxAge: -1,
	}

	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
