package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/luizgustavojunqueira/Blogo/internal/auth"
	"github.com/luizgustavojunqueira/Blogo/internal/templates/pages"
)

type AuthHandler struct {
	auth      *auth.Auth
	logger    *log.Logger
	blogName  string
	pagetitle string
}

func NewAuthHandler(auth *auth.Auth, logger *log.Logger, blogName, pagetitle string) *AuthHandler {
	return &AuthHandler{
		auth:      auth,
		logger:    logger,
		blogName:  blogName,
		pagetitle: pagetitle,
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {

		ctx := r.Context()

		loginPage := pages.LoginPage(h.blogName, h.pagetitle)

		page := pages.Root(h.blogName, loginPage)
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

	if h.auth.ValidateCredentials(username, password) == false {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	expiry := time.Now().Unix() + h.auth.GetTokenValidity()

	token := h.auth.GenerateToken(username, expiry)

	cookie := http.Cookie{
		Name:     h.auth.GetCookieName(),
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
