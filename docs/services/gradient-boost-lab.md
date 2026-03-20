# gradient-boost-lab

## Purpose

JupyterLab frontend for the Boosting suite.

## Image

- `jupyter/base-notebook:4.0`

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
