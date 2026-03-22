package auth

import "fmt"

// Action represents an operation a user might attempt.
type Action string

const (
	ActionViewStatus    Action = "view:status"
	ActionViewLogs      Action = "view:logs"
	ActionViewMetrics   Action = "view:metrics"
	ActionViewDoctor    Action = "view:doctor"
	ActionViewWorkspace Action = "view:workspace"

	ActionOpenLab Action = "suite:lab"
	ActionShell   Action = "suite:shell"
	ActionExec    Action = "suite:exec"

	ActionInstall  Action = "suite:install"
	ActionRemove   Action = "suite:remove"
	ActionStart    Action = "suite:start"
	ActionStop     Action = "suite:stop"
	ActionUpdate   Action = "suite:update"
	ActionRollback Action = "suite:rollback"
	ActionBackup   Action = "workspace:backup"
	ActionClean    Action = "workspace:clean"

	ActionReboot        Action = "system:reboot"
	ActionShutdown      Action = "system:shutdown"
	ActionRestartDocker Action = "system:restart-docker"
	ActionManageUsers   Action = "system:manage-users"
	ActionDriverWizard  Action = "system:driver-wizard"
	ActionSelfUpdate    Action = "system:self-update"
)

var rolePermissions = map[Role][]Action{
	RoleViewer: {
		ActionViewStatus, ActionViewLogs, ActionViewMetrics, ActionViewDoctor, ActionViewWorkspace,
	},
	RoleDeveloper: {
		ActionOpenLab, ActionShell, ActionExec,
	},
	RoleOperator: {
		ActionInstall, ActionRemove, ActionStart, ActionStop, ActionUpdate, ActionRollback, ActionBackup, ActionClean,
	},
	RoleAdmin: {
		ActionReboot, ActionShutdown, ActionRestartDocker, ActionManageUsers, ActionDriverWizard, ActionSelfUpdate,
	},
}

// Can returns true if a role is permitted to perform an action.
func Can(role Role, action Action) bool {
	for current := RoleViewer; current <= role; current++ {
		for _, allowed := range rolePermissions[current] {
			if allowed == action {
				return true
			}
		}
	}
	return false
}

// Require returns an error safe to expose to clients and UIs.
func Require(role Role, action Action) error {
	if Can(role, action) {
		return nil
	}
	return fmt.Errorf("role %s cannot perform %s", role, action)
}

// AllowedActions returns the flattened permission set for a role.
func AllowedActions(role Role) []Action {
	seen := map[Action]struct{}{}
	actions := make([]Action, 0, len(rolePermissions)*4)
	for current := RoleViewer; current <= role; current++ {
		for _, action := range rolePermissions[current] {
			if _, ok := seen[action]; ok {
				continue
			}
			seen[action] = struct{}{}
			actions = append(actions, action)
		}
	}
	return actions
}
