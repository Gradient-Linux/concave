package api

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Gradient-Linux/concave/internal/gpu"
	"github.com/Gradient-Linux/concave/internal/suite"
)

func (a *App) handleMetricsStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "streaming unsupported")
		return
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		payload := metricsPayload{
			Workspace: workspaceSnapshotOrNil(),
			Suites:    suiteSummaries(),
			GPU:       gpuMetrics(),
			CPU:       cpuMetrics(),
			Memory:    memoryMetrics(),
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}
		data, _ := json.Marshal(payload)
		_, _ = fmt.Fprintf(w, "event: metrics\ndata: %s\n\n", data)
		flusher.Flush()

		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
		}
	}
}

func (a *App) handleWorkspace(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	payload, err := workspaceSnapshot()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, payload)
}

func (a *App) handleDoctor(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"checks": runDoctorChecks()})
}

func (a *App) handleSystemInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	hostname, _ := os.Hostname()
	writeJSON(w, http.StatusOK, map[string]any{
		"hostname": hostname,
		"uptime":   hostUptime(),
		"kernel":   kernelVersion(),
		"os":       osReleaseName(),
		"concave":  a.version,
		"docker":   dockerServerVersion(),
		"services": []map[string]any{
			{"name": "concave-serve", "status": serviceStatus("concave-serve"), "user": "gradient-svc"},
			{"name": "docker", "status": serviceStatus("docker")},
		},
	})
}

func (a *App) handleSystemUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	users, err := gradientUsers()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"users": users})
}

func (a *App) handleUsersActivity(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	users, err := activeUsers()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"users": users})
}

func (a *App) handleSuites(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"suites": suiteSummaries()})
}

func suiteSummaries() []suiteSummary {
	names := suite.Names()
	summaries := make([]suiteSummary, 0, len(names))
	for _, name := range names {
		summaries = append(summaries, suiteSnapshot(name))
	}
	return summaries
}

func workspaceSnapshotOrNil() any {
	payload, err := workspaceSnapshot()
	if err != nil {
		return map[string]string{"error": err.Error()}
	}
	return payload
}

func gpuMetrics() gpuMetricsPayload {
	state, err := gpu.Detect()
	if err != nil {
		return gpuMetricsPayload{Error: err.Error()}
	}
	switch state {
	case gpu.GPUStateNone:
		return gpuMetricsPayload{Error: "no GPU detected"}
	case gpu.GPUStateAMD:
		return gpuMetricsPayload{Error: "GPU metrics unavailable for AMD hosts"}
	}

	devices, err := gpuMetricsDevices()
	if err != nil {
		return gpuMetricsPayload{Error: err.Error()}
	}
	return gpuMetricsPayload{Devices: devices}
}

func gpuMetricsDevices() ([]gpuDeviceMetric, error) {
	devices, err := gpu.NVIDIADevices()
	if err != nil {
		return nil, err
	}
	if len(devices) == 0 {
		return nil, fmt.Errorf("no GPU detected")
	}
	result := make([]gpuDeviceMetric, 0, len(devices))
	for _, device := range devices {
		result = append(result, gpuDeviceMetric{
			Name:        device.Name,
			Utilization: float64(device.Utilization),
			MemoryUsed:  int64(device.MemoryUsedMiB),
			MemoryTotal: int64(device.MemoryTotalMiB),
		})
	}
	return result, nil
}

func kernelVersion() string {
	out, err := exec.Command("uname", "-r").CombinedOutput()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

func osReleaseName() string {
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return "unknown"
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "PRETTY_NAME=") {
			return strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), "\"")
		}
	}
	return "unknown"
}

func hostUptime() string {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return "unknown"
	}
	fields := strings.Fields(string(data))
	if len(fields) == 0 {
		return "unknown"
	}
	seconds, _ := time.ParseDuration(fields[0] + "s")
	seconds = seconds.Round(time.Minute)
	days := int(seconds.Hours()) / 24
	hours := int(seconds.Hours()) % 24
	minutes := int(seconds.Minutes()) % 60
	if days > 0 {
		return fmt.Sprintf("%d days %d hours %d minutes", days, hours, minutes)
	}
	return fmt.Sprintf("%d hours %d minutes", hours, minutes)
}

func serviceStatus(name string) string {
	out, err := exec.Command("systemctl", "is-active", name).CombinedOutput()
	if err != nil {
		return strings.TrimSpace(string(out))
	}
	return strings.TrimSpace(string(out))
}

func gradientUsers() ([]map[string]any, error) {
	passwd, err := os.Open("/etc/passwd")
	if err != nil {
		return nil, err
	}
	defer passwd.Close()
	users := []map[string]any{}
	scanner := bufio.NewScanner(passwd)
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), ":")
		if len(fields) < 7 {
			continue
		}
		username := fields[0]
		role, err := resolveRole(username)
		if err != nil {
			continue
		}
		users = append(users, map[string]any{
			"username": username,
			"role":     role,
		})
	}
	return users, scanner.Err()
}

func activeUsers() ([]map[string]any, error) {
	containers, _ := labeledContainerStats()
	users, _ := gradientUsers()
	index := make(map[string]map[string]any, len(users))
	for _, entry := range users {
		username := entry["username"].(string)
		index[username] = map[string]any{
			"username":       username,
			"role":           entry["role"],
			"containers":     []map[string]any{},
			"gpu_memory_mib": 0,
			"last_active":    time.Time{},
		}
	}
	for _, container := range containers {
		username := container["user"].(string)
		if _, ok := index[username]; !ok {
			role, _ := resolveRole(username)
			index[username] = map[string]any{
				"username":       username,
				"role":           role,
				"containers":     []map[string]any{},
				"gpu_memory_mib": 0,
				"last_active":    time.Time{},
			}
		}
		entry := index[username]
		entry["containers"] = append(entry["containers"].([]map[string]any), container)
		if memoryMiB, ok := container["gpu_memory_mib"].(int); ok && memoryMiB > 0 {
			entry["gpu_memory_mib"] = entry["gpu_memory_mib"].(int) + memoryMiB
		}
		if lastActive, ok := container["last_active"].(time.Time); ok {
			current := entry["last_active"].(time.Time)
			if current.IsZero() || lastActive.After(current) {
				entry["last_active"] = lastActive
			}
		}
	}
	list := make([]map[string]any, 0, len(index))
	usernames := make([]string, 0, len(index))
	for username := range index {
		usernames = append(usernames, username)
	}
	sort.Strings(usernames)
	for _, username := range usernames {
		entry := index[username]
		if entry["last_active"].(time.Time).IsZero() {
			delete(entry, "last_active")
		}
		list = append(list, entry)
	}
	return list, nil
}

func labeledContainerStats() ([]map[string]any, error) {
	statsByName, _ := dockerStatsByName()
	out, err := exec.Command("docker", "ps", "-a", "--format", "{{.Names}}").CombinedOutput()
	if err != nil {
		return nil, err
	}
	names := strings.Fields(strings.TrimSpace(string(out)))
	stats := make([]map[string]any, 0, len(names))
	for _, name := range names {
		inspectOut, inspectErr := exec.Command("docker", "inspect", name).CombinedOutput()
		if inspectErr != nil {
			continue
		}
		var payload []struct {
			Config struct {
				Labels map[string]string `json:"Labels"`
			} `json:"Config"`
			State struct {
				Status    string `json:"Status"`
				StartedAt string `json:"StartedAt"`
			} `json:"State"`
		}
		if json.Unmarshal(inspectOut, &payload) != nil || len(payload) == 0 {
			continue
		}
		labels := payload[0].Config.Labels
		username := labels["gradient.user"]
		if username == "" {
			username = currentUser()
		}
		lastActive, _ := time.Parse(time.RFC3339Nano, payload[0].State.StartedAt)
		runtimeStat := statsByName[name]
		stats = append(stats, map[string]any{
			"name":           name,
			"suite":          labels["gradient.suite"],
			"status":         payload[0].State.Status,
			"user":           username,
			"cpu_percent":    runtimeStat.CPUPercent,
			"memory_mib":     runtimeStat.MemoryMiB,
			"gpu_memory_mib": 0,
			"last_active":    lastActive.UTC(),
		})
	}
	return stats, nil
}

type dockerRuntimeStat struct {
	CPUPercent float64
	MemoryMiB  int
}

func dockerStatsByName() (map[string]dockerRuntimeStat, error) {
	out, err := exec.Command("docker", "stats", "--no-stream", "--format", "{{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}").CombinedOutput()
	if err != nil {
		return nil, err
	}
	stats := map[string]dockerRuntimeStat{}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) != 3 {
			continue
		}
		stats[parts[0]] = dockerRuntimeStat{
			CPUPercent: parseDockerPercent(parts[1]),
			MemoryMiB:  parseDockerMiB(parts[2]),
		}
	}
	return stats, nil
}

func parseDockerPercent(raw string) float64 {
	value := strings.TrimSpace(strings.TrimSuffix(raw, "%"))
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}
	return parsed
}

func parseDockerMiB(raw string) int {
	used := strings.TrimSpace(strings.SplitN(raw, "/", 2)[0])
	if used == "" || used == "--" {
		return 0
	}
	index := 0
	for index < len(used) && ((used[index] >= '0' && used[index] <= '9') || used[index] == '.') {
		index++
	}
	if index == 0 {
		return 0
	}
	value, err := strconv.ParseFloat(used[:index], 64)
	if err != nil {
		return 0
	}
	unit := strings.TrimSpace(strings.ToLower(used[index:]))
	switch unit {
	case "gib", "gb":
		return int(value * 1024)
	case "mib", "mb":
		return int(value)
	case "kib", "kb":
		return int(value / 1024)
	case "b":
		return int(value / (1024 * 1024))
	default:
		return int(value)
	}
}
