#!/bin/bash
set -euo pipefail

REPO_URL="${REPO_URL:-https://packages.gradientlinux.io}"
KEYRING="/usr/share/keyrings/gradient-linux.gpg"
SOURCES="/etc/apt/sources.list.d/gradient-linux.list"
ARCH="$(dpkg --print-architecture)"

if [ "$(id -u)" -ne 0 ]; then
  echo "Run with sudo: sudo bash setup-repo.sh"
  exit 1
fi

curl -fsSL "${REPO_URL}/gpg" | gpg --dearmor -o "${KEYRING}"
chmod 644 "${KEYRING}"

echo "deb [arch=${ARCH} signed-by=${KEYRING}] ${REPO_URL}/apt stable main" > "${SOURCES}"

apt-get update -qq
apt-get install -y concave

echo ""
echo "concave installed. Run: concave setup"
