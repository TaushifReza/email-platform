package api

import (
	"net/http"

	"github.com/TaushifReza/email-platform/backend/internal/platform/config"
)

const csrfHeaderName = "X-CSRF-Token"

func cookieName(cfg config.Config) string {
	if cfg.CookieSecure {
		return "__Host-session"
	}
	return "session"
}

func setSessionCookie(w http.ResponseWriter, cfg config.Config, token string, maxAge int) {
	http.SetCookie(w, &http.Cookie{ // #nosec G124 -- attributes set below from cfg
		Name:     cookieName(cfg),
		Value:    token,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}

func clearSessionCookie(w http.ResponseWriter, cfg config.Config) {
	http.SetCookie(w, &http.Cookie{ // #nosec G124 -- attributes set below from cfg
		Name:     cookieName(cfg),
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}
