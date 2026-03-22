# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

### Added

- Initial repository bootstrap for `concave`
- Core CLI scaffold, workspace management, system checks, and Compose templates
- Base architecture, suite, and service documentation
- Signal-aware mutating commands with operation locking and local crash logging
- Verbose structured logging, retry/backoff for pull operations, and suite health polling
- Shell completion generation, Goreleaser packaging, Debian postinstall hook, APT scripts, and release workflow

### Changed

- `concave remove <suite>` now cleans suite state even when the rendered compose file is missing
- `concave setup` now resumes through `~/gradient/config/setup.json`
- Long-running mutating commands now use consistent progress reporting
