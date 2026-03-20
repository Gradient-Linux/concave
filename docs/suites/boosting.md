# Boosting Suite

## Purpose

Boosting is the first end-to-end suite for `concave`. It works on CPU-only systems and proves the full install, start, and lab flow without requiring GPU support.

## Services

- `gradient-boost-core`
- `gradient-boost-lab`
- `gradient-boost-track`

## Ports

- `8888` JupyterLab
- `5000` MLflow

## Volumes

- `/data`
- `/notebooks`
- `/models`
- `/outputs`
- `/mlruns`

## Failure Modes

- Docker unavailable
- Port 8888 or 5000 already in use
- Workspace missing required subdirectories
