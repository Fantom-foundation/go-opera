#!/usr/bin/env bash

PROJECT="${PROJECT:-lachesis}"
name="$PROJECT-net"
docker network rm "$name"
docker network create --driver=bridge --subnet="$1" "$name"
