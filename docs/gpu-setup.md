# GPU Setup

This guide covers the current `concave` GPU workflow for a new machine.

## What `concave gpu setup` detects

`concave gpu setup` checks one of three hardware states:

- CPU-only host
- NVIDIA GPU
- AMD GPU

The current build handles those states like this:

- CPU-only: exits cleanly and reports that no driver changes are required
- NVIDIA: checks Secure Boot, recommends a driver branch, verifies the container toolkit, and validates Docker GPU passthrough
- AMD: reports that ROCm support is planned for v0.3 and stops there

## Recommended first check

Inspect the machine before you change anything:

```bash
concave gpu info
concave gpu check
```

`concave gpu info` reports the detected GPU, current driver details, compute capability, and recommended NVIDIA driver branch when available. `concave gpu check` runs the GPU-specific health checks without entering the setup flow.

## Run the setup flow

Start the interactive setup path with:

```bash
concave gpu setup
```

On an NVIDIA machine, the current flow does this:

1. Detects the GPU state.
2. Checks whether Secure Boot is enabled.
3. If Secure Boot is enabled, offers two paths only:
   - continue and enroll a MOK key on the next reboot
   - exit and disable Secure Boot in firmware before rerunning setup
4. Maps compute capability to a recommended driver branch.
5. Verifies `nvidia-container-toolkit`.
6. Runs a Docker passthrough check with `nvidia-smi` inside a test container.

Current branch mapping:

- `7.x` -> `535`
- `8.0`, `8.6` -> `560`
- `8.9`, `9.0` -> `570`

## Secure Boot and MOK enrollment

If Secure Boot is enabled, `concave` does not try to disable it for you. The setup flow tells you to continue with MOK enrollment or stop and change firmware settings yourself.

You can check the platform state directly with:

```bash
mokutil --sb-state
```

## About `--dry-run`

The current build does not expose `concave gpu setup --dry-run`. If you want a non-mutating preview, use:

```bash
concave gpu info
concave gpu check
nvidia-ctk runtime configure --runtime=docker --dry-run
```

## Verify the result

After setup, confirm the machine and container runtime state:

```bash
concave gpu check
concave gpu info
concave check
docker run --rm --gpus all nvidia/cuda:12.4-base-ubuntu24.04 nvidia-smi
```

`concave gpu check` should report the detected GPU and a configured toolkit on NVIDIA hosts.

## CPU-only and AMD hosts

CPU-only machines are valid. `concave gpu setup` reports that no driver changes are required and returns without error.

AMD detection is present for visibility. The current release line does not configure ROCm. If the machine reports AMD hardware, stop after the warning and keep the host in CPU-only operation until the AMD path lands.

## If something goes wrong

Start with these commands:

```bash
concave gpu info
concave gpu check
concave check
```

If Docker GPU passthrough still fails, check the toolkit and runtime manually:

```bash
nvidia-ctk runtime configure --runtime=docker --dry-run
sudo systemctl restart docker
docker run --rm --gpus all nvidia/cuda:12.4-base-ubuntu24.04 nvidia-smi
```

If Secure Boot blocked the driver path, confirm the MOK enrollment step on reboot or disable Secure Boot in firmware and rerun:

```bash
concave gpu setup
```
