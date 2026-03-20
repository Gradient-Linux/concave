# Neural Suite

## Purpose

Neural provides GPU-oriented training, inference, and notebook workflows for deep learning workloads.

## Services

- `gradient-neural-torch`
- `gradient-neural-infer`
- `gradient-neural-lab`

## Ports

- `8000` inference API
- `8080` inference or auxiliary HTTP service
- `8888` JupyterLab

## GPU Requirement

Neural requires NVIDIA GPU support for the full runtime path. Detection should warn clearly when the host is CPU-only.
