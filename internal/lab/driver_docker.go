package lab

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Runner abstracts external command execution so the driver can be unit-tested
// without a real Docker daemon.
type Runner interface {
	RunCommand(ctx context.Context, name string, args ...string) ([]byte, error)
}

type defaultRunner struct{}

func (defaultRunner) RunCommand(ctx context.Context, name string, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, name, args...).CombinedOutput()
}

// PortAllocator picks a host port for a Jupyter environment. Defaults to a
// net.Listen-based allocator; tests override it.
type PortAllocator func() (int, error)

func allocateRandomPort() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, fmt.Errorf("allocate port: %w", err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

// DockerDriverConfig holds injectable dependencies for the Docker driver.
type DockerDriverConfig struct {
	Runner    Runner
	AllocPort PortAllocator
	NowFn     func() time.Time
}

// NewDockerDriver builds a Docker-backed Driver. Any nil field falls back to
// the production default.
func NewDockerDriver(cfg DockerDriverConfig) *DockerDriver {
	if cfg.Runner == nil {
		cfg.Runner = defaultRunner{}
	}
	if cfg.AllocPort == nil {
		cfg.AllocPort = allocateRandomPort
	}
	if cfg.NowFn == nil {
		cfg.NowFn = time.Now
	}
	return &DockerDriver{cfg: cfg}
}

// DockerDriver is the reference Driver implementation. It launches each env
// as a dedicated `docker run` container with the hot-tier directory mounted
// at /home/jovyan/work and a randomly-allocated host port forwarded to 8888.
type DockerDriver struct {
	cfg DockerDriverConfig
}

// Name implements Driver.
func (d *DockerDriver) Name() string { return "docker" }

// OwnsTTL implements Driver. The Docker driver does not own TTL — the reaper
// enforces expiry.
func (d *DockerDriver) OwnsTTL() bool { return false }

// Launch creates the hot-tier scratch directory, pulls the image (already
// cached when possible), runs the container detached, and returns the
// populated Env record.
func (d *DockerDriver) Launch(ctx context.Context, spec EnvSpec) (Env, error) {
	if err := validateSpec(spec); err != nil {
		return Env{}, err
	}
	if err := ensureDir(spec.HotTierPath); err != nil {
		return Env{}, err
	}

	id, err := newEnvID()
	if err != nil {
		return Env{}, err
	}
	envHot := filepath.Join(spec.HotTierPath, id)
	if err := ensureDir(envHot); err != nil {
		return Env{}, err
	}

	port, err := d.cfg.AllocPort()
	if err != nil {
		return Env{}, err
	}
	token, err := newJupyterToken()
	if err != nil {
		return Env{}, err
	}

	containerName := "gradient-lab-" + id
	args := []string{
		"run", "-d",
		"--name", containerName,
		"--label", "gradient.lab.env_id=" + id,
		"--label", "gradient.lab.owner=" + spec.Owner,
		"-p", fmt.Sprintf("127.0.0.1:%d:8888", port),
		"-v", envHot + ":/home/jovyan/work",
		"-e", "JUPYTER_TOKEN=" + token,
	}
	if spec.CPURequest != "" {
		args = append(args, "--cpus", spec.CPURequest)
	}
	if spec.MemRequest != "" {
		args = append(args, "--memory", spec.MemRequest)
	}
	if spec.GPUs > 0 {
		args = append(args, "--gpus", strconv.Itoa(spec.GPUs))
	}
	args = append(args, spec.Image,
		"start-notebook.sh",
		"--NotebookApp.token="+token,
		"--NotebookApp.ip=0.0.0.0",
	)

	out, err := d.cfg.Runner.RunCommand(ctx, "docker", args...)
	if err != nil {
		return Env{}, fmt.Errorf("docker run: %w: %s", err, strings.TrimSpace(string(out)))
	}

	now := d.cfg.NowFn()
	env := Env{
		ID:           id,
		Driver:       d.Name(),
		Owner:        spec.Owner,
		Image:        spec.Image,
		DisplayName:  spec.DisplayName,
		Status:       StatusRunning,
		ContainerID:  strings.TrimSpace(string(out)),
		JupyterURL:   fmt.Sprintf("http://127.0.0.1:%d/lab?token=%s", port, token),
		Token:        token,
		GPUs:         spec.GPUs,
		CPURequest:   spec.CPURequest,
		MemRequest:   spec.MemRequest,
		HotTierPath:  envHot,
		ColdTierPath: spec.ColdTierPath,
		CreatedAt:    now,
		ExpiresAt:    now.Add(spec.TTL),
	}
	return env, nil
}

// Inspect implements Driver by running `docker inspect --format {{.State.Status}}`.
func (d *DockerDriver) Inspect(ctx context.Context, env Env) (Env, error) {
	if env.ContainerID == "" {
		return env, nil
	}
	out, err := d.cfg.Runner.RunCommand(ctx, "docker", "inspect", "--format", "{{.State.Status}}", env.ContainerID)
	if err != nil {
		env.LastError = strings.TrimSpace(string(out))
		return env, nil
	}
	state := strings.TrimSpace(string(out))
	switch state {
	case "running", "restarting":
		env.Status = StatusRunning
	case "exited", "dead":
		if env.Status != StatusArchived {
			env.Status = StatusFailed
			env.LastError = "container exited: " + state
		}
	case "removing", "paused", "created":
		// leave status untouched; reaper will follow up
	}
	return env, nil
}

// ExtendTTL updates the env's expiry in-place. The reaper picks up the new
// deadline on its next tick.
func (d *DockerDriver) ExtendTTL(_ context.Context, env Env, until time.Time) (Env, error) {
	if until.Before(d.cfg.NowFn()) {
		return env, errors.New("extension target is in the past")
	}
	env.ExpiresAt = until
	if env.Status == StatusExpiring {
		env.Status = StatusRunning
	}
	return env, nil
}

// Archive creates `<cold_tier>/<env_id>.tar.gz` and a sidecar manifest.
func (d *DockerDriver) Archive(_ context.Context, env Env) (ArchiveRef, error) {
	if env.ColdTierPath == "" {
		return ArchiveRef{}, errors.New("cold_tier_path is not configured")
	}
	if err := ensureDir(env.ColdTierPath); err != nil {
		return ArchiveRef{}, err
	}
	archivePath := filepath.Join(env.ColdTierPath, env.ID+".tar.gz")
	size, err := archiveHotTier(env.HotTierPath, archivePath)
	if err != nil {
		return ArchiveRef{}, err
	}
	manifest := ArchiveRef{
		EnvID:     env.ID,
		Path:      archivePath,
		PeerID:    localPeerID(),
		SizeBytes: size,
		CreatedAt: d.cfg.NowFn(),
		OriginSpec: EnvSpec{
			Owner:        env.Owner,
			Image:        env.Image,
			DisplayName:  env.DisplayName,
			GPUs:         env.GPUs,
			CPURequest:   env.CPURequest,
			MemRequest:   env.MemRequest,
			HotTierPath:  env.HotTierPath,
			ColdTierPath: env.ColdTierPath,
			Driver:       env.Driver,
		},
	}
	if err := writeManifest(archivePath+".json", manifest); err != nil {
		return ArchiveRef{}, err
	}
	return manifest, nil
}

// Destroy stops and removes the container, and wipes the hot-tier directory.
func (d *DockerDriver) Destroy(ctx context.Context, env Env) error {
	if env.ContainerID != "" {
		if _, err := d.cfg.Runner.RunCommand(ctx, "docker", "rm", "-f", env.ContainerID); err != nil {
			return fmt.Errorf("docker rm %s: %w", env.ContainerID, err)
		}
	}
	if env.HotTierPath != "" {
		if err := os.RemoveAll(env.HotTierPath); err != nil {
			return fmt.Errorf("remove hot tier %s: %w", env.HotTierPath, err)
		}
	}
	return nil
}

func validateSpec(spec EnvSpec) error {
	if strings.TrimSpace(spec.Image) == "" {
		return errors.New("image is required")
	}
	if strings.TrimSpace(spec.Owner) == "" {
		return errors.New("owner is required")
	}
	if spec.TTL <= 0 {
		return errors.New("ttl must be positive")
	}
	if spec.TTL > 7*24*time.Hour {
		return errors.New("ttl may not exceed 7 days")
	}
	if spec.HotTierPath == "" {
		return errors.New("hot_tier_path is required")
	}
	return nil
}

func ensureDir(path string) error {
	if path == "" {
		return nil
	}
	return os.MkdirAll(path, 0o755)
}

func newJupyterToken() (string, error) {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", fmt.Errorf("generate jupyter token: %w", err)
	}
	return hex.EncodeToString(buf[:]), nil
}

func archiveHotTier(src, dest string) (int64, error) {
	out, err := os.Create(dest)
	if err != nil {
		return 0, fmt.Errorf("create archive: %w", err)
	}
	defer out.Close()

	gz := gzip.NewWriter(out)
	tw := tar.NewWriter(gz)
	err = filepath.Walk(src, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		hdr, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		hdr.Name = rel
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(tw, f)
		return err
	})
	if err != nil {
		return 0, fmt.Errorf("walk hot tier: %w", err)
	}
	if err := tw.Close(); err != nil {
		return 0, fmt.Errorf("close tar: %w", err)
	}
	if err := gz.Close(); err != nil {
		return 0, fmt.Errorf("close gzip: %w", err)
	}
	stat, err := out.Stat()
	if err != nil {
		return 0, fmt.Errorf("stat archive: %w", err)
	}
	return stat.Size(), nil
}

func writeManifest(path string, manifest ArchiveRef) error {
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("encode archive manifest: %w", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write archive manifest: %w", err)
	}
	return nil
}
