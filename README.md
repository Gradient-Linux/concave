# concave

Control-plane CLI for Gradient Linux machines and AI suite lifecycle.

## What it does

`concave` is the entrypoint for Gradient Linux operators and developers. It creates and maintains the fixed Gradient workspace, installs and manages Docker-based suites, checks machine health, handles GPU setup, and serves the local authenticated API consumed by `concave-tui` and `concave-web`. The binary is built as a static Go executable and is intended to be the only host-level tool most users need.

## Requirements

- Ubuntu 24.04 LTS
- Docker Engine 26+
- Go 1.25+ for source builds
- NVIDIA GPU optional; CPU-only machines are supported

## Install

The fastest install path uses the Gradient Linux package repository:

```bash
curl -fsSL https://packages.gradientlinux.io/install.sh | sudo bash
```

Release binaries are also published through GitHub Releases. For local development, build from source:

```bash
CGO_ENABLED=0 go build -o concave .
sudo install -m 0755 concave /usr/local/bin/concave
```

## Usage

These are the first commands most users run:

```bash
concave setup
concave install boosting
concave start
concave status
concave check
concave env status
concave fleet status
```

## Configuration

`concave` uses a fixed workspace rooted at `~/gradient/`. Runtime state lives in `~/gradient/config/`, including:

- `state.json` for installed suites
- `versions.json` for current and previous image tags
- `setup.json` for setup wizard progress

The local API binds to `127.0.0.1:7777` by default and can be moved with `concave serve --addr`.

## Architecture

`concave` is the infrastructure layer of Gradient Linux. It owns Docker lifecycle, GPU setup, workspace layout, and the local control-plane API. Companion services such as `concave-resolver`, `gradient-mesh`, and `gradient-lab` stay outside that boundary and report state back into `concave`. See [docs/architecture.md](docs/architecture.md) for the full stack view.

## Development

### Prerequisites

Install Go 1.25 or newer and Docker Engine. Some tests and manual flows expect Docker to be running locally.

### Build

```bash
CGO_ENABLED=0 go build -o concave .
```

### Test

```bash
go test ./...
go test -race ./...
```

### Repo layout

```text
concave/
  cmd/          CLI commands
  internal/     workspace, Docker, suite, GPU, auth, API, and system packages
  templates/    rendered Compose template sources
  scripts/      build, release, packaging, and service helpers
  docs/         user, admin, and contributor documentation
```

Packaged installs also place the runtime suite templates under
`/usr/local/share/concave/templates/` so the standalone binary can render
Compose files outside the source tree.

## Roadmap

The current line covers the v0.1 control-plane foundation. v0.2 adds environment intelligence through `concave-resolver`, v0.3 adds fleet discovery through `gradient-mesh`, v0.4 expands compute allocation, and v0.5 adds the dedicated `gradient-lab` collaboration layer.

## License

Released under the MIT License.
