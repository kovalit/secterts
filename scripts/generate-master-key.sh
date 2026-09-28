#!/usr/bin/env bash
# Generate a 32-byte AES-256 master key, base64 encoded.
# Usage: ./scripts/generate-master-key.sh
set -euo pipefail

KEY="$(openssl rand -base64 32)"
echo "APP_MASTER_KEY_V1_BASE64=${KEY}"
echo ""
echo "Store this key OUTSIDE the database (e.g. project-root .env)."
echo "If you lose it, previously encrypted secrets CANNOT be recovered."
