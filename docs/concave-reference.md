# concave Reference

## Binaries

- `concave`: the Cobra-based CLI

## Global Flags

- `-v`, `--verbose`: emit structured debug logs to stderr

## Exit Codes

- `0`: success
- `1`: user error, bad arguments, cancelled operation, or lock contention
- `2`: Docker or runtime failure
- `3`: panic; see `~/gradient/logs/concave.log`
- `130`: interrupted by `Ctrl+C`
- `143`: terminated by `SIGTERM`

## Core Commands

- `concave check`
- `concave gpu setup`
- `concave gpu check`
- `concave gpu info`
- `concave workspace init`
- `concave workspace status`
- `concave workspace backup`
- `concave workspace prune --outputs`

## Suite Commands

- `concave install [suite]`
- `concave remove [suite]`
- `concave update [suite]`
- `concave rollback [suite]`
- `concave changelog [suite]`
- `concave list`
- `concave start [suite?]`
- `concave stop [suite?]`
- `concave restart [suite]`
- `concave status`
- `concave logs [suite]`
- `concave lab`
- `concave shell [suite]`
- `concave exec [suite] -- [command]`

## Setup Commands

- `concave setup`
- `concave upgrade`
- `concave completion [bash|zsh|fish]` (hidden; used by build and packaging tooling)

## Environment Commands

- `concave env status [--group <name>]`
- `concave env diff [--group <name>]`
- `concave env export --group <name> [--layers python]`
- `concave env apply --group <name> [--backend cuda|rocm|cpu]`
- `concave env rollback --package <pkg> --group <name>`
- `concave env baseline set --group <name>`
- `concave env baseline show --group <name>`

## Fleet Commands

- `concave node status`
- `concave node set --visibility [public|private|hidden]`
- `concave fleet status`
- `concave fleet peers`
- `concave resolver status`
- `concave resolver logs [--follow]`
- `concave resolver restart`
- `concave mesh status`
- `concave mesh logs [--follow]`
- `concave mesh restart`

## Team Commands

- `concave team create --name <name> --preset <preset>`
- `concave team list`
- `concave team status [name]`
- `concave team add-user --group <name> --user <username>`
- `concave team remove-user --group <name> --user <username>`
- `concave team delete --name <name>`

## Examples

```bash
concave check
concave install boosting
concave start boosting
concave logs boosting --service gradient-boost-lab --follow
concave rollback boosting
concave --verbose update neural
```
