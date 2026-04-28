package lab

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

type fakeDriver struct {
	name   string
	owns   bool
	launch func(context.Context, EnvSpec) (Env, error)
	dest   func(context.Context, Env) error
	arch   func(context.Context, Env) (ArchiveRef, error)

	mu           sync.Mutex
	destroyed    []string
	archived     []string
	extended     []time.Time
	inspectCalls int
}

func (f *fakeDriver) Name() string  { return f.name }
func (f *fakeDriver) OwnsTTL() bool { return f.owns }

func (f *fakeDriver) Launch(ctx context.Context, spec EnvSpec) (Env, error) {
	if f.launch != nil {
		return f.launch(ctx, spec)
	}
	id, err := newEnvID()
	if err != nil {
		return Env{}, err
	}
	return Env{
		ID:           id,
		Driver:       f.name,
		Owner:        spec.Owner,
		Image:        spec.Image,
		Status:       StatusRunning,
		HotTierPath:  spec.HotTierPath,
		ColdTierPath: spec.ColdTierPath,
		CreatedAt:    time.Unix(0, 0),
		ExpiresAt:    time.Unix(0, 0).Add(spec.TTL),
	}, nil
}

func (f *fakeDriver) Inspect(ctx context.Context, env Env) (Env, error) {
	f.mu.Lock()
	f.inspectCalls++
	f.mu.Unlock()
	return env, nil
}

func (f *fakeDriver) ExtendTTL(_ context.Context, env Env, until time.Time) (Env, error) {
	f.mu.Lock()
	f.extended = append(f.extended, until)
	f.mu.Unlock()
	env.ExpiresAt = until
	return env, nil
}

func (f *fakeDriver) Archive(ctx context.Context, env Env) (ArchiveRef, error) {
	if f.arch != nil {
		return f.arch(ctx, env)
	}
	f.mu.Lock()
	f.archived = append(f.archived, env.ID)
	f.mu.Unlock()
	return ArchiveRef{EnvID: env.ID, Path: "/cold/" + env.ID + ".tar.gz", CreatedAt: time.Unix(0, 0)}, nil
}

func (f *fakeDriver) Destroy(ctx context.Context, env Env) error {
	if f.dest != nil {
		return f.dest(ctx, env)
	}
	f.mu.Lock()
	f.destroyed = append(f.destroyed, env.ID)
	f.mu.Unlock()
	return nil
}

func tempManager(t *testing.T, driver Driver, storage StorageConfig) *Manager {
	t.Helper()
	tmp := t.TempDir()
	store, err := NewStoreAt(filepath.Join(tmp, "lab-envs.json"))
	if err != nil {
		t.Fatalf("NewStoreAt: %v", err)
	}
	registry := NewRegistry()
	registry.Register(driver)
	return NewManager(registry, store, storage)
}

func TestStoreCreateAndList(t *testing.T) {
	tmp := t.TempDir()
	store, err := NewStoreAt(filepath.Join(tmp, "lab-envs.json"))
	if err != nil {
		t.Fatalf("NewStoreAt: %v", err)
	}
	created, err := store.Create(Env{CreatedAt: time.Unix(1, 0)})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.ID == "" {
		t.Fatal("expected auto-assigned ID")
	}
	envs := store.List()
	if len(envs) != 1 || envs[0].ID != created.ID {
		t.Fatalf("List = %+v", envs)
	}
	reopened, err := NewStoreAt(filepath.Join(tmp, "lab-envs.json"))
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	if _, ok := reopened.Get(created.ID); !ok {
		t.Fatal("expected record after reopen")
	}
}

func TestStoreUpdateAndDelete(t *testing.T) {
	tmp := t.TempDir()
	store, _ := NewStoreAt(filepath.Join(tmp, "lab-envs.json"))
	env, _ := store.Create(Env{ID: "env-1", CreatedAt: time.Unix(1, 0), Status: StatusRunning})
	env.Status = StatusArchived
	if err := store.Update(env); err != nil {
		t.Fatalf("Update: %v", err)
	}
	got, _ := store.Get("env-1")
	if got.Status != StatusArchived {
		t.Fatalf("status = %q", got.Status)
	}
	if err := store.Delete("env-1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if len(store.List()) != 0 {
		t.Fatal("expected empty list after delete")
	}
}

func TestStorageConfigRoundTrip(t *testing.T) {
	tmp := t.TempDir()
	prev := storagePath
	storagePath = func() string { return filepath.Join(tmp, "lab.json") }
	t.Cleanup(func() { storagePath = prev })

	if err := SaveStorage(StorageConfig{HotTier: "/tmp/hot", ColdTier: "/tmp/cold"}); err != nil {
		t.Fatalf("SaveStorage: %v", err)
	}
	loaded, err := LoadStorage()
	if err != nil {
		t.Fatalf("LoadStorage: %v", err)
	}
	if loaded.HotTier != "/tmp/hot" || loaded.ColdTier != "/tmp/cold" {
		t.Fatalf("loaded = %+v", loaded)
	}

	if err := SaveStorage(StorageConfig{HotTier: "/a", ColdTier: "/a"}); err == nil {
		t.Fatal("expected duplicate path rejection")
	}
	if err := SaveStorage(StorageConfig{HotTier: "relative", ColdTier: "/b"}); err == nil {
		t.Fatal("expected absolute-path rejection")
	}
}

func TestRegistryActiveAndLookup(t *testing.T) {
	r := NewRegistry()
	r.Register(&fakeDriver{name: "docker"})
	r.Register(&fakeDriver{name: "slurm", owns: true})
	if r.Active() != "docker" {
		t.Fatalf("Active = %q", r.Active())
	}
	if err := r.SetActive("slurm"); err != nil {
		t.Fatalf("SetActive: %v", err)
	}
	if r.Active() != "slurm" {
		t.Fatalf("Active = %q", r.Active())
	}
	if err := r.SetActive("proxmox"); !errors.Is(err, ErrDriverNotRegistered) {
		t.Fatalf("SetActive unknown = %v", err)
	}
	if _, err := r.Get("docker"); err != nil {
		t.Fatalf("Get docker: %v", err)
	}
}

func TestManagerLaunchExtendArchive(t *testing.T) {
	driver := &fakeDriver{name: "docker"}
	now := time.Unix(1_000, 0)
	mgr := tempManager(t, driver, StorageConfig{HotTier: "/h", ColdTier: "/c"})
	mgr.SetNowFn(func() time.Time { return now })
	audit := &MemoryAuditWriter{}
	mgr.SetAuditWriter(audit)

	launched, err := mgr.Launch(context.Background(), EnvSpec{
		Owner: "alice", Image: "jupyter/minimal", TTL: 30 * time.Minute,
	})
	if err != nil {
		t.Fatalf("Launch: %v", err)
	}
	if launched.Driver != "docker" || launched.HotTierPath != "/h" || launched.ColdTierPath != "/c" {
		t.Fatalf("launched = %+v", launched)
	}
	extended, err := mgr.ExtendTTL(context.Background(), launched.ID, time.Hour)
	if err != nil {
		t.Fatalf("ExtendTTL: %v", err)
	}
	if got := extended.ExpiresAt.Sub(launched.ExpiresAt); got != time.Hour {
		t.Fatalf("extension delta = %s", got)
	}
	archived, err := mgr.ArchiveAndDestroy(context.Background(), launched.ID)
	if err != nil {
		t.Fatalf("ArchiveAndDestroy: %v", err)
	}
	if archived.Status != StatusArchived {
		t.Fatalf("archived.Status = %q", archived.Status)
	}
	if len(driver.archived) != 1 || len(driver.destroyed) != 1 {
		t.Fatalf("driver counts: archived=%d destroyed=%d", len(driver.archived), len(driver.destroyed))
	}
	events := audit.Snapshot()
	actions := make([]string, 0, len(events))
	for _, e := range events {
		actions = append(actions, e.Action)
	}
	wantActions := []string{"lab.launch", "lab.extend", "lab.archive"}
	if len(events) != len(wantActions) {
		t.Fatalf("audit event count = %d (%v)", len(events), actions)
	}
	for i, want := range wantActions {
		if events[i].Action != want {
			t.Fatalf("audit[%d].Action = %q, want %q", i, events[i].Action, want)
		}
		if events[i].PeerID == "" {
			t.Fatalf("audit[%d].PeerID empty", i)
		}
	}
}

func TestReaperArchivesOnlyExpired(t *testing.T) {
	driver := &fakeDriver{name: "docker"}
	now := time.Unix(1_000, 0)
	mgr := tempManager(t, driver, StorageConfig{HotTier: "/h", ColdTier: "/c"})
	mgr.SetNowFn(func() time.Time { return now })

	expired, _ := mgr.Launch(context.Background(), EnvSpec{Owner: "a", Image: "i", TTL: time.Minute})
	alive, _ := mgr.Launch(context.Background(), EnvSpec{Owner: "a", Image: "i", TTL: time.Hour})

	mgr.SetNowFn(func() time.Time { return now.Add(2 * time.Minute) })
	if errs := mgr.ReapExpired(context.Background()); len(errs) != 0 {
		t.Fatalf("reap errors: %v", errs)
	}
	expiredEnv, _ := mgr.Get(context.Background(), expired.ID)
	aliveEnv, _ := mgr.Get(context.Background(), alive.ID)
	if expiredEnv.Status != StatusArchived {
		t.Fatalf("expired.Status = %q", expiredEnv.Status)
	}
	if aliveEnv.Status != StatusRunning {
		t.Fatalf("alive.Status = %q", aliveEnv.Status)
	}
}

func TestReaperSkipsSchedulerOwnedTTL(t *testing.T) {
	driver := &fakeDriver{name: "slurm", owns: true}
	now := time.Unix(1_000, 0)
	mgr := tempManager(t, driver, StorageConfig{HotTier: "/h", ColdTier: "/c"})
	mgr.SetNowFn(func() time.Time { return now })

	env, _ := mgr.Launch(context.Background(), EnvSpec{Owner: "a", Image: "i", TTL: time.Minute})
	mgr.SetNowFn(func() time.Time { return now.Add(time.Hour) })
	if errs := mgr.ReapExpired(context.Background()); len(errs) != 0 {
		t.Fatalf("reap errors: %v", errs)
	}
	refreshed, _ := mgr.Get(context.Background(), env.ID)
	if refreshed.Status != StatusRunning {
		t.Fatalf("scheduler-owned env archived unexpectedly: %q", refreshed.Status)
	}
	if len(driver.archived) != 0 {
		t.Fatalf("scheduler-owned driver should not archive, got %d", len(driver.archived))
	}
}

type stubRunner struct {
	calls []struct {
		name string
		args []string
	}
	out []byte
	err error
}

func (s *stubRunner) RunCommand(_ context.Context, name string, args ...string) ([]byte, error) {
	s.calls = append(s.calls, struct {
		name string
		args []string
	}{name, append([]string(nil), args...)})
	return s.out, s.err
}

func TestDockerDriverLaunchComposesDockerRun(t *testing.T) {
	runner := &stubRunner{out: []byte("abc123\n")}
	driver := NewDockerDriver(DockerDriverConfig{
		Runner:    runner,
		AllocPort: func() (int, error) { return 18888, nil },
		NowFn:     func() time.Time { return time.Unix(0, 0) },
	})
	hot := t.TempDir()

	env, err := driver.Launch(context.Background(), EnvSpec{
		Owner: "alice", Image: "jupyter/minimal", TTL: time.Hour,
		HotTierPath: hot, ColdTierPath: "/cold",
		GPUs: 1, CPURequest: "2", MemRequest: "4g",
	})
	if err != nil {
		t.Fatalf("Launch: %v", err)
	}
	if env.ContainerID != "abc123" {
		t.Fatalf("ContainerID = %q", env.ContainerID)
	}
	if env.JupyterURL == "" || env.Token == "" {
		t.Fatal("JupyterURL/Token should be populated")
	}
	if len(runner.calls) != 1 || runner.calls[0].name != "docker" {
		t.Fatalf("calls = %+v", runner.calls)
	}
	args := runner.calls[0].args
	mustContain(t, args, "--gpus")
	mustContain(t, args, "--cpus")
	mustContain(t, args, "--memory")
	mustContain(t, args, "jupyter/minimal")
}

func TestDockerDriverArchiveWritesTarball(t *testing.T) {
	runner := &stubRunner{}
	driver := NewDockerDriver(DockerDriverConfig{
		Runner:    runner,
		AllocPort: func() (int, error) { return 1, nil },
		NowFn:     func() time.Time { return time.Unix(1, 0) },
	})
	hot := t.TempDir()
	cold := t.TempDir()
	if err := os.WriteFile(filepath.Join(hot, "hello.txt"), []byte("hi"), 0o600); err != nil {
		t.Fatalf("write hot file: %v", err)
	}

	ref, err := driver.Archive(context.Background(), Env{
		ID: "env-xyz", HotTierPath: hot, ColdTierPath: cold, Driver: "docker",
	})
	if err != nil {
		t.Fatalf("Archive: %v", err)
	}
	if ref.Path != filepath.Join(cold, "env-xyz.tar.gz") {
		t.Fatalf("ref.Path = %q", ref.Path)
	}
	if ref.SizeBytes <= 0 {
		t.Fatalf("ref.SizeBytes = %d", ref.SizeBytes)
	}
	manifest, err := os.ReadFile(ref.Path + ".json")
	if err != nil {
		t.Fatalf("manifest: %v", err)
	}
	var decoded ArchiveRef
	if err := json.Unmarshal(manifest, &decoded); err != nil {
		t.Fatalf("manifest decode: %v", err)
	}
	if decoded.EnvID != "env-xyz" {
		t.Fatalf("manifest envID = %q", decoded.EnvID)
	}
}

func TestDockerDriverValidateSpec(t *testing.T) {
	cases := map[string]EnvSpec{
		"missing image": {Owner: "a", TTL: time.Minute, HotTierPath: "/h"},
		"missing owner": {Image: "i", TTL: time.Minute, HotTierPath: "/h"},
		"zero ttl":      {Owner: "a", Image: "i", HotTierPath: "/h"},
		"excess ttl":    {Owner: "a", Image: "i", TTL: 10 * 24 * time.Hour, HotTierPath: "/h"},
		"missing hot":   {Owner: "a", Image: "i", TTL: time.Minute},
	}
	for name, spec := range cases {
		t.Run(name, func(t *testing.T) {
			if err := validateSpec(spec); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestEnvRemainingAndExpired(t *testing.T) {
	base := time.Unix(1_000, 0)
	env := Env{ExpiresAt: base.Add(5 * time.Minute)}
	if env.Expired(base) {
		t.Fatal("should not be expired yet")
	}
	if env.Expired(base.Add(4 * time.Minute)) {
		t.Fatal("should not be expired at t+4m")
	}
	if !env.Expired(base.Add(6 * time.Minute)) {
		t.Fatal("should be expired at t+6m")
	}
	if got := env.Remaining(base); got != 5*time.Minute {
		t.Fatalf("Remaining = %s", got)
	}
	if got := env.Remaining(base.Add(10 * time.Minute)); got != 0 {
		t.Fatalf("Remaining after expiry = %s", got)
	}
}

func mustContain(t *testing.T, args []string, needle string) {
	t.Helper()
	for _, a := range args {
		if a == needle {
			return
		}
	}
	t.Fatalf("args missing %q: %v", needle, args)
}
