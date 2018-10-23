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
  - [Go](https://golang.org)
  - [Glide](https://glide.sh)
  - [batch-ethkey](https://github.com/SamuelMarks/batch-ethkey) with: `go get -v github.com/SamuelMarks/batch-ethkey`
  - [mage](github.com/magefile/mage)

# LACHESIS
## BFT Consensus platform for distributed applications.

[![CircleCI](https://circleci.com/gh/andrecronje/lachesis.svg?style=svg)](https://circleci.com/gh/andrecronje/lachesis)

## Dev

### Go
Lachesis is written in [Golang](https://golang.org/). Hence, the first step is to
install **Go version 1.9 or above** which is both the programming language and a
CLI tool for managing Go code. Go is very opinionated and will require you to
[define a workspace](https://golang.org/doc/code.html#Workspaces) where all your
go code will reside.

### Lachesis and dependencies
Clone the [repository](https://github.com/andrecronje/lachesis) in the appropriate
GOPATH subdirectory:

```bash
$ mkdir -p $GOPATH/src/github.com/andrecronje/
$ cd $GOPATH/src/github.com/andrecronje
[...]/andrecronje$ git clone https://github.com/andrecronje/lachesis.git
```
Lachesis uses [Glide](http://github.com/Masterminds/glide) to manage dependencies.

```bash
[...]/lachesis$ curl https://glide.sh/get | sh
[...]/lachesis$ glide install
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
?       github.com/andrecronje/lachesis/src/lachesis     [no test files]
ok      github.com/andrecronje/lachesis/src/common     0.015s
ok      github.com/andrecronje/lachesis/src/crypto     0.122s
ok      github.com/andrecronje/lachesis/src/poset  10.270s
?       github.com/andrecronje/lachesis/src/mobile     [no test files]
ok      github.com/andrecronje/lachesis/src/net        0.012s
ok      github.com/andrecronje/lachesis/src/node       19.171s
ok      github.com/andrecronje/lachesis/src/peers      0.038s
?       github.com/andrecronje/lachesis/src/proxy      [no test files]
ok      github.com/andrecronje/lachesis/src/proxy/dummy        0.013s
ok      github.com/andrecronje/lachesis/src/proxy/inmem        0.037s
ok      github.com/andrecronje/lachesis/src/proxy/socket       0.009s
?       github.com/andrecronje/lachesis/src/proxy/socket/app   [no test files]
?       github.com/andrecronje/lachesis/src/proxy/socket/lachesis        [no test files]
?       github.com/andrecronje/lachesis/src/service    [no test files]
?       github.com/andrecronje/lachesis/src/version    [no test files]
?       github.com/andrecronje/lachesis/cmd/lachesis     [no test files]
?       github.com/andrecronje/lachesis/cmd/lachesis/commands    [no test files]
?       github.com/andrecronje/lachesis/cmd/dummy      [no test files]
?       github.com/andrecronje/lachesis/cmd/dummy/commands     [no test files]

```

## Build from source

The easiest way to build binaries is to do so in a hermetic Docker container.
Use this simple command:

```bash
[...]/lachesis$ make dist
```
This will launch the build in a Docker container and write all the artifacts in
the build/ folder.

```bash
[...]/lachesis$ tree build
build/
├── dist
│   ├── lachesis_0.1.0_darwin_386.zip
│   ├── lachesis_0.1.0_darwin_amd64.zip
│   ├── lachesis_0.1.0_freebsd_386.zip
│   ├── lachesis_0.1.0_freebsd_arm.zip
│   ├── lachesis_0.1.0_linux_386.zip
│   ├── lachesis_0.1.0_linux_amd64.zip
│   ├── lachesis_0.1.0_linux_arm.zip
│   ├── lachesis_0.1.0_SHA256SUMS
│   ├── lachesis_0.1.0_windows_386.zip
│   └── lachesis_0.1.0_windows_amd64.zip
└── pkg
    ├── darwin_386
    │   └── lachesis
    ├── darwin_amd64
    │   └── lachesis
    ├── freebsd_386
    │   └── lachesis
    ├── freebsd_arm
    │   └── lachesis
    ├── linux_386
    │   └── lachesis
    ├── linux_amd64
    │   └── lachesis
    ├── linux_arm
    │   └── lachesis
    ├── windows_386
    │   └── lachesis.exe
    └── windows_amd64
        └── lachesis.exe
```
