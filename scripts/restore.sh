#!/usr/bin/env bash
# Restore a Secrets Center backup produced by backup.sh.
#
# Usage:
#   AGE_IDENTITY=/path/to/key.txt DATABASE_URL=... ./scripts/restore.sh backup/secrets_center_XXX.tar.gz.age
#
# Required env:
#   DATABASE_URL   target postgres connection string
#   AGE_IDENTITY   path to the age private key file
set -euo pipefail

: "${DATABASE_URL:?DATABASE_URL is required}"
: "${AGE_IDENTITY:?AGE_IDENTITY (private key file) is required}"

ENCRYPTED="${1:?Pass the path to the .tar.gz.age backup file}"
WORKDIR="$(mktemp -d)"

cleanup() {
    rm -rf "${WORKDIR}"
}
trap cleanup EXIT

echo "==> age decrypt"
age -d -i "${AGE_IDENTITY}" -o "${WORKDIR}/backup.tar.gz" "${ENCRYPTED}"

echo "==> extract"
tar -xzf "${WORKDIR}/backup.tar.gz" -C "${WORKDIR}"

DUMP="$(find "${WORKDIR}" -name '*.sql' | head -n1)"
if [[ -z "${DUMP}" ]]; then
    echo "No .sql file found in archive" >&2
    exit 1
fi

echo "==> psql restore from ${DUMP}"
echo "WARNING: this applies the dump to ${DATABASE_URL}."
psql "${DATABASE_URL}" -f "${DUMP}"

echo "==> done"
