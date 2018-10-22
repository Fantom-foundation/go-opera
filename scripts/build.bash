#!/usr/bin/env bash

set -euo pipefail

IFS=$'\n\t'

declare -r DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

. "$DIR/set_globals.bash"

declare -r n="$1"
