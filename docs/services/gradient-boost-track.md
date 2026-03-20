# gradient-boost-track

## Purpose

MLflow tracking UI and metadata backend for the Boosting suite.

## Image

- `ghcr.io/mlflow/mlflow:2.14`

## Ports

- `5000`

## Mounts

- `~/gradient/mlruns -> /mlruns`
- `~/gradient/outputs -> /outputs`

## Health and Logs

The container should expose the MLflow UI on port 5000 and persist tracking metadata under `~/gradient/mlruns`.
