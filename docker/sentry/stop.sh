#!/usr/bin/env bash
cd $(dirname $0)

if [ ! -f ".env" ]; then
    > .env
fi

docker-compose down 2> /dev/null # fine if not found
