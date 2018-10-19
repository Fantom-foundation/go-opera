#!/usr/bin/env bash

set -xeuo pipefail

declare -r DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

"$DIR/multi_build.bash"
"$DIR/multi_run.bash"
