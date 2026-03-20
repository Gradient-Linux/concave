# gradient-neural-lab

## Purpose

JupyterLab container for GPU-assisted notebook workflows in the Neural suite.

## Image

- `quay.io/jupyter/base-notebook:python-3.11.6`

## Ports

- `8888`

## Mounts

- `~/gradient/data -> /data`
- `~/gradient/notebooks -> /notebooks`
- `~/gradient/models -> /models`
- `~/gradient/outputs -> /outputs`
