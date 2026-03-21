#!/usr/bin/env bash
set -e

docker compose down --remove-orphans
docker network prune -f

DOCKER_BUILDKIT=0 docker compose build --pull=false

docker compose up "$@"
