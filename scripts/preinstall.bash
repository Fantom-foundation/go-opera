#!/usr/bin/env bash

function requires() {
  for cmd in "$@"; do
    if [ ! $(command -v "$cmd") ]; then
      >&2 echo "$cmd" 'must be installed';
      exit 1;
    fi
  done
}

requires go glide gcc make
