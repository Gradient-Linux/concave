package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/Gradient-Linux/concave/internal/auth"
	"github.com/Gradient-Linux/concave/internal/config"
	"github.com/Gradient-Linux/concave/internal/docker"
	"github.com/Gradient-Linux/concave/internal/gpu"
	"github.com/Gradient-Linux/concave/internal/suite"
	"github.com/Gradient-Linux/concave/internal/system"
)

func (a *App) handleSuiteSubroutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/suites/")
	if path == "" {
		writeError(w, http.StatusNotFound, "suite not specified")
		return
	}
	parts := strings.Split(path, "/")
	name := parts[0]

	if len(parts) == 1 {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		writeJSON(w, http.StatusOK, suiteSnapshot(name))
		return
	}

	action := parts[1]
	switch action {
	case "install":
		if err := auth.Require(ClaimsFromContextMust(r).Role, auth.ActionInstall); err != nil {
			writeError(w, http.StatusForbidden, "insufficient role")
			return
		}
		var req struct {
			ForgeComponents []string `json:"forge_components"`
		}
		if r.Body != nil {
			_ = json.NewDecoder(r.Body).Decode(&req)
		}
		selection := (*suite.ForgeSelection)(nil)
		if name == "forge" && len(req.ForgeComponents) > 0 {
			picked, err := suite.SelectionFromKeys(req.ForgeComponents)
			if err != nil {
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}
			selection = &picked
		}
		writeJSON(w, http.StatusAccepted, a.jobAccepted("install:"+name, func(rec *JobRecorder) (map[string]any, error) {
			return nil, installSuiteJob(name, selection, rec)
		}))
	case "remove":
		if err := auth.Require(ClaimsFromContextMust(r).Role, auth.ActionRemove); err != nil {
			writeError(w, http.StatusForbidden, "insufficient role")
			return
		}
		writeJSON(w, http.StatusAccepted, a.jobAccepted("remove:"+name, func(rec *JobRecorder) (map[string]any, error) {
			return nil, removeSuiteJob(name, rec)
		}))
	case "start":
		if err := auth.Require(ClaimsFromContextMust(r).Role, auth.ActionStart); err != nil {
			writeError(w, http.StatusForbidden, "insufficient role")
			return
		}
		writeJSON(w, http.StatusAccepted, a.jobAccepted("start:"+name, func(rec *JobRecorder) (map[string]any, error) {
			return nil, startSuiteJob(name, rec)
		}))
	case "stop":
		if err := auth.Require(ClaimsFromContextMust(r).Role, auth.ActionStop); err != nil {
			writeError(w, http.StatusForbidden, "insufficient role")
			return
		}
		writeJSON(w, http.StatusAccepted, a.jobAccepted("stop:"+name, func(rec *JobRecorder) (map[string]any, error) {
			return nil, stopSuiteJob(name, rec)
		}))
	case "update":
		if err := auth.Require(ClaimsFromContextMust(r).Role, auth.ActionUpdate); err != nil {
			writeError(w, http.StatusForbidden, "insufficient role")
			return
		}
		writeJSON(w, http.StatusAccepted, a.jobAccepted("update:"+name, func(rec *JobRecorder) (map[string]any, error) {
			return nil, updateSuiteJob(name, rec)
		}))
	case "rollback":
		if err := auth.Require(ClaimsFromContextMust(r).Role, auth.ActionRollback); err != nil {
			writeError(w, http.StatusForbidden, "insufficient role")
			return
		}
		writeJSON(w, http.StatusAccepted, a.jobAccepted("rollback:"+name, func(rec *JobRecorder) (map[string]any, error) {
			result, err := rollbackSuiteJob(name, rec)
			return result, err
		}))
	case "lab":
		if err := auth.Require(ClaimsFromContextMust(r).Role, auth.ActionOpenLab); err != nil {
			writeError(w, http.StatusForbidden, "insufficient role")
			return
		}
		url, err := labURL(name)
		if err != nil {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"url": url})
	case "changelog":
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		writeJSON(w, http.StatusOK, suiteChangelog(name))
	case "logs":
		a.handleSuiteLogs(w, r, name)
	default:
		writeError(w, http.StatusNotFound, "unknown suite action")
	}
}

func ClaimsFromContextMust(r *http.Request) auth.Claims {
	claims, _ := ClaimsFromContext(r.Context())
	return claims
}

func suiteChangelog(name string) map[string]any {
	s, err := suiteFromCurrentStateOrBase(name)
	if err != nil {
		return map[string]any{"error": err.Error()}
	}
	manifest, _ := config.LoadManifest()
	changes := make([]map[string]string, 0, len(s.Containers))
	for _, container := range s.Containers {
		current := container.Image
		if versions, ok := manifest[name]; ok {
			if version, ok := versions[container.Name]; ok && version.Current != "" {
				current = version.Current
			}
		}
		if current == container.Image {
			continue
		}
		changes = append(changes, map[string]string{
			"container": container.Name,
			"from":      current,
			"to":        container.Image,
		})
	}
	return map[string]any{"suite": name, "changes": changes}
}

func suiteFromCurrentStateOrBase(name string) (suite.Suite, error) {
	s, err := suiteFromCurrentState(name)
	if err == nil {
		return s, nil
	}
	return suite.Get(name)
}

func installSuiteJob(name string, selection *suite.ForgeSelection, rec *JobRecorder) error {
	unlock, err := system.Lock("api-install")
	if err != nil {
		return err
	}
	defer unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	state, _ := gpu.Detect()
	return withCapturedUI(func() error {
		return suite.Install(ctx, name, suite.InstallOptions{
			GPUAvailable:   state == gpu.GPUStateNVIDIA,
			ForgeSelection: selection,
		})
	}, rec.Line)
}

func removeSuiteJob(name string, rec *JobRecorder) error {
	unlock, err := system.Lock("api-remove")
	if err != nil {
		return err
	}
	defer unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	if exists, err := config.IsInstalled(name); err != nil {
		return err
	} else if !exists {
		return fmt.Errorf("suite %s is not installed", name)
	}

	composePath := docker.ComposePath(name)
	if _, statErr := os.Stat(composePath); statErr == nil {
		rec.Line("bringing compose stack down")
		out, err := exec.CommandContext(ctx, "docker", "compose", "-f", composePath, "down", "--rmi", "all").CombinedOutput()
		if strings.TrimSpace(string(out)) != "" {
			for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
				rec.Line(line)
			}
		}
		if err != nil {
			return err
		}
	} else if os.IsNotExist(statErr) {
		rec.Line("compose file missing — cleaning suite state directly")
		if err := cleanupSuiteContainersAPI(ctx, name); err != nil {
			return err
		}
	} else {
		return statErr
	}

	_ = os.Remove(composePath)
	if err := config.RemoveSuite(name); err != nil {
		return err
	}
	manifest, err := config.LoadManifest()
	if err != nil {
		return err
	}
	delete(manifest, name)
	if err := config.SaveManifest(manifest); err != nil {
		return err
	}
	s, err := suiteFromCurrentStateOrBase(name)
	if err == nil {
		_ = system.Deregister(s)
	}
	rec.Line("suite removed")
	return nil
}

func cleanupSuiteContainersAPI(ctx context.Context, name string) error {
	s, err := suiteFromCurrentStateOrBase(name)
	if err != nil {
		return err
	}
	names := make([]string, 0, len(s.Containers))
	for _, container := range s.Containers {
		names = append(names, container.Name)
	}
	if len(names) == 0 {
		return nil
	}
	args := append([]string{"rm", "-f"}, names...)
	if out, err := exec.CommandContext(ctx, "docker", args...).CombinedOutput(); err != nil {
		if !strings.Contains(strings.ToLower(err.Error()+" "+string(out)), "no such") {
			return err
		}
	}
	return nil
}

func startSuiteJob(name string, rec *JobRecorder) error {
	unlock, err := system.Lock("api-start")
	if err != nil {
		return err
	}
	defer unlock()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	s, err := suiteFromCurrentState(name)
	if err != nil {
		return err
	}
	rec.Line("starting compose stack")
	if err := docker.ComposeUp(ctx, docker.ComposePath(name), true); err != nil {
		return err
	}
	if err := system.Register(s); err != nil {
		return err
	}
	rec.Line("waiting for healthy containers")
	return suite.WaitHealthy(ctx, s, 60*time.Second, func(results []suite.HealthResult) {
		for _, result := range results {
			rec.Line(result.Container + " " + result.Status)
		}
	})
}

func stopSuiteJob(name string, rec *JobRecorder) error {
	unlock, err := system.Lock("api-stop")
	if err != nil {
		return err
	}
	defer unlock()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	s, err := suiteFromCurrentState(name)
	if err != nil {
		return err
	}
	rec.Line("stopping compose stack")
	if err := docker.ComposeDown(ctx, docker.ComposePath(name)); err != nil {
		return err
	}
	if err := system.Deregister(s); err != nil {
		return err
	}
	rec.Line("suite stopped")
	return nil
}

func updateSuiteJob(name string, rec *JobRecorder) error {
	unlock, err := system.Lock("api-update")
	if err != nil {
		return err
	}
	defer unlock()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	s, err := suiteFromCurrentState(name)
	if err != nil {
		return err
	}
	manifest, err := config.LoadManifest()
	if err != nil {
		return err
	}
	for _, container := range s.Containers {
		rec.Line("pulling " + container.Image)
		if err := docker.PullWithRollbackSafety(ctx, container.Image, func(line string) {
			rec.Line(line)
		}); err != nil {
			return err
		}
		manifest = config.RecordUpdate(manifest, s.Name, container.Name, container.Image)
	}
	if err := config.SaveManifest(manifest); err != nil {
		return err
	}
	if _, err := docker.WriteCompose(name); err != nil {
		return err
	}
	if err := docker.ComposeUp(ctx, docker.ComposePath(name), true); err != nil {
		return err
	}
	return suite.WaitHealthy(ctx, s, 60*time.Second, func(results []suite.HealthResult) {
		for _, result := range results {
			rec.Line(result.Container + " " + result.Status)
		}
	})
}

func rollbackSuiteJob(name string, rec *JobRecorder) (map[string]any, error) {
	unlock, err := system.Lock("api-rollback")
	if err != nil {
		return nil, err
	}
	defer unlock()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	manifest, err := config.LoadManifest()
	if err != nil {
		return nil, err
	}
	manifest, err = config.SwapForRollback(manifest, name)
	if err != nil {
		return nil, err
	}
	if err := config.SaveManifest(manifest); err != nil {
		return nil, err
	}
	if _, err := docker.WriteCompose(name); err != nil {
		return nil, err
	}
	if err := docker.ComposeDown(ctx, docker.ComposePath(name)); err != nil {
		return nil, err
	}
	if err := docker.ComposeUp(ctx, docker.ComposePath(name), true); err != nil {
		return nil, err
	}
	s, err := suiteFromCurrentState(name)
	if err != nil {
		return nil, err
	}
	if err := suite.WaitHealthy(ctx, s, 60*time.Second, func(results []suite.HealthResult) {
		for _, result := range results {
			rec.Line(result.Container + " " + result.Status)
		}
	}); err != nil {
		return nil, err
	}
	restored := ""
	for _, container := range s.Containers {
		if versions, ok := manifest[name]; ok {
			if version, ok := versions[container.Name]; ok {
				restored = version.Current
				break
			}
		}
	}
	return map[string]any{"restored": restored}, nil
}
