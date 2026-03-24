# concave CLI Reference

`concave` manages the Gradient Linux workspace, suite lifecycle, GPU setup, and the local authenticated API.

## Global flags

| Flag | Default | Description |
|---|---|---|
| `-v`, `--verbose` | `false` | Enable verbose debug output to stderr |
| `--version` | `false` | Print the current `concave` version |
| `-h`, `--help` | `false` | Show help for the active command |

## Global exit codes

| Code | Meaning |
|---|---|
| `0` | Success |
| `1` | User error, invalid arguments, cancelled operation, or permission failure |
| `2` | Docker or runtime failure |
| `3` | Panic; inspect `~/gradient/logs/concave.log` |
| `130` | Interrupted by `Ctrl+C` |
| `143` | Terminated by `SIGTERM` |

## Deprecated aliases

These commands still work, print a deprecation warning, and then run the replacement command:

| Old command | Replacement |
|---|---|
| `concave doctor` | `concave check` |
| `concave driver-wizard` | `concave gpu setup` |
| `concave self-update` | `concave upgrade` |
| `concave workspace clean` | `concave workspace prune` |

## Infrastructure

### concave install

Install a suite and write its Compose file and state.

**Usage:** `concave install [suite] [flags]`

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `--force` | `false` | Reinstall an already-installed suite |

**Exit codes:** Standard exit codes only.

**Example:**

```bash
concave install boosting
```

### concave remove

Remove a suite while preserving user data in the Gradient workspace.

**Usage:** `concave remove [suite] [flags]`

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `—` | `—` | No command-specific flags |

**Exit codes:** Standard exit codes only.

**Example:**

```bash
concave remove boosting
```

### concave start

Start one suite or every installed suite.

**Usage:** `concave start [suite] [flags]`

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `—` | `—` | No command-specific flags |

**Exit codes:** Standard exit codes only.

**Example:**

```bash
concave start boosting
```

### concave stop

Stop one suite or every installed suite.

**Usage:** `concave stop [suite] [flags]`

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `—` | `—` | No command-specific flags |

**Exit codes:** Standard exit codes only.

**Example:**

```bash
concave stop boosting
```

### concave restart

Restart one suite.

**Usage:** `concave restart [suite] [flags]`

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `—` | `—` | No command-specific flags |

**Exit codes:** Standard exit codes only.

**Example:**

```bash
concave restart boosting
```

### concave status

Show container status, suite ports, GPU summary, and workspace free space.

**Usage:** `concave status [flags]`

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `—` | `—` | No command-specific flags |

**Exit codes:** Standard exit codes only.

**Example:**

```bash
concave status
```

### concave list

List every built-in suite and the current recorded image for installed suites.

**Usage:** `concave list [flags]`

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `—` | `—` | No command-specific flags |

**Exit codes:** Standard exit codes only.

**Example:**

```bash
concave list
```

### concave logs

Tail logs from one suite or one service within a suite.

**Usage:** `concave logs [suite] [flags]`

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `--follow` | `true` | Follow log output |
| `--lines` | `50` | Number of historical log lines to show |
| `--service` | `""` | Tail one named Compose service |

**Exit codes:** Standard exit codes only.

**Example:**

```bash
concave logs boosting --service gradient-boost-lab --lines 100
```

### concave changelog

Show the local image delta between recorded and target suite images.

**Usage:** `concave changelog [suite] [flags]`

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `—` | `—` | No command-specific flags |

**Exit codes:** Standard exit codes only.

**Example:**

```bash
concave changelog boosting
```

### concave lab

Resolve a tokenized JupyterLab URL for `boosting` or `neural` and open it in the browser.

**Usage:** `concave lab [flags]`

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `--suite` | `""` | Open a specific suite instead of auto-selecting the first installed lab-capable suite |

**Exit codes:** Standard exit codes only.

**Example:**

```bash
concave lab --suite neural
```

### concave shell

Open an interactive shell in the primary container for a suite.

**Usage:** `concave shell [suite] [flags]`

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `—` | `—` | No command-specific flags |

**Exit codes:** Standard exit codes only.

**Example:**

```bash
concave shell boosting
```

### concave exec

Run a non-interactive command inside the primary container for a suite.

**Usage:** `concave exec [suite] -- [command] [flags]`

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `—` | `—` | No command-specific flags |

**Exit codes:** Standard exit codes plus the exit status returned by the container command when `docker exec` fails inside the container.

**Example:**

```bash
concave exec boosting -- python --version
```

## GPU

### concave gpu setup

Run the interactive GPU setup flow.

**Usage:** `concave gpu setup [flags]`

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `—` | `—` | No command-specific flags in the current build |

**Exit codes:** Standard exit codes only.

**Example:**

```bash
concave gpu setup
```

### concave gpu check

Run GPU-specific health checks.

**Usage:** `concave gpu check [flags]`

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `—` | `—` | No command-specific flags |

**Exit codes:** Standard exit codes only.

**Example:**

```bash
concave gpu check
```

### concave gpu info

Show detected GPU devices, compute capability, and recommended driver branch when available.

**Usage:** `concave gpu info [flags]`

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `—` | `—` | No command-specific flags |

**Exit codes:** Standard exit codes only.

**Example:**

```bash
concave gpu info
```

## Workspace

### concave workspace init

Create the fixed Gradient workspace tree.

**Usage:** `concave workspace init [flags]`

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `—` | `—` | No command-specific flags |

**Exit codes:** Standard exit codes only.

**Example:**

```bash
concave workspace init
```

### concave workspace status

Show workspace disk usage by managed subdirectory.

**Usage:** `concave workspace status [flags]`

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `—` | `—` | No command-specific flags |

**Exit codes:** Standard exit codes only.

**Example:**

```bash
concave workspace status
```

### concave workspace backup

Create a timestamped tarball of notebooks and models.

**Usage:** `concave workspace backup [flags]`

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `—` | `—` | No command-specific flags |

**Exit codes:** Standard exit codes only.

**Example:**

```bash
concave workspace backup
```

### concave workspace prune

Remove generated files from `~/gradient/outputs`.

**Usage:** `concave workspace prune [flags]`

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `--outputs` | `false` | Prune the contents of `~/gradient/outputs` |

**Exit codes:** Standard exit codes only.

**Example:**

```bash
concave workspace prune --outputs
```

## Environment

### concave env status

Show resolver status and group drift summaries when the resolver daemon is available.

**Usage:** `concave env status [flags]`

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `--group` | `""` | Limit output to one group |

**Exit codes:** Standard exit codes only.

**Example:**

```bash
concave env status --group research
```

### concave env diff

Show package drift details from the resolver daemon.

**Usage:** `concave env diff [flags]`

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `--group` | `""` | Limit output to one group |

**Exit codes:** Standard exit codes only.

**Example:**

```bash
concave env diff --group research
```

### concave env export

Preview command for exporting a portable environment snapshot. The current build prints a preview message and exits successfully.

**Usage:** `concave env export [flags]`

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `--group` | `""` | Group name |
| `--layers` | `"python"` | Environment layers to export |

**Exit codes:** `0` on preview output; standard error codes if argument parsing fails.

**Example:**

```bash
concave env export --group research --layers python
```

### concave env apply

Preview command for applying an environment snapshot to a target backend. The current build prints a preview message and exits successfully.

**Usage:** `concave env apply [flags]`

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `--group` | `""` | Group name |
| `--backend` | `"cuda"` | Target backend: `cuda`, `rocm`, or `cpu` |

**Exit codes:** `0` on preview output; standard error codes if argument parsing fails.

**Example:**

```bash
concave env apply --group research --backend rocm
```

### concave env rollback

Preview command for package-level rollback. The current build prints a preview message and exits successfully.

**Usage:** `concave env rollback [flags]`

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `--group` | `""` | Group name |
| `--package` | `""` | Package name |

**Exit codes:** `0` on preview output; standard error codes if argument parsing fails.

**Example:**

```bash
concave env rollback --group research --package pandas
```

### concave env baseline set

Preview command for setting a resolver baseline. The current build prints a preview message and exits successfully.

**Usage:** `concave env baseline set [flags]`

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `--group` | `""` | Group name |

**Exit codes:** `0` on preview output; standard error codes if argument parsing fails.

**Example:**

```bash
concave env baseline set --group research
```

### concave env baseline show

Preview command for showing a resolver baseline. The current build prints a preview message and exits successfully.

**Usage:** `concave env baseline show [flags]`

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `--group` | `""` | Group name |

**Exit codes:** `0` on preview output; standard error codes if argument parsing fails.

**Example:**

```bash
concave env baseline show --group research
```

## Fleet and teams

### concave node status

Show the local mesh node status when the mesh daemon is available.

**Usage:** `concave node status [flags]`

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `—` | `—` | No command-specific flags |

**Exit codes:** Standard exit codes only.

**Example:**

```bash
concave node status
```

### concave node set

Preview command for changing node visibility. The current build prints a preview message and exits successfully.

**Usage:** `concave node set [flags]`

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `--visibility` | `"public"` | Node visibility: `public`, `private`, or `hidden` |

**Exit codes:** `0` on preview output; standard error codes if argument parsing fails.

**Example:**

```bash
concave node set --visibility hidden
```

### concave fleet status

Show visible mesh peers when the mesh daemon is available.

**Usage:** `concave fleet status [flags]`

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `—` | `—` | No command-specific flags |

**Exit codes:** Standard exit codes only.

**Example:**

```bash
concave fleet status
```

### concave fleet peers

Show the current peer snapshot from the mesh daemon.

**Usage:** `concave fleet peers [flags]`

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `—` | `—` | No command-specific flags |

**Exit codes:** Standard exit codes only.

**Example:**

```bash
concave fleet peers
```

### concave team create

Preview command for creating a team. The current build prints a preview message and exits successfully.

**Usage:** `concave team create [flags]`

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `--name` | `""` | Team name |
| `--preset` | `""` | Team preset |

**Exit codes:** `0` on preview output; standard error codes if argument parsing fails.

**Example:**

```bash
concave team create --name research --preset research-team
```

### concave team list

Preview command for listing teams. The current build prints a preview message and exits successfully.

**Usage:** `concave team list [flags]`

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `—` | `—` | No command-specific flags |

**Exit codes:** `0` on preview output.

**Example:**

```bash
concave team list
```

### concave team status

Preview command for showing one team or group status. The current build prints a preview message and exits successfully.

**Usage:** `concave team status [name] [flags]`

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `—` | `—` | No command-specific flags |

**Exit codes:** `0` on preview output.

**Example:**

```bash
concave team status research
```

### concave team add-user

Preview command for adding a user to a team. The current build prints a preview message and exits successfully.

**Usage:** `concave team add-user [flags]`

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `--group` | `""` | Team group |
| `--user` | `""` | Username |

**Exit codes:** `0` on preview output; standard error codes if argument parsing fails.

**Example:**

```bash
concave team add-user --group research --user alice
```

### concave team remove-user

Preview command for removing a user from a team. The current build prints a preview message and exits successfully.

**Usage:** `concave team remove-user [flags]`

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `--group` | `""` | Team group |
| `--user` | `""` | Username |

**Exit codes:** `0` on preview output; standard error codes if argument parsing fails.

**Example:**

```bash
concave team remove-user --group research --user alice
```

### concave team delete

Preview command for deleting a team. The current build prints a preview message and exits successfully.

**Usage:** `concave team delete [flags]`

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `--name` | `""` | Team name |

**Exit codes:** `0` on preview output; standard error codes if argument parsing fails.

**Example:**

```bash
concave team delete --name research
```

### concave resolver status

Show resolver daemon status and drift summary when the resolver socket is available.

**Usage:** `concave resolver status [flags]`

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `—` | `—` | No command-specific flags |

**Exit codes:** Standard exit codes only.

**Example:**

```bash
concave resolver status
```

### concave resolver logs

Preview command for resolver logs. The current build prints a preview message and exits successfully.

**Usage:** `concave resolver logs [flags]`

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `--follow` | `false` | Follow resolver logs |

**Exit codes:** `0` on preview output.

**Example:**

```bash
concave resolver logs --follow
```

### concave resolver restart

Preview command for restarting the resolver service. The current build prints a preview message and exits successfully.

**Usage:** `concave resolver restart [flags]`

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `—` | `—` | No command-specific flags |

**Exit codes:** `0` on preview output.

**Example:**

```bash
concave resolver restart
```

### concave mesh status

Show mesh node status using the same local data as `concave node status`.

**Usage:** `concave mesh status [flags]`

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `—` | `—` | No command-specific flags |

**Exit codes:** Standard exit codes only.

**Example:**

```bash
concave mesh status
```

### concave mesh logs

Preview command for mesh logs. The current build prints a preview message and exits successfully.

**Usage:** `concave mesh logs [flags]`

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `--follow` | `false` | Follow mesh logs |

**Exit codes:** `0` on preview output.

**Example:**

```bash
concave mesh logs --follow
```

### concave mesh restart

Preview command for restarting the mesh service. The current build prints a preview message and exits successfully.

**Usage:** `concave mesh restart [flags]`

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `—` | `—` | No command-specific flags |

**Exit codes:** `0` on preview output.

**Example:**

```bash
concave mesh restart
```

## Developer tools

### concave serve

Run the authenticated local control-plane API.

**Usage:** `concave serve [flags]`

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `--addr` | `"127.0.0.1:7777"` | Bind address for the API server |

**Exit codes:** Standard exit codes only.

**Example:**

```bash
concave serve --addr 127.0.0.1:7777
```

### concave whoami

Show the current Unix user, resolved Gradient role, and allowed commands.

**Usage:** `concave whoami [flags]`

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `—` | `—` | No command-specific flags |

**Exit codes:** Standard exit codes only. The command still exits successfully when the current user has no Gradient Linux role.

**Example:**

```bash
concave whoami
```

### concave completion

Generate shell completion scripts. This command is hidden from the normal help output.

**Usage:** `concave completion [bash|zsh|fish]`

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `—` | `—` | No command-specific flags |

**Exit codes:** Standard exit codes only.

**Example:**

```bash
concave completion zsh
```

## System

### concave check

Run the full local health check.

**Usage:** `concave check [flags]`

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `—` | `—` | No command-specific flags |

**Exit codes:** Standard exit codes only.

**Example:**

```bash
concave check
```

### concave setup

Run the first-boot setup wizard.

**Usage:** `concave setup [flags]`

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `—` | `—` | No command-specific flags |

**Exit codes:** Standard exit codes only.

**Example:**

```bash
concave setup
```

### concave upgrade

Download a replacement binary, verify its SHA256, and atomically replace the current executable.

**Usage:** `concave upgrade [flags]`

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `—` | `—` | No command-specific flags |

**Exit codes:** Standard exit codes only.

**Example:**

```bash
concave upgrade
```
