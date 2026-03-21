#!/usr/bin/env bash
set -e

docker network prune -f

# Build with legacy builder (skips registry manifest checks when images are cached locally)
DOCKER_BUILDKIT=0 docker compose build --pull=false

docker compose up "$@"
