#!/usr/bin/env bash

set -euo pipefail
OPTIND=1         # Reset in case getopts has been used previously in the shell.

declare debug=0
declare args=
declare -r DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

while getopts "d" opt; do
    case "$opt" in
    d)  debug=1
        ;;
    esac
done

shift $((OPTIND-1))

[ "${1:-}" = "--" ] && shift

if [ "$debug" == "1" ]; then
  args="$args -d"
fi

"$DIR/multi_build.bash" $args
"$DIR/multi_run.bash" $args
