#!/usr/bin/env bash
set -e

docker compose --profile test down --remove-orphans
docker network prune -f

DOCKER_BUILDKIT=0 docker compose --profile test build --pull=false

docker compose --profile test up "$@"
