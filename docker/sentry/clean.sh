#!/usr/bin/env bash
cd $(dirname $0)

docker volume rm sentry-data
docker volume rm sentry-postgres

rm -f .env .dsn
