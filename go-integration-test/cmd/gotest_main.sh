#!/usr/bin/env sh

set -eu

port=$(docker compose --file docker-compose.db.yml ps --format json | \
    jq '.[].Publishers[].PublishedPort')
