package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/Gradient-Linux/concave/internal/auth"
	"github.com/Gradient-Linux/concave/internal/lab"
)

// labEnvRequest is the POST /api/v1/lab/envs body.
type labEnvRequest struct {
	Image       string `json:"image"`
	DisplayName string `json:"display_name"`
	Driver      string `json:"driver"`
	GPUs        int    `json:"gpus"`
	CPURequest  string `json:"cpu_request"`
	MemRequest  string `json:"mem_request"`
	TTLSeconds  int    `json:"ttl_seconds"`
}

type labExtendRequest struct {
	ExtendSeconds int `json:"extend_seconds"`
}

type labStorageRequest struct {
	HotTier  string `json:"hot_tier"`
	ColdTier string `json:"cold_tier"`
}

func (a *App) handleLabEnvs(w http.ResponseWriter, r *http.Request) {
	if a.lab == nil {
		writeError(w, http.StatusServiceUnavailable, "lab manager not initialised")
		return
	}
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, map[string]any{
			"envs":    a.lab.List(),
			"storage": a.lab.Storage(),
			"drivers": a.lab.Registry().Names(),
			"active":  a.lab.Registry().Active(),
		})
	case http.MethodPost:
		if err := auth.Require(ClaimsFromContextMust(r).Role, auth.ActionStart); err != nil {
			writeError(w, http.StatusForbidden, "insufficient role")
			return
		}
		var req labEnvRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if req.TTLSeconds <= 0 {
			writeError(w, http.StatusBadRequest, "ttl_seconds must be positive")
			return
		}
		claims := ClaimsFromContextMust(r)
		spec := lab.EnvSpec{
			Owner:       claims.Subject,
			Image:       req.Image,
			DisplayName: req.DisplayName,
			Driver:      req.Driver,
			GPUs:        req.GPUs,
			CPURequest:  req.CPURequest,
			MemRequest:  req.MemRequest,
			TTL:         time.Duration(req.TTLSeconds) * time.Second,
		}
		env, err := a.lab.Launch(r.Context(), spec)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, env)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (a *App) handleLabEnvSubroutes(w http.ResponseWriter, r *http.Request) {
	if a.lab == nil {
		writeError(w, http.StatusServiceUnavailable, "lab manager not initialised")
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/lab/envs/")
	if path == "" {
		writeError(w, http.StatusNotFound, "env id required")
		return
	}
	parts := strings.Split(path, "/")
	id := parts[0]

	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			env, err := a.lab.Get(r.Context(), id)
			if err != nil {
				writeError(w, http.StatusNotFound, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, env)
		case http.MethodDelete:
			if err := auth.Require(ClaimsFromContextMust(r).Role, auth.ActionStop); err != nil {
				writeError(w, http.StatusForbidden, "insufficient role")
				return
			}
			env, err := a.lab.ArchiveAndDestroy(r.Context(), id)
			if err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, env)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	action := parts[1]
	switch action {
	case "extend":
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		if err := auth.Require(ClaimsFromContextMust(r).Role, auth.ActionStart); err != nil {
			writeError(w, http.StatusForbidden, "insufficient role")
			return
		}
		var req labExtendRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if req.ExtendSeconds <= 0 {
			writeError(w, http.StatusBadRequest, "extend_seconds must be positive")
			return
		}
		env, err := a.lab.ExtendTTL(r.Context(), id, time.Duration(req.ExtendSeconds)*time.Second)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, env)
	case "archive":
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		if err := auth.Require(ClaimsFromContextMust(r).Role, auth.ActionStop); err != nil {
			writeError(w, http.StatusForbidden, "insufficient role")
			return
		}
		env, err := a.lab.ArchiveAndDestroy(r.Context(), id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, env)
	default:
		writeError(w, http.StatusNotFound, "unknown env action")
	}
}

func (a *App) handleLabStorage(w http.ResponseWriter, r *http.Request) {
	if a.lab == nil {
		writeError(w, http.StatusServiceUnavailable, "lab manager not initialised")
		return
	}
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, a.lab.Storage())
	case http.MethodPut:
		if ClaimsFromContextMust(r).Role < auth.RoleAdmin {
			writeError(w, http.StatusForbidden, "admin only")
			return
		}
		var req labStorageRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		next := lab.StorageConfig{HotTier: req.HotTier, ColdTier: req.ColdTier}
		if err := a.lab.UpdateStorage(next); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, a.lab.Storage())
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (a *App) handleLabDrivers(w http.ResponseWriter, r *http.Request) {
	if a.lab == nil {
		writeError(w, http.StatusServiceUnavailable, "lab manager not initialised")
		return
	}
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, map[string]any{
			"drivers": a.lab.Registry().Names(),
			"active":  a.lab.Registry().Active(),
		})
	case http.MethodPut:
		if ClaimsFromContextMust(r).Role < auth.RoleAdmin {
			writeError(w, http.StatusForbidden, "admin only")
			return
		}
		var req struct {
			Driver string `json:"driver"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if err := a.lab.Registry().SetActive(req.Driver); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"active": a.lab.Registry().Active()})
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}
