#!/usr/bin/env bash
cd $(dirname $0)

if [ ! -f ".env" ]; then
    > .env
fi

docker-compose down
