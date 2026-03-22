package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Gradient-Linux/concave/internal/auth"
)

func newTestApp(t *testing.T) *App {
	t.Helper()
	resetAPIDeps()
	t.Cleanup(resetAPIDeps)
	return New(Config{
		Addr:          "127.0.0.1:0",
		Version:       "test",
		WorkspaceRoot: t.TempDir(),
		Tokens: auth.TokenConfig{
			SigningKey: []byte("01234567890123456789012345678901"),
			TokenTTL:   24 * time.Hour,
		},
	})
}

func authRequest(t *testing.T, app *App, method, path string, role auth.Role, body []byte) *httptest.ResponseRecorder {
	t.Helper()
	token, err := auth.IssueToken(app.tokens, "alice", role)
	if err != nil {
		t.Fatalf("IssueToken() error = %v", err)
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: token})
	rr := httptest.NewRecorder()
	app.Handler().ServeHTTP(rr, req)
	return rr
}

func TestLoginSuccessSetsCookie(t *testing.T) {
	app := newTestApp(t)
	authenticateUser = func(username, password string) (auth.Role, error) {
		if username != "alice" || password != "secret" {
			t.Fatalf("Authenticate() called with %q / %q", username, password)
		}
		return auth.RoleAdmin, nil
	}

	body := []byte(`{"username":"alice","password":"secret"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	app.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	cookies := rr.Result().Cookies()
	if len(cookies) == 0 || cookies[0].Name != auth.SessionCookieName {
		t.Fatalf("cookies = %#v, want session cookie", cookies)
	}
	var payload map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if payload["role"] != "admin" {
		t.Fatalf("role = %v, want admin", payload["role"])
	}
}

func TestLoginWrongCredentialsAlwaysUnauthorized(t *testing.T) {
	app := newTestApp(t)
	authenticateUser = func(username, password string) (auth.Role, error) {
		return 0, auth.ErrInvalidCredentials
	}

	body := []byte(`{"username":"alice","password":"wrong"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	app.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rr.Code)
	}
	if got := rr.Body.String(); !bytes.Contains([]byte(got), []byte("invalid credentials")) {
		t.Fatalf("body = %q, want invalid credentials", got)
	}
}

func TestAuthMeRequiresCookie(t *testing.T) {
	app := newTestApp(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	rr := httptest.NewRecorder()
	app.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rr.Code)
	}
}

func TestRoleMiddlewareRejectsViewerForAdminRoute(t *testing.T) {
	app := newTestApp(t)
	rr := authRequest(t, app, http.MethodGet, "/api/v1/system/users", auth.RoleViewer, nil)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rr.Code)
	}
}

func TestSystemRebootRequiresConfirm(t *testing.T) {
	app := newTestApp(t)
	called := false
	runPrivileged = func(_ context.Context, description string, name string, args ...string) error {
		called = true
		return nil
	}

	rr := authRequest(t, app, http.MethodPost, "/api/v1/system/reboot", auth.RoleAdmin, []byte(`{"confirm":false}`))
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rr.Code)
	}
	if called {
		t.Fatal("runPrivileged() was called without confirmation")
	}
}

func TestSystemRebootAdminCallsPrivileged(t *testing.T) {
	app := newTestApp(t)
	called := false
	runPrivileged = func(_ context.Context, description string, name string, args ...string) error {
		called = true
		if name != "systemctl" || len(args) != 1 || args[0] != "reboot" {
			t.Fatalf("runPrivileged(%q, %q, %#v)", description, name, args)
		}
		return nil
	}

	rr := authRequest(t, app, http.MethodPost, "/api/v1/system/reboot", auth.RoleAdmin, []byte(`{"confirm":true}`))
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	if !called {
		t.Fatal("runPrivileged() was not called")
	}
}
