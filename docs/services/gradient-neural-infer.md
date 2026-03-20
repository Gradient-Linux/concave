# gradient-neural-infer

## Purpose

Inference runtime container for model serving and lightweight HTTP endpoints in the Neural suite.

## Image

- `nvidia/cuda:12.4-runtime-ubuntu24.04`

## Ports

- `8000`
- `8080`

## Mounts

- `~/gradient/data -> /data`
- `~/gradient/notebooks -> /notebooks`
- `~/gradient/models -> /models`
- `~/gradient/outputs -> /outputs`

## Troubleshooting

If container startup fails, verify the NVIDIA Container Toolkit and GPU passthrough test.
