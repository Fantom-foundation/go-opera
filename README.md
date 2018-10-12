# Lachesis

[documentation](http://docs.fantom.foundation) pages.


## Docker

Create an 3 node lachesis cluster with:

    n=3 BUILD_DIR="$PWD" ./scripts/docker/scale.bash

### Dependencies

  - [Docker](https://www.docker.com/get-started)
  - [jq](https://stedolan.github.io/jq)
  - [Bash](https://www.gnu.org/software/bash)
  - [git](https://git-scm.com)
  - glider base Docker Image, via:
    ```bash
    git clone https://github.com/Fantom-foundation/fantom-docker
    cd fantom-docker/glider
    docker build --compress --force-rm --tag "${PWD##*/}" .
    ```
  - [Go](https://golang.org)
  - [batch-ethkey](https://github.com/SamuelMarks/batch-ethkey) with: `go get -v github.com/SamuelMarks/batch-ethkey`
