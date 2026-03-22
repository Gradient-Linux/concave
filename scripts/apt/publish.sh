#!/bin/bash
set -euo pipefail

VERSION="${1:?Usage: publish.sh <version>}"
VERSION="${VERSION#v}"
DIST_DIR="${DIST_DIR:-dist}"
REPO_ROOT="${REPO_ROOT:-/var/www/packages.gradientlinux.io/apt}"
REMOTE="${PACKAGE_SERVER_REMOTE:-packages@packages.gradientlinux.io}"

mapfile -t DEB_FILES < <(find "${DIST_DIR}" -maxdepth 1 -type f -name "concave_${VERSION}_linux_*.deb" | sort)
if [ "${#DEB_FILES[@]}" -eq 0 ]; then
  echo "ERROR: no .deb files found for version ${VERSION} in ${DIST_DIR}"
  exit 1
fi

for deb in "${DEB_FILES[@]}"; do
  scp "${deb}" "${REMOTE}:${REPO_ROOT}/pool/main/c/concave/"
done

ssh "${REMOTE}" GPG_FINGERPRINT="${GPG_FINGERPRINT:?missing GPG_FINGERPRINT}" REPO_ROOT="${REPO_ROOT}" 'bash -s' <<'EOF'
set -euo pipefail
cd "${REPO_ROOT}"

mkdir -p dists/stable/main/binary-amd64 dists/stable/main/binary-arm64
dpkg-scanpackages --arch amd64 pool/main /dev/null > dists/stable/main/binary-amd64/Packages
gzip -fk dists/stable/main/binary-amd64/Packages
dpkg-scanpackages --arch arm64 pool/main /dev/null > dists/stable/main/binary-arm64/Packages
gzip -fk dists/stable/main/binary-arm64/Packages
apt-ftparchive release dists/stable > dists/stable/Release
gpg --batch --yes --pinentry-mode=loopback --local-user "${GPG_FINGERPRINT}" \
  --clearsign -o dists/stable/InRelease dists/stable/Release
gpg --batch --yes --pinentry-mode=loopback --local-user "${GPG_FINGERPRINT}" \
  --armor --detach-sign -o dists/stable/Release.gpg dists/stable/Release
echo "Index updated."
EOF

echo "Published: concave ${VERSION}"
