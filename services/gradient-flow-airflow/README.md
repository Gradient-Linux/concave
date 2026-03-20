# gradient-flow-airflow

## Purpose

Pipeline orchestration and scheduling service for the Flow suite.

## Image

- `apache/airflow:2.9.0`

## Ports

- `8080`

## Mounts

- `~/gradient/dags -> /opt/airflow/dags`
- `~/gradient/outputs -> /outputs`

## Environment

- `AIRFLOW__CORE__EXECUTOR=LocalExecutor`
- `AIRFLOW__CORE__LOAD_EXAMPLES=False`
