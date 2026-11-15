#!/usr/bin/env bash
# Creates .env from .env.example with random passwords. Never overwrites an
# existing .env. Passwords are hex, so they are safe inside connection URLs.
set -euo pipefail

cd "$(dirname "$0")/.."

if [ -f .env ]; then
    echo ".env already exists, leaving it untouched"
    exit 0
fi

random() { openssl rand -hex 24; }

umask 077

pg="$(random)"
mq="$(random)"

# In GitHub Actions, make sure these values are redacted if they ever reach a log.
if [ "${GITHUB_ACTIONS:-}" = "true" ]; then
    echo "::add-mask::$pg"
    echo "::add-mask::$mq"
fi

sed -e "s/^POSTGRES_PASSWORD=.*/POSTGRES_PASSWORD=$pg/" \
    -e "s/^RABBITMQ_PASSWORD=.*/RABBITMQ_PASSWORD=$mq/" \
    .env.example > .env

echo "created .env with generated passwords (not printed)"
