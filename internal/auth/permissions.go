package auth

import "fmt"

// Action represents an operation a user might attempt.
type Action string

const (
	ActionViewStatus    Action = "view:status"
	ActionViewLogs      Action = "view:logs"
	ActionViewMetrics   Action = "view:metrics"
	ActionViewCheck     Action = "view:check"
	ActionViewWorkspace Action = "view:workspace"
	ActionViewFleet     Action = "view:fleet"
	ActionViewEnv       Action = "view:env"

	ActionOpenLab  Action = "suite:lab"
	ActionShell    Action = "suite:shell"
	ActionExec     Action = "suite:exec"
	ActionGPUCheck Action = "gpu:check"
	ActionGPUInfo  Action = "gpu:info"

	ActionInstall       Action = "suite:install"
	ActionRemove        Action = "suite:remove"
	ActionStart         Action = "suite:start"
	ActionStop          Action = "suite:stop"
	ActionUpdate        Action = "suite:update"
	ActionRollback      Action = "suite:rollback"
	ActionBackup        Action = "workspace:backup"
	ActionPrune         Action = "workspace:prune"
	ActionTeamList      Action = "team:list"
	ActionTeamStatus    Action = "team:status"
	ActionTeamManage    Action = "team:manage"
	ActionNodeStatus    Action = "node:status"
	ActionNodeSet       Action = "node:set"
	ActionEnvStatus     Action = "env:status"
	ActionEnvDiff       Action = "env:diff"
	ActionEnvApply      Action = "env:apply"
	ActionEnvExport     Action = "env:export"
	ActionEnvRollback   Action = "env:rollback"
	ActionEnvBaseline   Action = "env:baseline"
	ActionResolverState Action = "resolver:status"
	ActionMeshState     Action = "mesh:status"

	ActionReboot          Action = "system:reboot"
	ActionShutdown        Action = "system:shutdown"
	ActionRestartDocker   Action = "system:restart-docker"
	ActionManageUsers     Action = "system:manage-users"
	ActionGPUSetup        Action = "system:gpu-setup"
	ActionUpgrade         Action = "system:upgrade"
	ActionResolverRestart Action = "system:resolver-restart"
	ActionMeshRestart     Action = "system:mesh-restart"
)

var rolePermissions = map[Role][]Action{
	RoleViewer: {
		ActionViewStatus, ActionViewLogs, ActionViewMetrics, ActionViewCheck, ActionViewWorkspace, ActionViewFleet, ActionViewEnv, ActionGPUCheck, ActionGPUInfo, ActionTeamList, ActionTeamStatus, ActionNodeStatus, ActionEnvStatus, ActionEnvDiff, ActionEnvBaseline, ActionResolverState, ActionMeshState,
	},
	RoleDeveloper: {
		ActionOpenLab, ActionShell, ActionExec,
	},
	RoleOperator: {
		ActionInstall, ActionRemove, ActionStart, ActionStop, ActionUpdate, ActionRollback, ActionBackup, ActionPrune, ActionTeamManage, ActionNodeSet, ActionEnvApply, ActionEnvExport, ActionEnvRollback, ActionEnvBaseline,
	},
	RoleAdmin: {
		ActionReboot, ActionShutdown, ActionRestartDocker, ActionManageUsers, ActionGPUSetup, ActionUpgrade, ActionResolverRestart, ActionMeshRestart,
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
