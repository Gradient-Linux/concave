# GPU Setup

`concave` supports three hardware states:

- NVIDIA GPU
- AMD GPU
- CPU-only host

CPU-only hosts are valid. They should not fail `concave doctor`, and `concave
driver-wizard` should exit cleanly with guidance that no driver changes are required.

## NVIDIA

The NVIDIA workflow is:

1. detect `nvidia-smi`
2. query compute capability with `nvidia-smi --query-gpu=compute_cap --format=csv,noheader`
3. map capability to a driver branch
4. verify `nvidia-ctk runtime configure --runtime=docker --dry-run`
5. verify Docker passthrough with `docker run --rm --gpus all nvidia/cuda:12.4-base-ubuntu24.04 nvidia-smi`

Driver branch mapping in `concave`:

- `7.x` -> `535`
- `8.0` and `8.6` -> `560`
- `8.9` and `9.0` -> `570`

## AMD

AMD detection is present for operator visibility. Full ROCm support is deferred and
should warn rather than fail the CLI. `concave driver-wizard` should stop after the AMD
warning rather than attempting NVIDIA-specific checks.

## Secure Boot

If Secure Boot is enabled, the driver wizard should never disable it automatically. The
user gets two paths only:

- continue and enroll a MOK key on the next reboot
- exit, disable Secure Boot in firmware, and rerun `concave driver-wizard`

No other automatic Secure Boot mutation is allowed.
