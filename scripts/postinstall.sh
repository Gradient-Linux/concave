#!/bin/bash
set -euo pipefail

if ! command -v docker >/dev/null 2>&1; then
  cat <<'EOF'

  Docker Engine not found.
  Install it with:
    curl -fsSL https://get.docker.com | sh
    sudo usermod -aG docker $USER
  Then log out and back in, and run: concave setup

EOF
else
  INSTALL_USER="${SUDO_USER:-${USER:-}}"
  if [ -n "${INSTALL_USER}" ] && ! id -nG "${INSTALL_USER}" 2>/dev/null | tr ' ' '\n' | grep -qx docker; then
    usermod -aG docker "${INSTALL_USER}"
    echo "  Added ${INSTALL_USER} to docker group."
    echo "  Log out and back in for this to take effect."
  fi
fi

echo ""
echo "  concave installed. Run: concave setup"
echo ""
