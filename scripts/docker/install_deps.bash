#!/usr/bin/env bash -euo pipefail

# Do a glide install if vendor directory does not exist
[[ -d vendor ]] || glide install
