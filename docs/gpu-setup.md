# GPU Setup

`concave` supports three hardware states:

- NVIDIA GPU
- AMD GPU
- CPU-only host

## NVIDIA

The driver workflow determines compute capability, recommends a driver branch, verifies the NVIDIA Container Toolkit, and confirms Docker GPU passthrough with a CUDA image.

## AMD

AMD detection is present for operator visibility. Full ROCm support is deferred and should warn rather than fail the CLI.

## Secure Boot

If Secure Boot is enabled, the driver wizard should never disable it automatically. The user either proceeds with MOK enrollment or exits to disable Secure Boot in firmware and rerun the wizard.
