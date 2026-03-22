# concave serve API

`concave serve` is the authenticated local control plane for the Gradient Linux
stack. It exposes HTTP and WebSocket endpoints for browser and terminal clients
without turning the CLI into a long-running UI process.

## Purpose

The server exists to provide:

- authenticated login using the host's Unix user database and PAM
- role-aware access to suite operations and machine controls
- a stable HTTP API for `concave-web`
- a sessioned backend for `concave-tui`
- long-running job polling for mutating operations
- container and host terminal access over WebSocket

The CLI remains authoritative for local shell usage. The server is the network
surface for remote or multi-client control.

## Authentication model

Authentication is handled by `internal/auth/`.

- Users authenticate with username and password through PAM.
- Authorization is derived from Unix group membership.
- Supported groups:
  - `gradient-viewer`
  - `gradient-developer`
  - `gradient-operator`
  - `gradient-admin`
- If a user is in multiple `gradient-*` groups, the highest role wins.
- If a user is in none of them, login is rejected.

Sessions are JWT-backed.

- Browser clients use the HTTP cookie path.
- TUI clients request the token in the JSON login response by sending
  `X-Concave-Client: tui`.
- Session persistence on the TUI side is stored in
  `~/.config/concave/session.json`.

## Service model

The packaged deployment runs:

- `concave serve`
- under `concave-serve.service`
- as the `gradient-svc` system user
- with configuration from `/etc/default/concave-serve`

The packaged postinstall also creates the auth groups, writes the service
environment file, provisions the sudoers rule for tightly-scoped privileged
operations, and prepares `/var/lib/gradient/`.

See [system-admin.md](system-admin.md) for the deployment details.

## Route groups

The server mounts endpoints under `/api/v1/`.

### Auth

- `POST /api/v1/auth/login`
- `POST /api/v1/auth/logout`
- `POST /api/v1/auth/refresh`
- `GET /api/v1/auth/me`

These routes handle login, logout, refresh, and identity lookup.

### Health and jobs

- `GET /api/v1/health`
- `GET /api/v1/jobs/{id}`
- `GET /api/v1/metrics/stream`

Mutating operations return a job identifier. Clients poll the job endpoint to
stream progress text and final state.

### Read-only operational endpoints

- `GET /api/v1/doctor`
- `GET /api/v1/workspace`
- `GET /api/v1/system/info`
- `GET /api/v1/system/users`
- `GET /api/v1/users/activity`
- `GET /api/v1/suites`
- `GET /api/v1/suites/{suite}`
- `GET /api/v1/suites/{suite}/changelog`
- `GET /api/v1/suites/{suite}/lab`

These endpoints back status dashboards and read-only views in the web and TUI
clients.

### Mutating operational endpoints

- `POST /api/v1/workspace/backup`
- `POST /api/v1/workspace/clean`
- `POST /api/v1/system/reboot`
- `POST /api/v1/system/shutdown`
- `POST /api/v1/system/restart-docker`
- `POST /api/v1/suites/{suite}/install`
- `POST /api/v1/suites/forge/install`
- `POST /api/v1/suites/{suite}/remove`
- `POST /api/v1/suites/{suite}/start`
- `POST /api/v1/suites/{suite}/stop`
- `POST /api/v1/suites/{suite}/update`
- `POST /api/v1/suites/{suite}/rollback`

These routes enqueue or perform privileged lifecycle actions after role checks.

## Role boundaries

The permission model is shared between CLI enforcement and server middleware.

- Viewer:
  - doctor, status, logs, metrics, workspace status
- Developer:
  - Viewer permissions
  - lab, shell, exec, container-terminal-level operations
- Operator:
  - Developer permissions
  - install, remove, start, stop, update, rollback, backup, clean
- Admin:
  - Operator permissions
  - reboot, shutdown, restart Docker, user-management views, host terminal

The role mapping is documented in `internal/auth/permissions.go`.

## Job model

Long-running operations use a lightweight in-memory job manager.

Each job provides:

- a unique identifier
- current status
- progress text lines
- completion state
- final error message when applicable

This lets the TUI and browser remain responsive while the server performs Docker,
workspace, or suite operations.

## Terminal endpoints

Two WebSocket endpoints are exposed:

- `GET /api/v1/terminal/container/{suite}/{container}`
- `GET /api/v1/terminal/host`

Container terminals are for Developer and above. Host terminals are Admin only.
The host shell is launched through the packaged helper path so the session runs as
the authenticated Unix user rather than as `gradient-svc`.

## Operational notes

- The API is intended for local or proxied use, not direct public exposure.
- `concave-web` should front it for browser use.
- `concave-tui` should use session caching and `auth/me` instead of embedding its
  own role logic.
- Logs, lab tokens, workspace status, and suite state should be read from the API
  rather than re-derived independently in clients.
