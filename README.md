# Lachesis

[documentation](http://docs.fantom.foundation) pages.
[whitepaper](http://www.swirlds.com/downloads/SWIRLDS-TR-2016-01.pdf)
[accompanying document](http://www.swirlds.com/downloads/SWIRLDS-TR-2016-02.pdf).
[patents](http://www.swirlds.com/ip/)


## Docker

Create an `n` node lachesis cluster with:

    n=3 BUILD_DIR="$PWD" ./docker/builder/scale.bash

### Dependencies

  - [Docker](https://www.docker.com/get-started)
  - [jq](https://stedolan.github.io/jq)
  - [batch-ethkey](https://github.com/SamuelMarks/batch-ethkey) with `go get -v github.com/SamuelMarks/batch-ethkey`
