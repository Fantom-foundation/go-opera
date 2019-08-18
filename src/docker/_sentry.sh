#!/usr/bin/env bash

declare -rl SENTRY="${SENTRY:-no}"

if [ "$SENTRY" == "yes" ]
then
    echo -e "\nStart Sentry:\n"

    NETWORK=${NETWORK} sentry/start.sh
    source sentry/.dsn
    SENTRY_DSN=--sentry=${SENTRY_DSN}
fi
