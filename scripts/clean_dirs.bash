#!/usr/bin/env bash

# NB: the values of $BUILD_DIR $DIR $parent_dir $gparent_dir should
# be set in the callee.

for dir in "$BUILD_DIR" "$DIR" "$parent_dir" "$gparent_dir"; do
  rm -rf "$dir"/{nodes,peers.json,lachesis_d*}
done
