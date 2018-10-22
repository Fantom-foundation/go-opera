#!/usr/bin/env bash

set -euo pipefail

# Do a glide install if vendor directory does not exist
[[ -d vendor ]] || glide install
