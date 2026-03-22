# Boosting Suite

## Purpose

Boosting is the first end-to-end `concave` suite. It is CPU-first, works on machines
without a GPU, and proves the install, start, status, logs, shell, exec, and `lab`
flows without any driver prerequisites.

## Containers

- `gradient-boost-core`
- `gradient-boost-lab`
- `gradient-boost-track`

## Host Ports

- `8888` → JupyterLab
- `5000` → MLflow

## Workspace Mounts

- `~/gradient/data` → `/data`
- `~/gradient/notebooks` → `/notebooks`
- `~/gradient/models` → `/models`
- `~/gradient/outputs` → `/outputs`
- `~/gradient/mlruns` → `/mlruns`

## Install and Lifecycle

- `concave install boosting` initializes the workspace, renders
  `~/gradient/compose/boosting.compose.yml`, validates it, pulls images, and records the
  suite in `state.json` and `versions.json`.
- `concave start boosting` uses the rendered Compose file to start the suite.
- `concave lab` resolves the tokenized Jupyter URL from `gradient-boost-lab` and opens
  it in the browser.
- `concave remove boosting` removes Compose-managed resources and config entries only.
- `concave rollback boosting` swaps image metadata and restarts containers without
  touching datasets, notebooks, or models.

## Service Reference

### `gradient-boost-core`

- Role: long-running Python runtime for ad hoc commands, notebooks support tasks, and
  CPU-bound experimentation
- Image: `python:3.12-slim`
- Ports: none
- Mounts: data, notebooks, models, outputs, mlruns
- Startup path: starts as a durable helper container with `sleep infinity`
- Dependencies: workspace exists, Docker Engine available, suite Compose file valid
- Health and logs: should remain in a running state with little output under normal use
- Troubleshooting: if the container exits immediately, inspect command overrides and
  the rendered Compose file
- Update and rollback behavior: image tag history is tracked in `versions.json`

### `gradient-boost-lab`

- Role: JupyterLab frontend for notebook-centric workflows
- Image: `quay.io/jupyter/base-notebook:python-3.11.6`
- Ports: `8888`
- Mounts: data, notebooks, models, outputs, mlruns
- Startup path: launched by Compose, then discovered by `concave lab`
- Dependencies: browser opener on the host, healthy Docker container state
- Health and logs: startup logs should expose the Jupyter URL or token
- Troubleshooting: if `concave lab` cannot find the tokenized URL, inspect container
  logs and verify port `8888` is not already claimed
- Update and rollback behavior: rollback restores the previously tracked image tag and
  reuses the existing workspace content

### `gradient-boost-track`

- Role: MLflow tracking UI and metadata backend
- Image: `ghcr.io/mlflow/mlflow:v2.14.1`
- Ports: `5000`
- Mounts: `mlruns`, `outputs`
- Startup path: starts MLflow in UI mode against the workspace-backed store
- Dependencies: writable `~/gradient/mlruns`, free host port `5000`
- Health and logs: should bind successfully and expose the UI on port `5000`
- Troubleshooting: if startup fails, inspect permissions on `~/gradient/mlruns` and
  check for host port conflicts
- Update and rollback behavior: image rollback does not remove MLflow data

## Failure Modes

- Docker unavailable
- user not in the Docker group
- port `8888` or `5000` already in use
- missing workspace directories
- invalid rendered Compose output
