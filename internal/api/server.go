package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Gradient-Linux/concave/internal/auth"
)

// Config holds the API server runtime configuration.
type Config struct {
	Addr          string
	Version       string
	WorkspaceRoot string
	Tokens        auth.TokenConfig
}

// App owns the API server routes and in-memory managers.
type App struct {
	addr          string
	version       string
	workspaceRoot string
	tokens        auth.TokenConfig
	jobs          *JobManager
	mux           *http.ServeMux
}

func New(cfg Config) *App {
	if cfg.Addr == "" {
		cfg.Addr = "127.0.0.1:7777"
	}
	app := &App{
		addr:          cfg.Addr,
		version:       cfg.Version,
		workspaceRoot: cfg.WorkspaceRoot,
		tokens:        cfg.Tokens,
		jobs:          NewJobManager(),
		mux:           http.NewServeMux(),
	}
	app.routes()
	return app
}

func (a *App) Handler() http.Handler {
	return a.mux
}

func (a *App) ListenAndServe(ctx context.Context) error {
	server := &http.Server{
		Addr:              a.addr,
		Handler:           a.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (a *App) routes() {
	a.mux.HandleFunc("/api/v1/health", a.handleHealth)
	a.mux.HandleFunc("/api/v1/auth/login", a.handleLogin)
	a.mux.Handle("/api/v1/auth/logout", a.authMiddleware(http.HandlerFunc(a.handleLogout)))
	a.mux.Handle("/api/v1/auth/refresh", a.authMiddleware(http.HandlerFunc(a.handleRefresh)))
	a.mux.Handle("/api/v1/auth/me", a.authMiddleware(http.HandlerFunc(a.handleMe)))

	a.mux.Handle("/api/v1/jobs/", a.authMiddleware(RoleMiddleware(auth.RoleViewer, http.HandlerFunc(a.handleJob))))
	a.mux.Handle("/api/v1/metrics/stream", a.authMiddleware(RoleMiddleware(auth.RoleViewer, http.HandlerFunc(a.handleMetricsStream))))
	a.mux.Handle("/api/v1/check", a.authMiddleware(RoleMiddleware(auth.RoleViewer, http.HandlerFunc(a.handleDoctor))))
	a.mux.Handle("/api/v1/doctor", a.authMiddleware(RoleMiddleware(auth.RoleViewer, http.HandlerFunc(a.handleDoctor))))
	a.mux.Handle("/api/v1/env/status", a.authMiddleware(RoleMiddleware(auth.RoleViewer, http.HandlerFunc(a.handleEnvStatus))))
	a.mux.Handle("/api/v1/env/diff", a.authMiddleware(RoleMiddleware(auth.RoleViewer, http.HandlerFunc(a.handleEnvDiff))))
	a.mux.Handle("/api/v1/node/status", a.authMiddleware(RoleMiddleware(auth.RoleViewer, http.HandlerFunc(a.handleNodeStatus))))
	a.mux.Handle("/api/v1/fleet/status", a.authMiddleware(RoleMiddleware(auth.RoleViewer, http.HandlerFunc(a.handleFleetStatus))))
	a.mux.Handle("/api/v1/fleet/peers", a.authMiddleware(RoleMiddleware(auth.RoleViewer, http.HandlerFunc(a.handleFleetPeers))))
	a.mux.Handle("/api/v1/teams", a.authMiddleware(RoleMiddleware(auth.RoleViewer, http.HandlerFunc(a.handleTeams))))
	a.mux.Handle("/api/v1/workspace", a.authMiddleware(RoleMiddleware(auth.RoleViewer, http.HandlerFunc(a.handleWorkspace))))
	a.mux.Handle("/api/v1/workspace/backup", a.authMiddleware(RoleMiddleware(auth.RoleOperator, http.HandlerFunc(a.handleWorkspaceBackup))))
	a.mux.Handle("/api/v1/workspace/prune", a.authMiddleware(RoleMiddleware(auth.RoleOperator, http.HandlerFunc(a.handleWorkspaceClean))))
	a.mux.Handle("/api/v1/workspace/clean", a.authMiddleware(RoleMiddleware(auth.RoleOperator, http.HandlerFunc(a.handleWorkspaceClean))))
	a.mux.Handle("/api/v1/system/info", a.authMiddleware(RoleMiddleware(auth.RoleViewer, http.HandlerFunc(a.handleSystemInfo))))
	a.mux.Handle("/api/v1/system/users", a.authMiddleware(RoleMiddleware(auth.RoleAdmin, http.HandlerFunc(a.handleSystemUsers))))
	a.mux.Handle("/api/v1/system/reboot", a.authMiddleware(RoleMiddleware(auth.RoleAdmin, http.HandlerFunc(a.handleSystemReboot))))
	a.mux.Handle("/api/v1/system/shutdown", a.authMiddleware(RoleMiddleware(auth.RoleAdmin, http.HandlerFunc(a.handleSystemShutdown))))
	a.mux.Handle("/api/v1/system/restart-docker", a.authMiddleware(RoleMiddleware(auth.RoleAdmin, http.HandlerFunc(a.handleSystemRestartDocker))))
	a.mux.Handle("/api/v1/users/activity", a.authMiddleware(RoleMiddleware(auth.RoleAdmin, http.HandlerFunc(a.handleUsersActivity))))
	a.mux.Handle("/api/v1/suites", a.authMiddleware(RoleMiddleware(auth.RoleViewer, http.HandlerFunc(a.handleSuites))))
	a.mux.Handle("/api/v1/suites/", a.authMiddleware(RoleMiddleware(auth.RoleViewer, http.HandlerFunc(a.handleSuiteSubroutes))))
	a.mux.Handle("/api/v1/terminal/container/", a.authMiddleware(RoleMiddleware(auth.RoleDeveloper, http.HandlerFunc(a.handleContainerTerminal))))
	a.mux.Handle("/api/v1/terminal/host", a.authMiddleware(RoleMiddleware(auth.RoleAdmin, http.HandlerFunc(a.handleHostTerminal))))
}

func (a *App) issueSession(w http.ResponseWriter, username string, role auth.Role) (auth.Claims, string, error) {
	token, err := auth.IssueToken(a.tokens, username, role)
	if err != nil {
		return auth.Claims{}, "", err
	}
	claims, err := auth.ValidateToken(a.tokens, token)
	if err != nil {
		return auth.Claims{}, "", err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     auth.SessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  claims.ExpiresAt.Time,
	})
	return claims, token, nil
}

func clientWantsToken(r *http.Request) bool {
	return r.Header.Get("X-Concave-Client") == "tui"
}

func (a *App) jobAccepted(name string, fn func(*JobRecorder) (map[string]any, error)) map[string]any {
	job := a.jobs.Start(name, fn)
	return map[string]any{"job_id": job.ID}
}

func (a *App) routeNotFound(w http.ResponseWriter, _ *http.Request) {
	writeError(w, http.StatusNotFound, fmt.Sprintf("route not found"))
}
