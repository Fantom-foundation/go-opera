# Lachesis

[documentation](http://docs.fantom.foundation) pages.


## Docker

Create an 3 node lachesis cluster with:

    n=3 BUILD_DIR="$PWD" ./docker/builder/scale.bash

### Dependencies

  - [Docker](https://www.docker.com/get-started)
  - [jq](https://stedolan.github.io/jq)
  - [Bash](https://www.gnu.org/software/bash)
  - glider base Docker Image with:
    ```bash
    git clone https://github.com/SamuelMarks/evm  # or `cd $GOPATH/src/github.com/SamuelMarks`
    cd evm/docker/glider
    docker build --compress --squash --force-rm --tag "${PWD##*/}" .
    ```
  - [Go](https://golang.org)
  - [batch-ethkey](https://github.com/SamuelMarks/batch-ethkey) with: `go get -v github.com/SamuelMarks/batch-ethkey`
  - [mage](github.com/magefile/mage)