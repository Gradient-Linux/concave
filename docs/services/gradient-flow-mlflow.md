# gradient-flow-mlflow

## Purpose

Experiment tracking service for the Flow suite.

## Image

- `ghcr.io/mlflow/mlflow:2.14`

## Ports

- `5000`

## Mounts

- `~/gradient/mlruns -> /mlruns`
- `~/gradient/outputs -> /outputs`

## Notes

Port 5000 is shared with Boosting and must be deduplicated by the port registry when both suites are installed.
