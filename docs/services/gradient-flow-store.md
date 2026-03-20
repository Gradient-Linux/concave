# gradient-flow-store

## Purpose

Object storage service for models and artifacts in the Flow suite.

## Image

- `minio/minio:RELEASE.2024-04-06T05-26-02Z`

## Ports

- `9001`

## Mounts

- `~/gradient/models -> /data`
- `~/gradient/outputs -> /outputs`
