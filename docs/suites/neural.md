# Neural Suite

## Purpose

Neural provides GPU-oriented training, inference, and notebook workflows for deep
learning workloads. Unlike Boosting, Neural depends on NVIDIA-capable Docker runtime
support for the full execution path.

## Containers

- `gradient-neural-torch`
- `gradient-neural-infer`
- `gradient-neural-lab`

## Host Ports

- `8000` → inference API
- `8080` → auxiliary HTTP endpoint
- `8888` → JupyterLab

## Workspace Mounts

- `~/gradient/data` → `/data`
- `~/gradient/notebooks` → `/notebooks`
- `~/gradient/models` → `/models`
- `~/gradient/outputs` → `/outputs`

## GPU Requirement

- Neural is marked `GPURequired=true` in the suite registry.
- CPU-only hosts may still render config, but install and runtime flows should warn that
  the suite requires NVIDIA support.
- AMD detection is informational only in v0.1.

## Service Reference

### `gradient-neural-torch`

- Role: primary PyTorch runtime for model training and interactive experimentation
- Image: `pytorch/pytorch:2.6.0-cuda12.4-cudnn9-runtime`
- Ports: none
- Mounts: data, notebooks, models, outputs
- Environment and config: container requests all NVIDIA GPUs through Compose device
  reservations
- Startup path: long-running helper container with `sleep infinity`
- Dependencies: NVIDIA driver, NVIDIA Container Toolkit, working Docker GPU passthrough
- Health and logs: should remain running and see GPUs through the container runtime
- Troubleshooting: if startup fails, rerun `concave doctor` and `concave driver-wizard`
- Update and rollback behavior: rollback restores the previously recorded training image

### `gradient-neural-infer`

- Role: inference runtime for serving and lightweight HTTP workloads
- Image: `nvidia/cuda:12.4.1-runtime-ubuntu22.04`
- Ports: `8000`, `8080`
- Mounts: data, notebooks, models, outputs
- Environment and config: requests NVIDIA devices through Compose reservations
- Startup path: durable runtime container launched by Compose
- Dependencies: same GPU runtime requirements as `gradient-neural-torch`
- Health and logs: should bind `8000` and `8080` after successful container startup
- Troubleshooting: verify host ports are free and Docker can launch a GPU-enabled
  container
- Update and rollback behavior: image tag history is managed in `versions.json`

### `gradient-neural-lab`

- Role: JupyterLab environment for GPU-assisted notebook workflows
- Image: `quay.io/jupyter/base-notebook:python-3.11.6`
- Ports: `8888`
- Mounts: data, notebooks, models, outputs
- Startup path: started by the suite Compose file and opened through `concave lab`
- Dependencies: healthy notebook container, free port `8888`, host browser opener
- Health and logs: startup logs should expose a tokenized Jupyter URL
- Troubleshooting: if `lab` cannot open, inspect logs and verify the service is running
- Update and rollback behavior: rollback changes image version only, not notebook data

## Failure Modes

- no NVIDIA runtime available
- Docker GPU passthrough check fails
- port `8000`, `8080`, or `8888` already in use
- rendered Compose file invalid
