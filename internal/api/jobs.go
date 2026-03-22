package api

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

// JobSnapshot is the API-facing view of a background operation.
type JobSnapshot struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Status      string                 `json:"status"`
	Lines       []string               `json:"lines"`
	Result      map[string]any         `json:"result,omitempty"`
	Error       string                 `json:"error,omitempty"`
	StartedAt   time.Time              `json:"started_at"`
	CompletedAt time.Time              `json:"completed_at,omitempty"`
}

type JobManager struct {
	mu   sync.RWMutex
	jobs map[string]*jobState
}

type jobState struct {
	mu sync.RWMutex
	JobSnapshot
}

func NewJobManager() *JobManager {
	return &JobManager{jobs: map[string]*jobState{}}
}

func (m *JobManager) Start(name string, fn func(*JobRecorder) (map[string]any, error)) JobSnapshot {
	state := &jobState{
		JobSnapshot: JobSnapshot{
			ID:        randomID(),
			Name:      name,
			Status:    "running",
			Lines:     []string{},
			StartedAt: time.Now().UTC(),
		},
	}

	m.mu.Lock()
	m.jobs[state.ID] = state
	m.mu.Unlock()

	go func() {
		recorder := &JobRecorder{state: state}
		result, err := fn(recorder)
		state.mu.Lock()
		defer state.mu.Unlock()
		state.CompletedAt = time.Now().UTC()
		if err != nil {
			state.Status = "failed"
			state.Error = err.Error()
		} else {
			state.Status = "completed"
			state.Result = result
		}
	}()

	return state.snapshot()
}

func (m *JobManager) Snapshot(id string) (JobSnapshot, bool) {
	m.mu.RLock()
	state, ok := m.jobs[id]
	m.mu.RUnlock()
	if !ok {
		return JobSnapshot{}, false
	}
	return state.snapshot(), true
}

type JobRecorder struct {
	state *jobState
}

func (r *JobRecorder) Line(text string) {
	r.state.mu.Lock()
	defer r.state.mu.Unlock()
	r.state.Lines = append(r.state.Lines, text)
	if len(r.state.Lines) > 500 {
		r.state.Lines = append([]string(nil), r.state.Lines[len(r.state.Lines)-500:]...)
	}
}

func (s *jobState) snapshot() JobSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	lines := append([]string(nil), s.Lines...)
	result := map[string]any(nil)
	if s.Result != nil {
		result = make(map[string]any, len(s.Result))
		for key, value := range s.Result {
			result[key] = value
		}
	}
	return JobSnapshot{
		ID:          s.ID,
		Name:        s.Name,
		Status:      s.Status,
		Lines:       lines,
		Result:      result,
		Error:       s.Error,
		StartedAt:   s.StartedAt,
		CompletedAt: s.CompletedAt,
	}
}

func randomID() string {
	buf := make([]byte, 12)
	if _, err := rand.Read(buf); err != nil {
		return hex.EncodeToString([]byte(time.Now().UTC().Format("150405.000000")))
	}
	return hex.EncodeToString(buf)
}
