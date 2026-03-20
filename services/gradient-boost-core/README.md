# gradient-boost-core

## Purpose

Long-running Python environment for CPU-first experimentation and command execution.

## Image

- `python:3.12-slim`

## Ports

- None exposed to the host

## Mounts

- `~/gradient/data -> /data`
- `~/gradient/notebooks -> /notebooks`
- `~/gradient/models -> /models`
- `~/gradient/outputs -> /outputs`
- `~/gradient/mlruns -> /mlruns`

## Dependencies

- Workspace must exist
- Compose network must be present

## Lifecycle Notes

Install pulls the image and renders it into the Boosting Compose file. Remove must never touch user data directories.
