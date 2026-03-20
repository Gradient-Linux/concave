# gradient-boost-lab

## Purpose

JupyterLab frontend for the Boosting suite.

## Image

- `quay.io/jupyter/base-notebook:python-3.11.6`

## Ports

- `8888`

## Mounts

- `~/gradient/data -> /data`
- `~/gradient/notebooks -> /notebooks`
- `~/gradient/models -> /models`
- `~/gradient/outputs -> /outputs`
- `~/gradient/mlruns -> /mlruns`

## Startup Path

The `concave lab` command resolves a tokenized Jupyter URL from this container and opens it in the browser.
