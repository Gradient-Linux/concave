# Suite Guide

`concave` manages four suite modes:

## Boosting

CPU-first classical machine learning environment with:

- a long-running Python worker container
- JupyterLab on port `8888`
- MLflow on port `5000`

Detailed service, volume, environment, and lifecycle notes live in
[docs/suites/boosting.md](suites/boosting.md).

## Neural

GPU-oriented training and inference environment with:

- a PyTorch runtime
- an inference container exposing `8000` and `8080`
- JupyterLab on port `8888`

Detailed GPU and container notes live in [docs/suites/neural.md](suites/neural.md).

## Flow

Multi-service MLOps stack with:

- MLflow
- Airflow
- Prometheus
- Grafana
- MinIO
- BentoML

Detailed service reference lives in [docs/suites/flow.md](suites/flow.md).

## Forge

User-composed suite mode that selects components from the other suites, resolves port
conflicts, and generates a custom Compose file.

Detailed selection and generation rules live in [docs/suites/forge.md](suites/forge.md).
