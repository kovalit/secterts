#!/usr/bin/env bash
# Minimal encrypted backup for Secrets Center.
#
# Steps:
#   1. pg_dump the database (values are already encrypted at rest).
#   2. Pack into a tar.gz.
#   3. Encrypt with `age` using a recipient public key.
#   4. Optionally upload with rclone.
#   5. Remove temporary plaintext dump.
#
# Required env:
#   DATABASE_URL          postgres connection string
#   AGE_RECIPIENT         age public key (age1...)
# Optional env:
#   BACKUP_DIR            local output dir (default ./backup)
#   RCLONE_REMOTE         e.g. "myremote:secrets-center" to upload the .age file
set -euo pipefail

: "${DATABASE_URL:?DATABASE_URL is required}"
: "${AGE_RECIPIENT:?AGE_RECIPIENT (age public key) is required}"

BACKUP_DIR="${BACKUP_DIR:-./backup}"
mkdir -p "${BACKUP_DIR}"

STAMP="$(date -u +%Y%m%dT%H%M%SZ)"
DUMP="${BACKUP_DIR}/secrets_center_${STAMP}.sql"
ARCHIVE="${BACKUP_DIR}/secrets_center_${STAMP}.tar.gz"
ENCRYPTED="${ARCHIVE}.age"

cleanup() {
    rm -f "${DUMP}" "${ARCHIVE}"
}
trap cleanup EXIT

echo "==> pg_dump"
pg_dump "${DATABASE_URL}" --no-owner --no-privileges -f "${DUMP}"

echo "==> tar.gz"
tar -czf "${ARCHIVE}" -C "${BACKUP_DIR}" "$(basename "${DUMP}")"

echo "==> age encrypt"
age -r "${AGE_RECIPIENT}" -o "${ENCRYPTED}" "${ARCHIVE}"

echo "==> sha256"
if command -v sha256sum >/dev/null; then
    sha256sum "${ENCRYPTED}"
else
    shasum -a 256 "${ENCRYPTED}"
fi

if [[ -n "${RCLONE_REMOTE:-}" ]]; then
    echo "==> rclone upload to ${RCLONE_REMOTE}"
    rclone copy "${ENCRYPTED}" "${RCLONE_REMOTE}"
fi

echo "==> done: ${ENCRYPTED}"
