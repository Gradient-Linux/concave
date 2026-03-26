# Canvas Edition Scaffold

This directory contains the desktop-configuration layer for Gradient Linux Canvas Edition.
It is intentionally ISO-neutral: the files can be consumed by a future image pipeline, cloud-init flow,
or a manual post-install script on top of Ubuntu 24.04.

## Contents

- `scripts/canvas/install.sh`: installs XFCE, LightDM greeter settings, desktop shortcuts, and user config.
- `scripts/canvas/generate-wallpaper.sh`: regenerates the committed wallpaper asset.
- `scripts/canvas/assets/gradient-canvas-wallpaper.svg`: committed wallpaper using the concave-web palette.
- `scripts/canvas/lightdm/`: greeter configuration.
- `scripts/canvas/xfce/`: XFCE, GTK, and keyboard shortcut defaults.
- `scripts/canvas/applications/`: desktop entries for `concave setup`, `concave check`, and the browser surface.

## Smoke Test

1. Start from a fresh Ubuntu 24.04 VM.
2. Install `concave` and copy the repo contents.
3. Run `sudo ./scripts/canvas/install.sh`.
4. Reboot or restart LightDM.
5. Select the XFCE session and confirm:
   - the Gradient wallpaper is active
   - the panel contains setup, check, and browser launchers
   - `Super+C` runs `concave check`
   - `Super+G` runs `concave setup`
   - desktop shortcuts launch without modification
