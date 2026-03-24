# System Administration

This guide covers local roles, services, workspace paths, and the current machine-level command surface.

## Roles

`concave` derives access from Unix group membership:

| Group | Role | Typical access |
|---|---|---|
| `gradient-viewer` | Viewer | status, logs, check, workspace status, environment status, fleet status |
| `gradient-developer` | Developer | Viewer access plus lab, shell, exec |
| `gradient-operator` | Operator | Developer access plus install, remove, start, stop, update, rollback, backup, prune |
| `gradient-admin` | Admin | Operator access plus serve, GPU setup, setup, upgrade, resolver and mesh restarts |

Use:

```bash
concave whoami
```

to confirm the current user, role, and allowed command set.

## Workspace paths

The default workspace root is:

```text
~/gradient/
```

Important subdirectories:

- `data/`
- `notebooks/`
- `models/`
- `outputs/`
- `mlruns/`
- `dags/`
- `compose/`
- `config/`
- `backups/`
- `logs/`

Important config files:

- `~/gradient/config/state.json`
- `~/gradient/config/versions.json`
- `~/gradient/config/setup.json`

## Local API

`concave serve` starts the authenticated local API server. The default bind address is:

```text
127.0.0.1:7777
```

Start it with:

```bash
concave serve
```

The TUI and web clients authenticate against this server. CLI role checks do not use the API session cookie path.

## Current service commands

The current command names are:

- `concave check`
- `concave gpu setup`
- `concave gpu check`
- `concave gpu info`
- `concave workspace prune`
- `concave upgrade`

Environment, fleet, node, team, resolver, and mesh commands are also present:

- `concave env status`
- `concave env diff`
- `concave node status`
- `concave fleet status`
- `concave team list`
- `concave resolver status`
- `concave mesh status`

Deprecated aliases still work and print warnings:

- `concave doctor`
- `concave driver-wizard`
- `concave self-update`
- `concave workspace clean`

## Systemd services

The current stack includes these service names:

- `concave-serve.service`
- `gradient-resolver.service`
- `gradient-mesh.service`
- `gradient-lab.service`

`concave-serve.service` backs the local API. The other services are companion components and may only be installed on systems that enable those layers.

## Resolver, mesh, and lab

Resolver and mesh are local daemons that report state back into `concave`:

- Resolver socket: `/run/gradient/resolver.sock`
- Mesh socket: `/run/gradient/mesh.sock`

Gradient Lab is the notebook-facing layer built on top of JupyterHub and can be deployed separately from the core CLI.

## Administrative actions

Viewer-safe machine inspection:

```bash
concave check
concave status
concave env status
concave fleet status
```

Operator and admin actions:

```bash
concave install boosting
concave workspace backup
concave gpu setup
concave upgrade
```

## Notes

- `concave` keeps workspace state local to the machine.
- TUI and web clients should read machine state from the local API instead of re-deriving it.
- `gradient-resolver.service`, `gradient-mesh.service`, and `gradient-lab.service` extend the stack, but `concave` remains the host control plane.
