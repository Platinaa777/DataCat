#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
if [[ -e .env ]]; then
  echo '.env already exists; leaving it unchanged.'
  exit 0
fi
umask 077
password=$(openssl rand -hex 24)
key=$(openssl rand -base64 32)
admin_password=$(openssl rand -hex 24)
sed -e "s/replace-with-a-local-password/$password/g" \
    -e "s|generate-with-scripts-init-env|$key|" .env.example > .env
echo "KEYCLOAK_ADMIN_PASSWORD=$admin_password" >> .env
echo 'Created .env. Local login: developer / local-development-only.'
