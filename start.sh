#!/usr/bin/env bash
set -e

docker network prune -f
docker compose up --build --pull never "$@"
