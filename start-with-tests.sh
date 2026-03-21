#!/usr/bin/env bash
set -e

docker compose --profile test up --build "$@"
