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

- `concave doctor`
- `concave workspace init`
- `concave workspace status`
- `concave workspace backup`
- `concave workspace clean --outputs`

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
- `concave driver-wizard`
- `concave self-update`
- `concave completion [bash|zsh|fish]` (hidden; used by build and packaging tooling)

## Examples

```bash
concave doctor
concave install boosting
concave start boosting
concave logs boosting --service gradient-boost-lab --follow
concave rollback boosting
concave --verbose update neural
```
