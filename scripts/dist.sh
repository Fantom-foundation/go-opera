#!/usr/bin/env bash
set -e

# Get the version from the environment, or try to figure it out.
if [ -z "$VERSION" ]; then
	declare -i n=10;
	VERSION='';
	while read -r l; do
	  (( n-- )) || true;
	  if [ $n -eq 0 ]; then
	    break;
	  fi
	  l=${l:12};
	  if [[ "${l:0:1}" == '"' ]]; then
	    VERSION="${VERSION}"."${l:1:-1}";
	  fi
	done<src/version/version.go
    VERSION="${VERSION:1}"
fi
if [ -z "$VERSION" ]; then
    echo "Please specify a version."
    exit 1
fi
echo "==> Building version $VERSION..."

# Get the parent directory of where this script is.
SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ] ; do SOURCE="$(readlink "$SOURCE")"; done
DIR="$( cd -P "$( dirname "$SOURCE" )/.." && pwd )"

# Change into that dir because we expect that.
cd "$DIR"

# Delete the old dir
echo "==> Removing old directory..."
rm -rf build/pkg
mkdir -p build/pkg

# Do a hermetic build inside a Docker container.
docker run --rm  \
    -u `id -u $USER` \
    -e "BUILD_TAGS=$BUILD_TAGS" \
    -v "$(pwd)":/go/src/github.com/andrecronje/lachesis \
    -w /go/src/github.com/andrecronje/lachesis \
    offscale/golang-builder-alpine3.8 ./scripts/dist_build.sh

# Add "lachesis" and $VERSION prefix to package name.
rm -rf ./build/dist
mkdir -p ./build/dist
for FILENAME in $(find ./build/pkg -mindepth 1 -maxdepth 1 -type f); do
  FILENAME=$(basename "$FILENAME")
	cp "./build/pkg/${FILENAME}" "./build/dist/lachesis_${VERSION}_${FILENAME}"
done

# Make the checksums.
pushd ./build/dist
shasum -a256 ./* > "./lachesis_${VERSION}_SHA256SUMS"
popd

# Done
echo
echo "==> Results:"
ls -hl ./build/dist

exit 0
