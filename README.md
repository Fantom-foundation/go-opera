# Lachesis
## BFT Consensus platform for distributed applications.
[![Build Status](https://travis-ci.org/andrecronje/lachesis.svg?branch=master)](https://travis-ci.org/andrecronje/lachesis)

[documentation](http://docs.fantom.foundation) pages.

## Dev

### Docker

Create an 3 node lachesis cluster with:

    n=3 BUILD_DIR="$PWD" ./scripts/docker/scale.bash

### Dependencies

  - [Docker](https://www.docker.com/get-started)
  - [jq](https://stedolan.github.io/jq)
  - [Bash](https://www.gnu.org/software/bash)
  - [git](https://git-scm.com)
  - [Go](https://golang.org)
  - [Glide](https://glide.sh)
  - [batch-ethkey](https://github.com/SamuelMarks/batch-ethkey) with: `go get -v github.com/SamuelMarks/batch-ethkey`
  - [mage](github.com/magefile/mage)
  - [protocol buffers 3](https://github.com/protocolbuffers/protobuf), with: installation of [a release]([here](https://github.com/protocolbuffers/protobuf/releases)) & `go get -u github.com/golang/protobuf/protoc-gen-go`

### Protobuffer 3

This project uses protobuffer 3 for the communication between posets.
To use it, you have to install both `protoc` and the plugin for go code
generation.

Once the stack is setup, you can compile the proto messages by
running this command:

```bash
make proto
```

### Lachesis and dependencies
Clone the [repository](https://github.com/andrecronje/lachesis) in the appropriate
GOPATH subdirectory:

```bash
$ d="$GOPATH/src/github.com/andrecronje"
$ mkdir -p "$d"
$ git clone https://github.com/andrecronje/lachesis.git "$d"
```
Lachesis uses [Glide](http://github.com/Masterminds/glide) to manage dependencies.

```bash
$ curl https://glide.sh/get | sh
$ cd "$GOPATH/src/github.com/andrecronje" && glide install
```
This will download all dependencies and put them in the **vendor** folder.

### Other requirements

Bash scripts used in this project assume the use of GNU versions of coreutils.
Please ensure you have GNU versions of these programs installed:-

example for macos:
```
# --with-default-names makes the `sed` and `awk` commands default to gnu sed and gnu awk respectively.
brew install gnu-sed gawk --with-default-names
```

### Testing

Lachesis has extensive unit-testing. Use the Go tool to run tests:
```bash
[...]/lachesis$ make test
```

If everything goes well, it should output something along these lines:
```
?   	github.com/andrecronje/lachesis/cmd/dummy	[no test files]
?   	github.com/andrecronje/lachesis/cmd/dummy/commands	[no test files]
?   	github.com/andrecronje/lachesis/cmd/dummy_client	[no test files]
?   	github.com/andrecronje/lachesis/cmd/lachesis	[no test files]
?   	github.com/andrecronje/lachesis/cmd/lachesis/commands	[no test files]
?   	github.com/andrecronje/lachesis/tester	[no test files]
ok  	github.com/andrecronje/lachesis/src/common	(cached)
ok  	github.com/andrecronje/lachesis/src/crypto	(cached)
ok  	github.com/andrecronje/lachesis/src/difftool	(cached)
ok  	github.com/andrecronje/lachesis/src/dummy	0.522s
?   	github.com/andrecronje/lachesis/src/lachesis	[no test files]
?   	github.com/andrecronje/lachesis/src/log	[no test files]
?   	github.com/andrecronje/lachesis/src/mobile	[no test files]
ok  	github.com/andrecronje/lachesis/src/net	(cached)
ok  	github.com/andrecronje/lachesis/src/node	9.832s
?   	github.com/andrecronje/lachesis/src/pb	[no test files]
ok  	github.com/andrecronje/lachesis/src/peers	(cached)
ok  	github.com/andrecronje/lachesis/src/poset	9.627s
ok  	github.com/andrecronje/lachesis/src/proxy	1.019s
?   	github.com/andrecronje/lachesis/src/proxy/internal	[no test files]
?   	github.com/andrecronje/lachesis/src/proxy/proto	[no test files]
?   	github.com/andrecronje/lachesis/src/service	[no test files]
?   	github.com/andrecronje/lachesis/src/utils	[no test files]
?   	github.com/andrecronje/lachesis/src/version	[no test files]
```

## Cross-build from source

The easiest way to build binaries is to do so in a hermetic Docker container.
Use this simple command:

```bash
[...]/lachesis$ make dist
```
This will launch the build in a Docker container and write all the artifacts in
the build/ folder.

```bash
[...]/lachesis$ tree --charset=nwildner build
build
|-- dist
|   |-- lachesis_0.4.3_SHA256SUMS
|   |-- lachesis_0.4.3_darwin_386.zip
|   |-- lachesis_0.4.3_darwin_amd64.zip
|   |-- lachesis_0.4.3_freebsd_386.zip
|   |-- lachesis_0.4.3_freebsd_arm.zip
|   |-- lachesis_0.4.3_linux_386.zip
|   |-- lachesis_0.4.3_linux_amd64.zip
|   |-- lachesis_0.4.3_linux_arm.zip
|   |-- lachesis_0.4.3_windows_386.zip
|   `-- lachesis_0.4.3_windows_amd64.zip
|-- lachesis
`-- pkg
    |-- darwin_386
    |   `-- lachesis
    |-- darwin_386.zip
    |-- darwin_amd64
    |   `-- lachesis
    |-- darwin_amd64.zip
    |-- freebsd_386
    |   `-- lachesis
    |-- freebsd_386.zip
    |-- freebsd_arm
    |   `-- lachesis
    |-- freebsd_arm.zip
    |-- linux_386
    |   `-- lachesis
    |-- linux_386.zip
    |-- linux_amd64
    |   `-- lachesis
    |-- linux_amd64.zip
    |-- linux_arm
    |   `-- lachesis
    |-- linux_arm.zip
    |-- windows_386
    |   `-- lachesis.exe
    |-- windows_386.zip
    |-- windows_amd64
    |   `-- lachesis.exe
    `-- windows_amd64.zip

11 directories, 29 files
```
