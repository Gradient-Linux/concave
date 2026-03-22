package api

import (
	"net/http"
	"strings"
	"time"
)

func (a *App) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":    "ok",
		"version":   a.version,
		"timestamp": time.Now().UTC(),
	})
}

func (a *App) handleJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	id := r.PathValue("id")
	if id == "" {
		id = r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:]
	}
	job, ok := a.jobs.Snapshot(id)
	if !ok {
		writeError(w, http.StatusNotFound, "job not found")
		return
	}
	writeJSON(w, http.StatusOK, job)
}
