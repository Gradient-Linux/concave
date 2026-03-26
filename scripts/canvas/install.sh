#!/usr/bin/env bash
set -euo pipefail
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TARGET_USER="${SUDO_USER:-${USER:-}}"
if [[ -z "${TARGET_USER}" ]]; then
  echo "Unable to determine target user." >&2
  exit 1
fi
if [[ "$(id -u)" -ne 0 ]]; then
  echo "Run as root via sudo to install Canvas Edition scaffolding." >&2
  exit 1
fi
TARGET_HOME="$(getent passwd "${TARGET_USER}" | cut -d: -f6)"
if [[ -z "${TARGET_HOME}" ]]; then
  echo "Unable to resolve home directory for ${TARGET_USER}." >&2
  exit 1
fi
apt-get update
DEBIAN_FRONTEND=noninteractive apt-get install -y \
  xfce4 xfce4-goodies lightdm lightdm-gtk-greeter \
  fonts-ibm-plex papirus-icon-theme xdg-utils
install -d -m 0755 /usr/share/backgrounds/gradient-linux
install -m 0644 "${ROOT_DIR}/assets/gradient-canvas-wallpaper.svg" /usr/share/backgrounds/gradient-linux/gradient-canvas-wallpaper.svg
install -d -m 0755 /etc/lightdm/lightdm-gtk-greeter.conf.d
install -m 0644 "${ROOT_DIR}/lightdm/90-gradient-canvas.conf" /etc/lightdm/lightdm-gtk-greeter.conf.d/90-gradient-canvas.conf
install -d -m 0755 /usr/share/applications
install -m 0644 "${ROOT_DIR}/applications/concave-setup.desktop" /usr/share/applications/concave-setup.desktop
install -m 0644 "${ROOT_DIR}/applications/concave-check.desktop" /usr/share/applications/concave-check.desktop
install -m 0644 "${ROOT_DIR}/applications/concave-web.desktop" /usr/share/applications/concave-web.desktop
install -d -m 0755 "${TARGET_HOME}/Desktop"
install -m 0755 "${ROOT_DIR}/applications/concave-setup.desktop" "${TARGET_HOME}/Desktop/concave-setup.desktop"
install -m 0755 "${ROOT_DIR}/applications/concave-check.desktop" "${TARGET_HOME}/Desktop/concave-check.desktop"
install -m 0755 "${ROOT_DIR}/applications/concave-web.desktop" "${TARGET_HOME}/Desktop/concave-web.desktop"
install -d -m 0755 "${TARGET_HOME}/.config/xfce4/xfconf/xfce-perchannel-xml"
install -d -m 0755 "${TARGET_HOME}/.config/xfce4/panel"
install -d -m 0755 "${TARGET_HOME}/.config/gtk-3.0"
install -d -m 0755 "${TARGET_HOME}/.config/gtk-4.0"
install -m 0644 "${ROOT_DIR}/xfce/xfce4-panel.xml" "${TARGET_HOME}/.config/xfce4/xfconf/xfce-perchannel-xml/xfce4-panel.xml"
install -m 0644 "${ROOT_DIR}/xfce/xfce4-desktop.xml" "${TARGET_HOME}/.config/xfce4/xfconf/xfce-perchannel-xml/xfce4-desktop.xml"
install -m 0644 "${ROOT_DIR}/xfce/xfce4-keyboard-shortcuts.xml" "${TARGET_HOME}/.config/xfce4/xfconf/xfce-perchannel-xml/xfce4-keyboard-shortcuts.xml"
install -m 0644 "${ROOT_DIR}/xfce/xsettings.xml" "${TARGET_HOME}/.config/xfce4/xfconf/xfce-perchannel-xml/xsettings.xml"
install -m 0644 "${ROOT_DIR}/xfce/gtk.css" "${TARGET_HOME}/.config/gtk-3.0/gtk.css"
install -m 0644 "${ROOT_DIR}/xfce/gtk.css" "${TARGET_HOME}/.config/gtk-4.0/gtk.css"
chown -R "${TARGET_USER}:${TARGET_USER}" "${TARGET_HOME}/.config/xfce4" "${TARGET_HOME}/.config/gtk-3.0" "${TARGET_HOME}/.config/gtk-4.0" "${TARGET_HOME}/Desktop"
cat <<MSG
Canvas Edition scaffolding installed for ${TARGET_USER}.
Next steps:
  1. Reboot or restart LightDM.
  2. Select the XFCE session at login.
  3. Run 'concave check' and 'concave setup' from the desktop shortcuts.
MSG
