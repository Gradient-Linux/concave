# Flow

Flow is the orchestration and observability suite for experiment tracking, scheduling, dashboards, artifact storage, and model serving. It is aimed at users who want a local MLOps stack on one machine.

## Containers

| Name | Image | Role |
|---|---|---|
| `gradient-flow-mlflow` | `ghcr.io/mlflow/mlflow:v2.14.1` | Experiment tracking |
| `gradient-flow-airflow` | `apache/airflow:2.9.0` | Orchestration |
| `gradient-flow-prometheus` | `prom/prometheus:v2.51.0` | Metrics |
| `gradient-flow-grafana` | `grafana/grafana:10.4.0` | Dashboards |
| `gradient-flow-store` | `minio/minio:RELEASE.2024-04-06T05-26-02Z` | Artifact storage |
| `gradient-flow-serve` | `bentoml/model-server:latest` | Model serving |

## Ports

| Port | Service |
|---|---|
| `5000` | MLflow |
| `8080` | Airflow |
| `9090` | Prometheus |
| `3000` | Grafana |
| `9001` | MinIO console |
| `3100` | BentoML endpoint |

## Volume mounts

| Host path | Container path |
|---|---|
| `~/gradient/mlruns` | `/mlruns` |
| `~/gradient/dags` | `/dags` |
| `~/gradient/outputs` | `/outputs` |
| `~/gradient/models` | `/models` |

## GPU requirements

Flow does not require a GPU.

## Install and start

```bash
concave install flow
concave start flow
```

## Open the primary UIs

Direct local URLs after startup:

- MLflow: `http://localhost:5000`
- Airflow: `http://localhost:8080`
- Prometheus: `http://localhost:9090`
- Grafana: `http://localhost:3000`
- MinIO console: `http://localhost:9001`
- BentoML endpoint: `http://localhost:3100`

## Notes

- Flow and Boosting both expose MLflow on port `5000`. The port registry treats that as a shared service.
- `gradient-flow-serve` currently uses `bentoml/model-server:latest`, which is the target image recorded in the suite registry for this build.
