package api

import (
	"encoding/json"
	"net/http"

	"github.com/Gradient-Linux/concave/internal/auth"
)

func (a *App) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	role, err := authenticateUser(req.Username, req.Password)
	if err != nil {
		if auth.IsRateLimited(err) {
			w.Header().Set("Retry-After", "60")
			writeError(w, http.StatusTooManyRequests, "too many attempts. try again in 60s")
			return
		}
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	claims, token, err := a.issueSession(w, req.Username, role)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]any{
		"username":   claims.Subject,
		"role":       claims.Role,
		"expires_at": claims.ExpiresAt.Time,
	}
	if clientWantsToken(r) {
		response["token"] = token
	}
	writeJSON(w, http.StatusOK, response)
}

func (a *App) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     auth.SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
	writeJSON(w, http.StatusOK, map[string]string{"status": "logged out"})
}

func (a *App) handleRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	token := authTokenFromRequest(r)
	if token == "" {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	refreshed, err := refreshSession(a.tokens, token)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "session cannot be refreshed")
		return
	}
	claims, err := validateSession(a.tokens, refreshed)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "session cannot be refreshed")
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     auth.SessionCookieName,
		Value:    refreshed,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  claims.ExpiresAt.Time,
	})
	response := map[string]any{
		"username":   claims.Subject,
		"role":       claims.Role,
		"expires_at": claims.ExpiresAt.Time,
	}
	if clientWantsToken(r) {
		response["token"] = refreshed
	}
	writeJSON(w, http.StatusOK, response)
}

func (a *App) handleMe(w http.ResponseWriter, r *http.Request) {
	claims, ok := ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"username":   claims.Subject,
		"role":       claims.Role,
		"expires_at": claims.ExpiresAt.Time,
	})
}
