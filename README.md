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


# How to: Install Go 1.9.1 on Ubuntu 16.04

Introduction

Gois an open source, modern programming language developed by Google that uses high-level syntax similar to scripting languages and makes it easy to build simple, reliable, and efficient software. It is popular for many applications, at many companies, and has a robust set of tools and over 90,000 repos.

This tutorial will walk you through downloading and installing Go 1.9.1, as well as building a simple Hello World application. It’s an update/edit of my other story “How to: Install Go 1.8 on Ubuntu 16.04”

## Prerequisites

* One sudo non-root user

## Step 1 — Installing Go

Let’s install go1.9.1 on your PC or server

If you are ready, update and upgrade the Ubuntu packages on your machine. This ensures that you have the latest security patches and fixes, as well as updated repos for your new packages.

    sudo apt-get update
    sudo apt-get -y upgrade

With that complete, you can begin downloading the latest package for Go by running this command, which will pull down the Go package file, and save it to your current working directory, which you can determine by running pwd.

    sudo curl -O [https://storage.googleapis.com/golang/go1.9.1.linux-amd64.tar.gz](https://storage.googleapis.com/golang/go1.9.1.linux-amd64.tar.gz)

Next, use tar to unpack the package. This command will use the Tar tool to open and expand the downloaded file, and creates a folder using the package name, and then moves it to /usr/local.

    sudo tar -xvf go1.9.1.linux-amd64.tar.gz

    sudo mv go /usr/local

Some users prefer different locations for their Go installation, or may have mandated software locations. The Go package is now in /usr/local which also ensures Go is in your $PATH for Linux. It is possible to install Go to an alternate location but the $PATH information will change. The location you pick to house your Go folder will be referenced later in this tutorial, so remember where you placed it if the location is different than /usr/local.

## Step 2 — Setting Go Paths

In this step, we’ll set some paths that Go needs. The paths in this step are all given are relative to the location of your Go installation in /usr/local. If you chose a new directory, or left the file in download location, modify the commands to match your new location.

First, set Go’s root value, which tells Go where to look for its files.

    sudo nano ~/.profile

At the end of the file, add this line:

    export PATH=$PATH:/usr/local/go/bin

If you chose an alternate installation location for Go, add these lines instead to the same file. This example shows the commands if Go is installed in your home directory:

    export GOROOT=$HOME/go
    export PATH=$PATH:$GOROOT/bin

With the appropriate line pasted into your profile, save and close the file. Next, refresh your profile by running:

    source ~/.profile

## Step 3 — Testing your go 1.9.1 installation

Now that Go is installed and the paths are set for your machine, you can test to ensure that Go is working as expected.

Easy and simplest way: type

    go version //and it should print the installed go version 1.9.1

Create a new directory for your Go workspace, which is where Go will build its files.

    mkdir $HOME/work

Now you can point Go to the new workspace you just created by exporting GOPATH.

    export GOPATH=$HOME/work

For me, the perfect GOPATH is $HOME

    export GOPATH=$HOME

Then, create a directory hierarchy in this folder through this command in order for you to create your test file. You can replace the value user with your GitHub username if you plan to use Git to commit and store your Go code on GitHub. If you do not plan to use GitHub to store and manage your code, your folder structure could be something different, like ~/my_project.

    mkdir -p work/src/github.com/user/hello

Next, you can create a simple “Hello World” Go file.

    nano work/src/github.com/user/hello/hello.go

Inside your editor, paste in the content below, which uses the main Go packages, imports the formatted IO content component, and sets a new function to print ‘Hello World’ when run.

    package main

    import "fmt"

    func main() {
        fmt.Printf("hello, world\n")
    }

This file will show “Hello, World” if it successfully runs, which shows that Go is building files correctly. Save and close the file, then compile it invoking the Go command install.

    go install github.com/user/hello

With the file compiled, you can run it by simply referring to the file at your Go path.

    sudo $GOPATH/bin/hello

If that command returns “Hello World”, then Go is successfully installed and functional [1].

— That’s it-go1.9.1 is installed

## Conclusion

By downloading and installing the latest Go package and setting its paths, you now have a PC/machine to use for Go development.

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
