package auth

import (
	"fmt"
	"sync"
)

var (
	cliRoleMu      sync.RWMutex
	cliRole        Role
	cliRoleLoaded  bool
	cliRoleErr     error
	currentUserFn  = userCurrentFunc
)

// CLIRole resolves the current Unix user's Gradient role.
func CLIRole() (Role, error) {
	cliRoleMu.RLock()
	if cliRoleLoaded {
		defer cliRoleMu.RUnlock()
		return cliRole, cliRoleErr
	}
	cliRoleMu.RUnlock()
	return InitCLIRole()
}

// InitCLIRole resolves and caches the current CLI user's role.
func InitCLIRole() (Role, error) {
	cliRoleMu.Lock()
	defer cliRoleMu.Unlock()
	if cliRoleLoaded {
		return cliRole, cliRoleErr
	}
	current, err := currentUserFn()
	if err != nil {
		cliRoleErr = fmt.Errorf("could not determine current user: %w", err)
		cliRoleLoaded = true
		return 0, cliRoleErr
	}
	cliRole, cliRoleErr = ResolveRole(current.Username)
	cliRoleLoaded = true
	return cliRole, cliRoleErr
}

// RequireCLIRole checks the minimum role for a command.
func RequireCLIRole(minRole Role) error {
	role, err := CLIRole()
	if err != nil {
		return err
	}
	if role < minRole {
		return fmt.Errorf(
			"this command requires role %s or higher — you have role %s\nAsk a system administrator to add you to the %s group",
			minRole, role, RoleGroup(minRole),
		)
	}
	return nil
}

// ResetCLIAuth clears cached CLI role state for tests.
func ResetCLIAuth() {
	cliRoleMu.Lock()
	defer cliRoleMu.Unlock()
	cliRole = 0
	cliRoleErr = nil
	cliRoleLoaded = false
}

// SetCLIRoleForTesting forces a cached CLI role and returns a restore function.
func SetCLIRoleForTesting(role Role) func() {
	cliRoleMu.Lock()
	oldRole := cliRole
	oldErr := cliRoleErr
	oldLoaded := cliRoleLoaded
	cliRole = role
	cliRoleErr = nil
	cliRoleLoaded = true
	cliRoleMu.Unlock()

	return func() {
		cliRoleMu.Lock()
		defer cliRoleMu.Unlock()
		cliRole = oldRole
		cliRoleErr = oldErr
		cliRoleLoaded = oldLoaded
	}
}
