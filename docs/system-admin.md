# System Administration

`concave` Phase 3 adds a local authenticated control plane alongside the CLI.

## Roles

Access is derived from Unix group membership. There is no separate app user database.

- `gradient-viewer`: read-only status, logs, metrics, check, workspace views
- `gradient-developer`: viewer access plus lab, shell, and exec actions
- `gradient-operator`: developer access plus install, remove, start, stop, update, rollback, backup, and prune
- `gradient-admin`: operator access plus `concave serve`, restart-docker, reboot, shutdown, gpu setup, and upgrade

If a user belongs to multiple `gradient-*` groups, the highest role wins at runtime.
That is treated as a host misconfiguration and should be corrected by an admin.

## CLI Auth

The CLI resolves the current Unix user role directly from their `gradient-*` group
membership on each invocation.

Diagnostic commands remain ungated:

- `concave`
- `concave --help`
- `concave --version`
- `concave check`
- `concave whoami`

Use `concave whoami` to confirm the current user, groups, role, and allowed command set.

## concave serve

`concave serve` runs the authenticated local API server, by default on `127.0.0.1:7777`.

It provides:

- `/api/v1/auth/*` login, logout, refresh, and identity endpoints
- authenticated suite, workspace, check, users, and system endpoints
- WebSocket endpoints for container and host terminals

JWT signing material is stored in the auth config under the configured workspace root.

## Packaged Service

The Debian package installs:

- the `gradient-svc` system user
- the four `gradient-*` role groups
- `concave-serve.service`
- `/etc/default/concave-serve`
- a locked-down sudoers rule for `gradient-svc`

The packaged service is intended to run as `gradient-svc`, not as a login shell user.

## Notes

- TUI and web clients authenticate against `concave serve`
- CLI role enforcement does not use JWTs
- host-level control actions are Admin-only
