# gradient-neural-torch

## Purpose

Primary PyTorch training environment for the Neural suite.

## Image

- `pytorch/pytorch:2.6.0-cuda12.4-cudnn9-runtime`

## Ports

- None exposed to the host

## Mounts

- `~/gradient/data -> /data`
- `~/gradient/notebooks -> /notebooks`
- `~/gradient/models -> /models`
- `~/gradient/outputs -> /outputs`

## Dependencies

- NVIDIA runtime configuration
- GPU-capable Docker host
