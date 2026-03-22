#!/usr/bin/env bash
set -e

docker compose --profile test down -v --remove-orphans
docker network prune -f

DOCKER_BUILDKIT=0 docker compose --profile test build --pull=false

# Start the database and wait for it to be healthy.
docker compose up -d --wait db

# Apply test-DB migrations (depends_on is not enforced in run mode,
# but db is already healthy above).
docker compose --profile test run --rm migrate-test

# Run the test suite against the test database.
docker compose --profile test run --rm run-tests
