package api

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/Gradient-Linux/concave/internal/config"
	"github.com/Gradient-Linux/concave/internal/docker"
	"github.com/Gradient-Linux/concave/internal/gpu"
	"github.com/Gradient-Linux/concave/internal/suite"
	"github.com/Gradient-Linux/concave/internal/system"
	"github.com/Gradient-Linux/concave/internal/ui"
	"github.com/Gradient-Linux/concave/internal/workspace"
)

type suiteSummary struct {
	Name          string              `json:"name"`
	Installed     bool                `json:"installed"`
	State         string              `json:"state"`
	Current       string              `json:"current,omitempty"`
	Previous      string              `json:"previous,omitempty"`
	Ports         []apiPortMapping    `json:"ports"`
	Containers    []containerInfo     `json:"containers"`
	GPURequired   bool                `json:"gpu_required"`
	Error         string              `json:"error,omitempty"`
	ComposeExists bool                `json:"compose_exists"`
}

type apiPortMapping struct {
	Host        int    `json:"host"`
	Container   int    `json:"container"`
	Description string `json:"description,omitempty"`
}

type containerInfo struct {
	Name     string `json:"name"`
	Image    string `json:"image"`
	Role     string `json:"role"`
	Status   string `json:"status"`
	Current  string `json:"current,omitempty"`
	Previous string `json:"previous,omitempty"`
}

type workspacePayload struct {
	Root   string            `json:"root"`
	Total  uint64            `json:"total"`
	Free   uint64            `json:"free"`
	Used   uint64            `json:"used"`
	Usages map[string]uint64 `json:"usages"`
}

type doctorCheck struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Detail   string `json:"detail"`
	Recovery string `json:"recovery,omitempty"`
}

func installedSuiteSet() (map[string]struct{}, config.State, error) {
	state, err := config.LoadState()
	if err != nil {
		return nil, config.State{}, err
	}
	set := make(map[string]struct{}, len(state.Installed))
	for _, name := range state.Installed {
		set[name] = struct{}{}
	}
	return set, state, nil
}

func suiteFromCurrentState(name string) (suite.Suite, error) {
	s, err := suite.Get(name)
	if err != nil {
		return suite.Suite{}, err
	}
	if name != "forge" {
		return s, nil
	}

	manifest, err := config.LoadManifest()
	if err != nil {
		return suite.Suite{}, err
	}
	entries, ok := manifest["forge"]
	if !ok || len(entries) == 0 {
		return suite.Suite{}, fmt.Errorf("forge has no recorded component selection")
	}
	names := make([]string, 0, len(entries))
	overrides := make(map[string]string, len(entries))
	for containerName, version := range entries {
		names = append(names, containerName)
		overrides[containerName] = version.Current
	}
	sort.Strings(names)
	selection, err := suite.SelectionFromContainerNames(names, overrides)
	if err != nil {
		return suite.Suite{}, err
	}
	s.Containers = selection.Containers
	s.Ports = selection.Ports
	s.Volumes = selection.Volumes
	return s, nil
}

func suiteSnapshot(name string) suiteSummary {
	base, err := suite.Get(name)
	if err != nil {
		return suiteSummary{Name: name, Error: err.Error(), State: "error"}
	}
	installed, _ := config.IsInstalled(name)
	composeExists := fileExists(docker.ComposePath(name))
	if !installed {
		return suiteSummary{
			Name:          name,
			Installed:     false,
			State:         "not-installed",
			Ports:         apiPorts(base.Ports),
			GPURequired:   base.GPURequired,
			ComposeExists: composeExists,
		}
	}

	s, err := suiteFromCurrentState(name)
	if err != nil {
		return suiteSummary{
			Name:          name,
			Installed:     true,
			State:         "unconfigured",
			Ports:         apiPorts(base.Ports),
			GPURequired:   base.GPURequired,
			ComposeExists: composeExists,
			Error:         err.Error(),
		}
	}

	manifest, _ := config.LoadManifest()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	containers := make([]containerInfo, 0, len(s.Containers))
	running := 0
	stopped := 0
	for _, container := range s.Containers {
		status, err := docker.ContainerStatus(ctx, container.Name)
		if err != nil {
			status = "error"
		}
		switch status {
		case "running":
			running++
		case "stopped", "not found":
			stopped++
		}
		version := config.ImageVersion{}
		if suiteVersions, ok := manifest[s.Name]; ok {
			version = suiteVersions[container.Name]
		}
		containers = append(containers, containerInfo{
			Name:     container.Name,
			Image:    container.Image,
			Role:     container.Role,
			Status:   status,
			Current:  version.Current,
			Previous: version.Previous,
		})
	}

	state := "stopped"
	switch {
	case running == len(containers) && len(containers) > 0:
		state = "running"
	case running > 0 && stopped > 0:
		state = "degraded"
	}

	current := ""
	previous := ""
	if len(containers) > 0 {
		current = containers[0].Current
		previous = containers[0].Previous
	}
	return suiteSummary{
		Name:          s.Name,
		Installed:     true,
		State:         state,
		Current:       current,
		Previous:      previous,
		Ports:         apiPorts(s.Ports),
		Containers:    containers,
		GPURequired:   s.GPURequired,
		ComposeExists: composeExists,
	}
}

func apiPorts(ports []suite.PortMapping) []apiPortMapping {
	mapped := make([]apiPortMapping, 0, len(ports))
	for _, port := range ports {
		mapped = append(mapped, apiPortMapping{
			Host:        port.Port,
			Container:   port.Port,
			Description: port.Service,
		})
	}
	return mapped
}

func workspaceSnapshot() (workspacePayload, error) {
	if err := workspace.EnsureLayout(); err != nil {
		return workspacePayload{}, err
	}
	var stat syscall.Statfs_t
	if err := syscall.Statfs(workspace.Root(), &stat); err != nil {
		return workspacePayload{}, err
	}
	usages, err := workspace.Status()
	if err != nil {
		return workspacePayload{}, err
	}
	result := workspacePayload{
		Root:   workspace.Root(),
		Total:  stat.Blocks * uint64(stat.Bsize),
		Free:   stat.Bavail * uint64(stat.Bsize),
		Used:   (stat.Blocks - stat.Bfree) * uint64(stat.Bsize),
		Usages: map[string]uint64{},
	}
	for _, usage := range usages {
		result.Usages[usage.Name] = uint64(usage.Bytes)
	}
	return result, nil
}

func runDoctorChecks() []doctorCheck {
	results := []doctorCheck{}
	if ok, err := system.DockerRunning(); err != nil {
		results = append(results, doctorCheck{Name: "Docker", Status: "fail", Detail: err.Error()})
	} else if ok {
		version := dockerServerVersion()
		results = append(results, doctorCheck{Name: "Docker", Status: "pass", Detail: "running (" + version + ")"})
	} else {
		results = append(results, doctorCheck{Name: "Docker", Status: "fail", Detail: "not running"})
	}
	if ok, err := system.UserInDockerGroup(); err != nil {
		results = append(results, doctorCheck{Name: "Docker group", Status: "fail", Detail: err.Error()})
	} else if ok {
		results = append(results, doctorCheck{Name: "Docker group", Status: "pass", Detail: "user in docker group"})
	} else {
		results = append(results, doctorCheck{Name: "Docker group", Status: "warn", Detail: "user not in docker group"})
	}
	if ok, err := system.InternetReachable(); err != nil {
		results = append(results, doctorCheck{Name: "Internet", Status: "warn", Detail: err.Error()})
	} else if ok {
		results = append(results, doctorCheck{Name: "Internet", Status: "pass", Detail: "reachable"})
	} else {
		results = append(results, doctorCheck{Name: "Internet", Status: "warn", Detail: "not reachable"})
	}

	switch state, err := gpu.Detect(); {
	case err != nil:
		results = append(results, doctorCheck{Name: "GPU", Status: "warn", Detail: err.Error()})
	case state == gpu.GPUStateNVIDIA:
		detail := "NVIDIA detected"
		if devices, devErr := gpu.NVIDIADevices(); devErr == nil && len(devices) > 0 {
			detail = devices[0].Name
		}
		results = append(results, doctorCheck{Name: "GPU", Status: "pass", Detail: detail})
	case state == gpu.GPUStateAMD:
		results = append(results, doctorCheck{Name: "GPU", Status: "warn", Detail: "AMD detected — ROCm support coming later"})
	default:
		results = append(results, doctorCheck{Name: "GPU", Status: "warn", Detail: "not detected — CPU-only mode"})
	}

	if payload, err := workspaceSnapshot(); err != nil {
		results = append(results, doctorCheck{Name: "Workspace", Status: "fail", Detail: err.Error()})
	} else {
		results = append(results, doctorCheck{Name: "Workspace", Status: "pass", Detail: payload.Root})
	}

	for _, name := range suite.Names() {
		summary := suiteSnapshot(name)
		switch {
		case !summary.Installed:
			results = append(results, doctorCheck{Name: name, Status: "skip", Detail: "not installed"})
		case summary.State == "degraded":
			recovery := "concave start " + name
			results = append(results, doctorCheck{Name: name, Status: "warn", Detail: fmt.Sprintf("%d / %d containers running", countRunning(summary.Containers), len(summary.Containers)), Recovery: recovery})
		case summary.State == "running":
			results = append(results, doctorCheck{Name: name, Status: "pass", Detail: fmt.Sprintf("%d / %d containers running", countRunning(summary.Containers), len(summary.Containers))})
		case summary.State == "unconfigured":
			results = append(results, doctorCheck{Name: name, Status: "warn", Detail: summary.Error, Recovery: "concave remove " + name})
		default:
			results = append(results, doctorCheck{Name: name, Status: "fail", Detail: "stopped", Recovery: "concave start " + name})
		}
	}
	return results
}

func countRunning(containers []containerInfo) int {
	count := 0
	for _, container := range containers {
		if container.Status == "running" {
			count++
		}
	}
	return count
}

func dockerServerVersion() string {
	out, err := exec.Command("docker", "version", "--format", "{{.Server.Version}}").CombinedOutput()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

func withCapturedUI(fn func() error, line func(string)) error {
	buffer := &bytes.Buffer{}
	ui.SetOutput(buffer)
	defer ui.ResetOutput()
	err := fn()
	for _, entry := range strings.Split(strings.TrimSpace(buffer.String()), "\n") {
		entry = strings.TrimSpace(entry)
		if entry != "" {
			line(entry)
		}
	}
	return err
}

func orderInstalled(installed []string, reverse bool) []string {
	set := make(map[string]struct{}, len(installed))
	for _, name := range installed {
		set[name] = struct{}{}
	}
	ordered := make([]string, 0, len(installed))
	for _, name := range []string{"boosting", "neural", "flow", "forge"} {
		if _, ok := set[name]; ok {
			ordered = append(ordered, name)
		}
	}
	if reverse {
		for left, right := 0, len(ordered)-1; left < right; left, right = left+1, right-1 {
			ordered[left], ordered[right] = ordered[right], ordered[left]
		}
	}
	return ordered
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func currentUser() string {
	if user := os.Getenv("USER"); user != "" {
		return user
	}
	return "unknown"
}

func composeWithUser(name string) (string, error) {
	path, err := docker.WriteCompose(name)
	if err == nil {
		return path, nil
	}
	return "", err
}

func labURL(name string) (string, error) {
	s, err := suiteFromCurrentState(name)
	if err != nil {
		return "", err
	}
	container, ok := suite.JupyterContainer(s)
	if !ok {
		return "", fmt.Errorf("suite %s has no JupyterLab service", name)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	status, err := docker.ContainerStatus(ctx, container)
	if err != nil {
		return "", err
	}
	if status != "running" {
		return "", fmt.Errorf("suite not running")
	}
	if reachableGradientLabURL, ok := gradientLabURL(); ok {
		return reachableGradientLabURL, nil
	}
	out, err := exec.CommandContext(ctx, "docker", "exec", container, "jupyter", "server", "list", "--json").CombinedOutput()
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var server struct {
			URL   string `json:"url"`
			Token string `json:"token"`
		}
		if err := json.Unmarshal([]byte(line), &server); err == nil && server.Token != "" {
			return "http://localhost:8888/lab?token=" + server.Token, nil
		}
	}
	return "", fmt.Errorf("unable to resolve Jupyter token")
}

func gradientLabURL() (string, bool) {
	conn, err := net.DialTimeout("tcp", "127.0.0.1:8889", 750*time.Millisecond)
	if err != nil {
		return "", false
	}
	_ = conn.Close()
	return "http://127.0.0.1:8889/lab", true
}

func streamComposeLogs(ctx context.Context, composePath, service string, lines int, follow bool, onLine func(string)) error {
	args := []string{"compose", "-f", composePath, "logs", "--no-color"}
	if follow {
		args = append(args, "--follow")
	}
	if lines <= 0 {
		lines = 50
	}
	args = append(args, "--tail", fmt.Sprintf("%d", lines))
	if service != "" {
		args = append(args, service)
	}

	cmd := exec.CommandContext(ctx, "docker", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	done := make(chan error, 2)
	consume := func(reader io.Reader) {
		scanner := bufio.NewScanner(reader)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		for scanner.Scan() {
			line := normalizeComposeLogLine(scanner.Text(), service == "")
			if line != "" {
				onLine(line)
			}
		}
		done <- scanner.Err()
	}
	go consume(stdout)
	go consume(stderr)
	for i := 0; i < 2; i++ {
		if scanErr := <-done; scanErr != nil && ctx.Err() == nil {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
			return scanErr
		}
	}
	waitErr := cmd.Wait()
	if ctx.Err() != nil {
		return nil
	}
	return waitErr
}

func normalizeComposeLogLine(line string, includeServicePrefix bool) string {
	line = strings.TrimRight(line, "\r\n")
	if line == "" {
		return ""
	}
	index := strings.Index(line, " | ")
	if index <= 0 {
		return line
	}
	service := strings.TrimSpace(line[:index])
	message := strings.TrimSpace(line[index+3:])
	if !includeServicePrefix {
		return message
	}
	if service == "" {
		return message
	}
	return "[" + service + "] " + message
}
