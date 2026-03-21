#!/usr/bin/env bash
set -e

docker network prune -f
docker compose --profile test up --build --pull never "$@"
