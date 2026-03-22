package api

import (
	"context"
	"encoding/json"
	"net/http"
	"os/exec"
	"os/user"
	"strings"

	"github.com/Gradient-Linux/concave/internal/docker"
	"github.com/Gradient-Linux/concave/internal/system"
	"github.com/Gradient-Linux/concave/internal/workspace"
	"github.com/creack/pty"
	"github.com/gorilla/websocket"
)

func (a *App) handleWorkspaceBackup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSON(w, http.StatusAccepted, a.jobAccepted("workspace-backup", func(rec *JobRecorder) (map[string]any, error) {
		unlock, err := system.Lock("api-workspace-backup")
		if err != nil {
			return nil, err
		}
		defer unlock()
		path, err := workspace.Backup()
		if err != nil {
			return nil, err
		}
		rec.Line("backup created at " + path)
		return map[string]any{"path": path}, nil
	}))
}

func (a *App) handleWorkspaceClean(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSON(w, http.StatusAccepted, a.jobAccepted("workspace-clean", func(rec *JobRecorder) (map[string]any, error) {
		unlock, err := system.Lock("api-workspace-clean")
		if err != nil {
			return nil, err
		}
		defer unlock()
		if err := workspace.CleanOutputs(); err != nil {
			return nil, err
		}
		rec.Line("outputs cleaned")
		return nil, nil
	}))
}

func (a *App) handleSystemReboot(w http.ResponseWriter, r *http.Request) {
	a.handleSystemCommand(w, r, "reboot", "systemctl", "reboot")
}

func (a *App) handleSystemShutdown(w http.ResponseWriter, r *http.Request) {
	a.handleSystemCommand(w, r, "shutdown", "systemctl", "poweroff")
}

func (a *App) handleSystemRestartDocker(w http.ResponseWriter, r *http.Request) {
	a.handleSystemCommand(w, r, "restart-docker", "systemctl", "restart", "docker")
}

func (a *App) handleSystemCommand(w http.ResponseWriter, r *http.Request, name string, command string, args ...string) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req struct {
		Confirm bool `json:"confirm"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || !req.Confirm {
		writeError(w, http.StatusBadRequest, "confirm: true required")
		return
	}
	claims := ClaimsFromContextMust(r)
	system.Logger.Info("system control requested", "action", name, "user", claims.Subject)
	if err := runPrivileged(r.Context(), "Running "+name, command, args...); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": name + " initiated"})
}

var terminalUpgrader = websocket.Upgrader{
	CheckOrigin: func(*http.Request) bool { return true },
}

func (a *App) handleContainerTerminal(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/terminal/container/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		writeError(w, http.StatusBadRequest, "suite and container required")
		return
	}
	suiteName := parts[0]
	containerName := parts[1]
	s, err := suiteFromCurrentStateOrBase(suiteName)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	allowed := false
	for _, container := range s.Containers {
		if container.Name == containerName {
			allowed = true
			break
		}
	}
	if !allowed {
		writeError(w, http.StatusBadRequest, "container not valid for suite")
		return
	}
	status, err := docker.ContainerStatus(r.Context(), containerName)
	if err != nil || status != "running" {
		writeError(w, http.StatusConflict, "container not running")
		return
	}
	a.handlePTY(w, r, exec.Command("docker", "exec", "-it", containerName, "bash"))
}

func (a *App) handleHostTerminal(w http.ResponseWriter, r *http.Request) {
	claims := ClaimsFromContextMust(r)
	current, err := user.Current()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	var cmd *exec.Cmd
	if current.Username == claims.Subject {
		cmd = exec.Command("/bin/bash")
	} else {
		if _, err := user.Lookup(claims.Subject); err != nil {
			writeError(w, http.StatusBadRequest, "unknown user")
			return
		}
		cmd = exec.Command("sudo", "/usr/local/libexec/concave-host-shell", claims.Subject)
	}
	a.handlePTY(w, r, cmd)
}

func (a *App) handlePTY(w http.ResponseWriter, r *http.Request, cmd *exec.Cmd) {
	conn, err := terminalUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	ptmx, err := pty.Start(cmd)
	if err != nil {
		_ = conn.WriteJSON(map[string]string{"type": "error", "data": "failed to start shell"})
		return
	}
	defer func() {
		_ = ptmx.Close()
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
	}()

	go func() {
		for {
			var msg struct {
				Type string `json:"type"`
				Data string `json:"data"`
				Rows uint16 `json:"rows"`
				Cols uint16 `json:"cols"`
			}
			if err := conn.ReadJSON(&msg); err != nil {
				return
			}
			switch msg.Type {
			case "data":
				_, _ = ptmx.Write([]byte(msg.Data))
			case "resize":
				_ = pty.Setsize(ptmx, &pty.Winsize{Rows: msg.Rows, Cols: msg.Cols})
			}
		}
	}()

	buf := make([]byte, 4096)
	for {
		n, err := ptmx.Read(buf)
		if n > 0 {
			_ = conn.WriteJSON(map[string]string{"type": "data", "data": string(buf[:n])})
		}
		if err != nil {
			return
		}
	}
}

func (a *App) handleSuiteLogs(w http.ResponseWriter, r *http.Request, name string) {
	container := r.URL.Query().Get("container")
	if container == "" {
		s, err := suiteFromCurrentStateOrBase(name)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if len(s.Containers) == 0 {
			writeError(w, http.StatusBadRequest, "suite has no containers")
			return
		}
		container = s.Containers[0].Name
	}
	conn, err := terminalUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	go func() {
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				cancel()
				return
			}
		}
	}()
	_ = streamDockerLogs(ctx, container, func(line string) {
		_ = conn.WriteJSON(map[string]string{"type": "line", "line": line})
	})
}
