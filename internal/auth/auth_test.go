package auth

import (
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func restoreAuthGlobals(t *testing.T) {
	t.Helper()
	oldLookupUser := lookupUser
	oldLookupGroupID := lookupGroupID
	oldLookupGroupIDs := lookupGroupIDs
	oldCurrentUser := currentUserFn
	ResetCLIAuth()
	t.Cleanup(func() {
		lookupUser = oldLookupUser
		lookupGroupID = oldLookupGroupID
		lookupGroupIDs = oldLookupGroupIDs
		currentUserFn = oldCurrentUser
		ResetCLIAuth()
	})
}

func TestResolveRoleAdminGroup(t *testing.T) {
	restoreAuthGlobals(t)

	lookupUser = func(username string) (*user.User, error) {
		return &user.User{Username: username}, nil
	}
	lookupGroupIDs = func(*user.User) ([]string, error) { return []string{"10", "20"}, nil }
	lookupGroupID = func(gid string) (*user.Group, error) {
		switch gid {
		case "10":
			return &user.Group{Name: "gradient-viewer"}, nil
		case "20":
			return &user.Group{Name: "gradient-admin"}, nil
		default:
			return &user.Group{Name: "other"}, nil
		}
	}

	role, err := ResolveRole("alice")
	if err != nil {
		t.Fatalf("ResolveRole() error = %v", err)
	}
	if role != RoleAdmin {
		t.Fatalf("ResolveRole() role = %v, want %v", role, RoleAdmin)
	}
}

func TestResolveRoleNoGradientGroup(t *testing.T) {
	restoreAuthGlobals(t)

	lookupUser = func(username string) (*user.User, error) {
		return &user.User{Username: username}, nil
	}
	lookupGroupIDs = func(*user.User) ([]string, error) { return []string{"10"}, nil }
	lookupGroupID = func(gid string) (*user.Group, error) {
		return &user.Group{Name: "docker"}, nil
	}

	_, err := ResolveRole("bob")
	if err == nil || err.Error() == "" {
		t.Fatalf("ResolveRole() error = %v, want group membership error", err)
	}
}

func TestResolveRoleMultipleGroupsHighestWins(t *testing.T) {
	restoreAuthGlobals(t)

	lookupUser = func(username string) (*user.User, error) {
		return &user.User{Username: username}, nil
	}
	lookupGroupIDs = func(*user.User) ([]string, error) { return []string{"10", "11", "12"}, nil }
	lookupGroupID = func(gid string) (*user.Group, error) {
		switch gid {
		case "10":
			return &user.Group{Name: "gradient-viewer"}, nil
		case "11":
			return &user.Group{Name: "gradient-developer"}, nil
		case "12":
			return &user.Group{Name: "gradient-operator"}, nil
		default:
			return &user.Group{Name: "docker"}, nil
		}
	}

	role, err := ResolveRole("carol")
	if err != nil {
		t.Fatalf("ResolveRole() error = %v", err)
	}
	if role != RoleOperator {
		t.Fatalf("ResolveRole() role = %v, want %v", role, RoleOperator)
	}
}

func TestCanRolesCumulative(t *testing.T) {
	if !Can(RoleAdmin, ActionViewLogs) {
		t.Fatal("admin should inherit viewer permissions")
	}
	if !Can(RoleDeveloper, ActionOpenLab) {
		t.Fatal("developer should open lab")
	}
	if Can(RoleDeveloper, ActionInstall) {
		t.Fatal("developer must not install suites")
	}
	if !Can(RoleOperator, ActionInstall) {
		t.Fatal("operator should install suites")
	}
	if Can(RoleOperator, ActionReboot) {
		t.Fatal("operator must not reboot")
	}
}

func TestAllowedActionsIncludesInherited(t *testing.T) {
	actions := AllowedActions(RoleAdmin)
	if len(actions) == 0 {
		t.Fatal("AllowedActions(admin) returned no actions")
	}
	seen := map[Action]bool{}
	for _, action := range actions {
		seen[action] = true
	}
	for _, action := range []Action{ActionViewLogs, ActionInstall, ActionReboot} {
		if !seen[action] {
			t.Fatalf("AllowedActions(admin) missing %s", action)
		}
	}
}

func TestIssueAndValidateToken(t *testing.T) {
	cfg := TokenConfig{SigningKey: []byte("01234567890123456789012345678901"), TokenTTL: 24 * time.Hour}
	token, err := IssueToken(cfg, "alice", RoleAdmin)
	if err != nil {
		t.Fatalf("IssueToken() error = %v", err)
	}

	claims, err := ValidateToken(cfg, token)
	if err != nil {
		t.Fatalf("ValidateToken() error = %v", err)
	}
	if claims.Subject != "alice" || claims.Role != RoleAdmin {
		t.Fatalf("ValidateToken() = %#v, want subject alice role admin", claims)
	}
}

func TestValidateTokenExpired(t *testing.T) {
	cfg := TokenConfig{SigningKey: []byte("01234567890123456789012345678901"), TokenTTL: 24 * time.Hour}
	claims := Claims{
		Role: RoleViewer,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "alice",
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(-time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC().Add(-2 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(cfg.SigningKey)
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}

	if _, err := ValidateToken(cfg, tokenStr); err == nil {
		t.Fatal("ValidateToken() error = nil, want expiry error")
	}
}

func TestValidateTokenWrongSignature(t *testing.T) {
	cfg := TokenConfig{SigningKey: []byte("01234567890123456789012345678901"), TokenTTL: 24 * time.Hour}
	other := TokenConfig{SigningKey: []byte("abcdefghijklmnopqrstuvwxyz123456"), TokenTTL: 24 * time.Hour}
	token, err := IssueToken(other, "alice", RoleAdmin)
	if err != nil {
		t.Fatalf("IssueToken() error = %v", err)
	}
	if _, err := ValidateToken(cfg, token); err == nil {
		t.Fatal("ValidateToken() error = nil, want signature error")
	}
}

func TestRefreshTokenWithinWindow(t *testing.T) {
	cfg := TokenConfig{SigningKey: []byte("01234567890123456789012345678901"), TokenTTL: 24 * time.Hour}
	claims := Claims{
		Role: RoleDeveloper,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "alice",
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC().Add(-23 * time.Hour)),
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(30 * time.Minute)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(cfg.SigningKey)
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}
	refreshed, err := RefreshToken(cfg, tokenStr)
	if err != nil {
		t.Fatalf("RefreshToken() error = %v", err)
	}
	if refreshed == tokenStr {
		t.Fatal("RefreshToken() returned original token")
	}
}

func TestRefreshTokenTooEarly(t *testing.T) {
	cfg := TokenConfig{SigningKey: []byte("01234567890123456789012345678901"), TokenTTL: 24 * time.Hour}
	token, err := IssueToken(cfg, "alice", RoleDeveloper)
	if err != nil {
		t.Fatalf("IssueToken() error = %v", err)
	}
	if _, err := RefreshToken(cfg, token); err == nil {
		t.Fatal("RefreshToken() error = nil, want too-early error")
	}
}

func TestCLIRoleAdminGroup(t *testing.T) {
	restoreAuthGlobals(t)

	currentUserFn = func() (*user.User, error) {
		return &user.User{Username: "alice"}, nil
	}
	lookupUser = func(username string) (*user.User, error) {
		return &user.User{Username: username}, nil
	}
	lookupGroupIDs = func(*user.User) ([]string, error) { return []string{"10"}, nil }
	lookupGroupID = func(gid string) (*user.Group, error) {
		return &user.Group{Name: "gradient-admin"}, nil
	}

	role, err := CLIRole()
	if err != nil {
		t.Fatalf("CLIRole() error = %v", err)
	}
	if role != RoleAdmin {
		t.Fatalf("CLIRole() role = %v, want %v", role, RoleAdmin)
	}
}

func TestCLIRoleNoGroup(t *testing.T) {
	restoreAuthGlobals(t)

	currentUserFn = func() (*user.User, error) {
		return &user.User{Username: "bob"}, nil
	}
	lookupUser = func(username string) (*user.User, error) {
		return &user.User{Username: username}, nil
	}
	lookupGroupIDs = func(*user.User) ([]string, error) { return []string{"10"}, nil }
	lookupGroupID = func(gid string) (*user.Group, error) {
		return &user.Group{Name: "docker"}, nil
	}

	if _, err := CLIRole(); err == nil {
		t.Fatal("CLIRole() error = nil, want missing group error")
	}
}

func TestRequireCLIRoleNamesRequiredGroup(t *testing.T) {
	restoreAuthGlobals(t)

	restore := SetCLIRoleForTesting(RoleViewer)
	defer restore()

	err := RequireCLIRole(RoleOperator)
	if err == nil {
		t.Fatal("RequireCLIRole() error = nil, want permission error")
	}
	if got := err.Error(); got == "" || !containsAll(got, "gradient-operator", "viewer") {
		t.Fatalf("RequireCLIRole() error = %q", got)
	}
}

func TestSessionRoundTrip(t *testing.T) {
	restoreAuthGlobals(t)

	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	session := Session{
		Token:     "token",
		Username:  "alice",
		Role:      RoleAdmin,
		ExpiresAt: time.Now().UTC().Add(time.Hour),
	}
	if err := SaveSession(session); err != nil {
		t.Fatalf("SaveSession() error = %v", err)
	}

	path := SessionPath()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat(%s) error = %v", path, err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("session mode = %#o, want 0600", got)
	}

	loaded, err := LoadSession()
	if err != nil {
		t.Fatalf("LoadSession() error = %v", err)
	}
	if loaded.Username != session.Username || loaded.Role != session.Role {
		t.Fatalf("LoadSession() = %#v, want %#v", loaded, session)
	}

	if err := ClearSession(); err != nil {
		t.Fatalf("ClearSession() error = %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("session file still exists at %s", path)
	}
}

func TestLoadOrCreateTokenConfigHonorsOverride(t *testing.T) {
	dir := t.TempDir()
	override := filepath.Join(dir, "auth.json")
	t.Setenv("CONCAVE_AUTH_CONFIG_PATH", override)

	cfg, err := LoadOrCreateTokenConfig(filepath.Join(dir, "workspace"))
	if err != nil {
		t.Fatalf("LoadOrCreateTokenConfig() error = %v", err)
	}
	if len(cfg.SigningKey) != 32 {
		t.Fatalf("signing key length = %d, want 32", len(cfg.SigningKey))
	}
	if _, err := os.Stat(override); err != nil {
		t.Fatalf("Stat(%s) error = %v", override, err)
	}
}

func containsAll(value string, parts ...string) bool {
	for _, part := range parts {
		if !strings.Contains(value, part) {
			return false
		}
	}
	return true
}
