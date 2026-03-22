#!/bin/bash
set -euo pipefail

INSTALL_USER="${SUDO_USER:-${USER:-}}"
SERVICE_WORKSPACE_ROOT="/var/lib/gradient"
SUDOERS_PATH="/etc/sudoers.d/gradient-svc"
SERVICE_ENV_PATH="/etc/default/concave-serve"

echo "  Preparing concave system users and groups..."

for group in gradient-admin gradient-operator gradient-developer gradient-viewer; do
  if ! getent group "${group}" >/dev/null 2>&1; then
    groupadd --system "${group}"
    echo "  Created group: ${group}"
  fi
done

if ! id gradient-svc >/dev/null 2>&1; then
  useradd \
    --system \
    --home-dir "${SERVICE_WORKSPACE_ROOT}" \
    --create-home \
    --shell /usr/sbin/nologin \
    --comment "Gradient Linux API Service" \
    gradient-svc
  echo "  Created system user: gradient-svc"
fi

if getent group docker >/dev/null 2>&1; then
  usermod -aG docker gradient-svc 2>/dev/null || true
fi
usermod -aG gradient-admin gradient-svc 2>/dev/null || true

if command -v docker >/dev/null 2>&1; then
  if [ -n "${INSTALL_USER}" ] && getent group docker >/dev/null 2>&1; then
    if ! id -nG "${INSTALL_USER}" 2>/dev/null | tr ' ' '\n' | grep -qx docker; then
      usermod -aG docker "${INSTALL_USER}"
      echo "  Added ${INSTALL_USER} to docker group."
      echo "  Log out and back in for this to take effect."
    fi
  fi
else
  cat <<'EOF'

  Docker Engine not found.
  Install it with:
    curl -fsSL https://get.docker.com | sh
    sudo usermod -aG docker $USER
  Then log out and back in, and run: concave setup

EOF
fi

install -d -m 0755 /etc/default
cat >"${SERVICE_ENV_PATH}" <<EOF
GRADIENT_WORKSPACE_ROOT=${SERVICE_WORKSPACE_ROOT}
CONCAVE_AUTH_CONFIG_PATH=${SERVICE_WORKSPACE_ROOT}/config/auth.json
CONCAVE_SERVE_ADDR=127.0.0.1:7777
EOF
chmod 0644 "${SERVICE_ENV_PATH}"

install -d -o gradient-svc -g gradient-svc -m 0755 "${SERVICE_WORKSPACE_ROOT}"
install -d -o gradient-svc -g gradient-svc -m 0700 "${SERVICE_WORKSPACE_ROOT}/config"
install -d -o gradient-svc -g gradient-svc -m 0755 "${SERVICE_WORKSPACE_ROOT}/compose"
install -d -o gradient-svc -g gradient-svc -m 0755 "${SERVICE_WORKSPACE_ROOT}/backups"
install -d -o gradient-svc -g gradient-svc -m 0755 "${SERVICE_WORKSPACE_ROOT}/logs"

cat >"${SUDOERS_PATH}" <<'EOF'
# Written by concave postinstall
# DO NOT EDIT MANUALLY
gradient-svc ALL=(root) NOPASSWD: /bin/systemctl reboot
gradient-svc ALL=(root) NOPASSWD: /bin/systemctl poweroff
gradient-svc ALL=(root) NOPASSWD: /bin/systemctl restart docker
gradient-svc ALL=(root) NOPASSWD: /usr/local/libexec/concave-host-shell *
EOF
chmod 0440 "${SUDOERS_PATH}"

if command -v systemctl >/dev/null 2>&1 && [ -d /run/systemd/system ]; then
  systemctl daemon-reload
  systemctl enable concave-serve.service >/dev/null 2>&1 || true
  systemctl restart concave-serve.service >/dev/null 2>&1 || true
  echo "  Enabled concave-serve.service"
fi

echo ""
echo "  concave installed."
echo "  API service workspace: ${SERVICE_WORKSPACE_ROOT}"
echo "  Start using the CLI with: concave whoami"
echo ""
