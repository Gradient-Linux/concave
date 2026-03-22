# Flow Suite

## Purpose

Flow is the orchestration and observability suite for experiment tracking, pipelines,
serving, object storage, metrics, and dashboards.

## Containers

- `gradient-flow-mlflow`
- `gradient-flow-airflow`
- `gradient-flow-prometheus`
- `gradient-flow-grafana`
- `gradient-flow-store`
- `gradient-flow-serve`

## Host Ports

- `5000` → MLflow
- `8080` → Airflow UI
- `9090` → Prometheus
- `3000` → Grafana
- `9001` → MinIO console
- `3100` → BentoML

## Workspace Mounts

- `~/gradient/mlruns` → `/mlruns`
- `~/gradient/dags` → `/dags`
- `~/gradient/outputs` → `/outputs`
- `~/gradient/models` → `/models`

## Service Reference

### `gradient-flow-mlflow`

- Role: experiment tracking backend for the Flow suite
- Image: `ghcr.io/mlflow/mlflow:v2.14.1`
- Ports: `5000`
- Mounts: `mlruns`, `outputs`
- Startup path: MLflow UI mode against workspace-backed tracking storage
- Dependencies: writable `~/gradient/mlruns`, free port `5000`
- Health and logs: should expose the UI and persist metadata in `~/gradient/mlruns`
- Troubleshooting: if Boosting is also installed, shared MLflow port handling is owned
  by the shared port registry
- Update and rollback behavior: image swaps do not remove tracking data

### `gradient-flow-airflow`

- Role: pipeline orchestration and scheduling
- Image: `apache/airflow:2.9.0`
- Ports: `8080`
- Mounts: `dags`, `outputs`
- Environment and config:
  - `AIRFLOW__CORE__EXECUTOR=LocalExecutor`
  - `AIRFLOW__CORE__LOAD_EXAMPLES=False`
- Startup path: runs `airflow standalone`
- Dependencies: writable DAGs directory, free port `8080`
- Health and logs: logs should show scheduler and webserver startup
- Troubleshooting: inspect permissions and Airflow startup logs if the web UI does not
  appear
- Update and rollback behavior: rollback restores image tag only

### `gradient-flow-prometheus`

- Role: metrics collection
- Image: `prom/prometheus:v2.51.0`
- Ports: `9090`
- Mounts: `outputs`
- Startup path: standard Prometheus container startup
- Dependencies: free port `9090`
- Health and logs: container should stay running and expose the UI on `9090`
- Troubleshooting: inspect Compose config and logs if the service exits on boot
- Update and rollback behavior: no user artifacts are removed during image changes

### `gradient-flow-grafana`

- Role: dashboards and visualization
- Image: `grafana/grafana:10.4.0`
- Ports: `3000`
- Mounts: `outputs`
- Startup path: standard Grafana server startup
- Dependencies: free port `3000`
- Health and logs: should expose the login page on `3000`
- Troubleshooting: verify no other local Grafana instance owns the same port
- Update and rollback behavior: rollback changes the image tag only

### `gradient-flow-store`

- Role: object storage for model and artifact workflows
- Image: `minio/minio:RELEASE.2024-04-06T05-26-02Z`
- Ports: `9001`
- Mounts: `models` mounted at `/data`, `outputs`
- Startup path: runs MinIO with the console exposed on `9001`
- Dependencies: free port `9001`, writable mounted data path
- Health and logs: startup logs should confirm the API and console endpoints
- Troubleshooting: inspect mounts and permissions if object storage fails to initialize
- Update and rollback behavior: image rollback preserves stored artifacts

### `gradient-flow-serve`

- Role: model serving endpoint container
- Image: `bentoml/model-server:latest`
- Ports: `3100`
- Mounts: `models`, `outputs`
- Startup path: long-running serving runtime
- Dependencies: free port `3100`, accessible model artifacts in the workspace
- Health and logs: should stay running and expose the service endpoint on `3100`
- Troubleshooting: inspect service logs and model mounts if requests fail
- Update and rollback behavior: rollback restores the previous serving image tag

## Failure Modes

- MLflow port collision with Boosting
- port `8080`, `9090`, `3000`, `9001`, or `3100` already in use
- invalid Compose output
- missing DAG or tracking directories
