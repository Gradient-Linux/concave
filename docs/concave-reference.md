# concave Reference

## Binaries

- `concave`: the Cobra-based CLI
- `concave-tui`: the Bubble Tea terminal interface with parity for dashboard, suite lifecycle, logs, workspace, and doctor flows

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

## TUI Views

- `Dashboard`: high-level suite health, GPU summary, workspace free space
- `Suites`: install, remove, update, rollback, start, stop, restart, shell, exec, lab
- `Logs`: live per-container log following with search and bounded history
- `Workspace`: usage, backup, and outputs cleanup
- `Doctor`: asynchronous system and suite health checks

## Setup Commands

- `concave setup`
- `concave driver-wizard`
- `concave self-update`
