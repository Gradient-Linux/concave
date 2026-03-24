package api

import (
	"errors"
	"net/http"

	"github.com/Gradient-Linux/concave/internal/meshclient"
	"github.com/Gradient-Linux/concave/internal/resolverclient"
)

func (a *App) handleEnvStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	status, err := resolverclient.QueryStatus("")
	if errors.Is(err, resolverclient.ErrUnavailable) {
		writeJSON(w, http.StatusOK, map[string]any{"available": false, "message": "resolver not configured"})
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, status)
}

func (a *App) handleEnvDiff(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	reports, err := resolverclient.QueryDrift("", r.URL.Query().Get("group"))
	if errors.Is(err, resolverclient.ErrUnavailable) {
		writeJSON(w, http.StatusOK, map[string]any{"available": false, "message": "resolver not configured", "reports": []any{}})
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"reports": reports})
}

func (a *App) handleNodeStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	node, err := meshclient.QuerySelf("")
	if errors.Is(err, meshclient.ErrUnavailable) {
		writeJSON(w, http.StatusOK, map[string]any{"available": false, "message": "mesh not configured"})
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, node)
}

func (a *App) handleFleetStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	peers, err := meshclient.QueryFleet("")
	if errors.Is(err, meshclient.ErrUnavailable) {
		writeJSON(w, http.StatusOK, map[string]any{"available": false, "message": "mesh not configured", "peers": []any{}})
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"count": len(peers), "peers": peers})
}

func (a *App) handleFleetPeers(w http.ResponseWriter, r *http.Request) {
	a.handleFleetStatus(w, r)
}

func (a *App) handleTeams(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"available": false,
		"message":   "team management is not yet implemented — available in Gradient Linux Maxima after concave-resolver and compute engine are configured",
		"teams":     []any{},
	})
}
