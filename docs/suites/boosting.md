# Boosting

Boosting is the CPU-first suite for classical machine learning, notebook-driven experimentation, and MLflow tracking. It is the easiest suite to bring up on a new host because it does not require GPU runtime support.

## Containers

| Name | Image | Role |
|---|---|---|
| `gradient-boost-core` | `python:3.12-slim` | Core ML stack |
| `gradient-boost-lab` | `quay.io/jupyter/base-notebook:python-3.11.6` | JupyterLab |
| `gradient-boost-track` | `ghcr.io/mlflow/mlflow:v2.14.1` | MLflow tracking |

## Ports

| Port | Service |
|---|---|
| `8888` | JupyterLab |
| `5000` | MLflow |

## Volume mounts

| Host path | Container path |
|---|---|
| `~/gradient/data` | `/data` |
| `~/gradient/notebooks` | `/notebooks` |
| `~/gradient/models` | `/models` |
| `~/gradient/outputs` | `/outputs` |
| `~/gradient/mlruns` | `/mlruns` |

## GPU requirements

Boosting does not require a GPU.

## Install and start

```bash
concave install boosting
concave start boosting
```

## Open the primary UI

Open JupyterLab with:

```bash
concave lab --suite boosting
```

Direct local URLs after startup:

- JupyterLab: `http://localhost:8888`
- MLflow: `http://localhost:5000`

## Notes

- Boosting is the default recommendation for first-time installs.
- `concave rollback boosting` restores the previous image tags recorded in `versions.json` and keeps user data intact.
- Boosting and Flow both expose MLflow on port `5000`. The shared port is handled by the suite port logic.
