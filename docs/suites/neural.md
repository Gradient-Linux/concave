# Neural

Neural is the GPU-oriented suite for training, inference, and notebook workflows. It is aimed at users who want a PyTorch runtime, a model-serving container, and a JupyterLab entrypoint on the same machine.

## Containers

| Name | Image | Role |
|---|---|---|
| `gradient-neural-torch` | `pytorch/pytorch:2.6.0-cuda12.4-cudnn9-runtime` | Training |
| `gradient-neural-infer` | `nvidia/cuda:12.4.1-runtime-ubuntu22.04` | Inference |
| `gradient-neural-lab` | `quay.io/jupyter/base-notebook:python-3.11.6` | JupyterLab |

## Ports

| Port | Service |
|---|---|
| `8888` | JupyterLab |
| `8000` | vLLM API |
| `8080` | llama.cpp |

## Volume mounts

| Host path | Container path |
|---|---|
| `~/gradient/data` | `/data` |
| `~/gradient/notebooks` | `/notebooks` |
| `~/gradient/models` | `/models` |
| `~/gradient/outputs` | `/outputs` |

## GPU requirements

Neural is the only built-in suite marked as GPU-required. It is designed for NVIDIA runtime support. CPU-only machines can still install the suite, but `concave` warns that NVIDIA support is recommended and the inference and training containers will not be useful without it.

## Install and start

```bash
concave install neural
concave start neural
```

## Open the primary UI

Open JupyterLab with:

```bash
concave lab --suite neural
```

Direct local URLs after startup:

- JupyterLab: `http://localhost:8888`
- Inference API: `http://localhost:8000`
- Auxiliary HTTP endpoint: `http://localhost:8080`

## Notes

- Run `concave gpu check` before installing Neural on a fresh host.
- The suite uses NVIDIA-oriented images and Compose device reservations.
- AMD hardware is detected, but ROCm setup is not part of the current release line.
