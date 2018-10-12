#!/usr/bin/env bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

. "$DIR/set_globals.bash"
"$DIR/clean.bash"

# Config
n="${n:-1000}"
ip_start="${ip_start:-192.168.0.2}"
subnet="${subnet:-16}"
ip_range="$ip_start/$subnet"

# Install deps
"$DIR/install_deps.bash"

env GOOS=linux GOARCH=amd64 go build -o lachesis_linux cmd/lachesis/main.go || exit 1

# Run
batch-ethkey -dir "$BUILD_DIR/nodes" -network "$ip_start" -n "$n" > "$PEERS_DIR/peers.json"
ID=$(docker build --compress --force-rm --tag "$PROJECT" "$BUILD_DIR" | tail -1 | sed 's/.*Successfully built \(.*\)$/\1/')
docker tag "$ID" "offscale/$PROJECT":latest

"$DIR/network.bash" "$ip_range"
"$DIR/spin_multi.bash" "$n"

docker start $(docker ps -a --no-trunc --filter name='^/'"$PROJECT" --format '{{.Names}}')
