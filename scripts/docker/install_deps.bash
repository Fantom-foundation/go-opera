#!/usr/bin/env bash

set -euo pipefail

glide=${GLIDE:-glide}
# Do a glide install if vendor directory does not exist
[[ -d vendor ]] || "$glide" install
