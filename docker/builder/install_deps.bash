#!/usr/bin/env bash

# Do a glide install if vendor directory does not exist
[[ -d vendor ]] || glide install

