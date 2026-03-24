# Forge

Forge is the custom-composition suite for users who want a tailored mix of Boosting, Neural, and Flow components instead of one of the canned suites. It is aimed at power users who know which services they want on a single machine.

## Selectable containers

| Name | Image | Role |
|---|---|---|
| `gradient-boost-core` | `python:3.12-slim` | Core ML stack |
| `gradient-boost-lab` | `quay.io/jupyter/base-notebook:python-3.11.6` | JupyterLab |
| `gradient-boost-track` | `ghcr.io/mlflow/mlflow:v2.14.1` | MLflow tracking |
| `gradient-neural-torch` | `pytorch/pytorch:2.6.0-cuda12.4-cudnn9-runtime` | Training |
| `gradient-neural-infer` | `nvidia/cuda:12.4.1-runtime-ubuntu22.04` | Inference |
| `gradient-neural-lab` | `quay.io/jupyter/base-notebook:python-3.11.6` | JupyterLab |
| `gradient-flow-mlflow` | `ghcr.io/mlflow/mlflow:v2.14.1` | Experiment tracking |
| `gradient-flow-airflow` | `apache/airflow:2.9.0` | Orchestration |
| `gradient-flow-prometheus` | `prom/prometheus:v2.51.0` | Metrics |
| `gradient-flow-grafana` | `grafana/grafana:10.4.0` | Dashboards |
| `gradient-flow-store` | `minio/minio:RELEASE.2024-04-06T05-26-02Z` | Artifact storage |
| `gradient-flow-serve` | `bentoml/model-server:latest` | Model serving |

## Potential ports

| Port | Service |
|---|---|
| `8888` | JupyterLab |
| `8000` | vLLM API |
| `8080` | Airflow or llama.cpp |
| `5000` | MLflow |
| `3000` | Grafana |
| `9090` | Prometheus |
| `9001` | MinIO console |
| `3100` | BentoML endpoint |

## Volume mounts

| Host path | Container path |
|---|---|
| `~/gradient/data` | `/data` |
| `~/gradient/notebooks` | `/notebooks` |
| `~/gradient/models` | `/models` |
| `~/gradient/outputs` | `/outputs` |
| `~/gradient/mlruns` | `/mlruns` |
| `~/gradient/dags` | `/dags` |

## GPU requirements

Forge is not marked GPU-required as a whole. Selections that include Neural training or inference containers still expect NVIDIA runtime support to be useful.

## Install and start

```bash
concave install forge
concave start forge
```

During install, `concave` opens a checklist and asks which components to include. The resulting selection is written into `~/gradient/compose/forge.compose.yml`.

## Open the primary UI

Forge has no single primary UI. The available URLs depend on the selected components:

- JupyterLab if a `*-lab` component is selected
- MLflow if a `*-track` or `gradient-flow-mlflow` component is selected
- Airflow, Grafana, Prometheus, MinIO, or BentoML if the corresponding Flow component is selected

## Notes

- Forge reuses the same images, ports, and mounts as the built-in suites.
- Shared JupyterLab and MLflow components are deduplicated during selection.
- Port conflicts are checked before the Compose file is written.
