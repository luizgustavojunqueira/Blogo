package handlers

import (
	"net/http"
	"time"

	"github.com/luizgustavojunqueira/Blog/internal/templates"
)

type AuthHandler struct {
	username string
	password string
}

func NewAuthHandler(username, password string) *AuthHandler {
	return &AuthHandler{
		username: username,
		password: password,
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {

		ctx := r.Context()

		page := templates.LoginPage()
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

	if username != h.username || password != h.password {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	expiration := time.Now().Add(1 * time.Minute)
	cookie := http.Cookie{
		Name:     "session",
		Value:    "authenticated",
		Expires:  expiration,
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
