#!/usr/bin/env bash

set -euo pipefail

declare -r DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
declare -r parent_dir="${DIR%/*}"

. "$parent_dir/set_globals.bash"
"$DIR/clean.bash"

# Config
declare -r n="${n:-1000}"
declare -r ip_start="${ip_start:-192.168.0.2}"
declare -r subnet="${subnet:-16}"
declare -r ip_range="$ip_start/$subnet"

# Install deps
"$DIR/install_deps.bash"

# Use -tags="netgo multi" in bgo build below to build multu lachesis version for testing
env GOOS=linux GOARCH=amd64 go build -tags="netgo" -ldflags "-linkmode external -extldflags -static -s -w" -o lachesis_linux cmd/lachesis/main.go || exit 1

# Run
batch-ethkey -dir "$BUILD_DIR/nodes" -network "$ip_start" -n "$n" > "$PEERS_DIR/peers.json"
docker build --compress --force-rm --tag "$PROJECT" "$BUILD_DIR"

"$DIR/network.bash" "$ip_range"
"$DIR/spin_multi.bash" "$n"

docker start $(docker ps -a --no-trunc --filter name='^/'"$PROJECT" --format '{{.Names}}')
