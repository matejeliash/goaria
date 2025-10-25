package session

import (
	"net/http"

	"github.com/gorilla/sessions"
)

var (
	store *sessions.CookieStore // store for easy cookie creation and retrieval
)

func init() {
	// !!!! later change for ENV variable
	store = sessions.NewCookieStore([]byte("later-will-be-replaced"))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   60 * 60 * 24 * 7, // now 7 days
		HttpOnly: true,
		Secure:   false, // temporarily false for localhost
		SameSite: http.SameSiteStrictMode,
	}
}

func GetStore() *sessions.CookieStore {
	return store
}

func GetSessionName() string {
	return "goaria-session"
}
