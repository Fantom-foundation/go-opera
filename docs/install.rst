.. _install:

Install
=======

From Source
^^^^^^^^^^^

Clone the `repository <https://github.com/Fantom-foundation/go-lachesis>`__ in the appropriate GOPATH subdirectory:

::

    $ mkdir -p $GOPATH/src/github.com/Fantom-foundation/
    $ cd $GOPATH/src/github.com/Fantom-foundation
    [...]/Fantom-foundation$ git clone https://github.com/Fantom-foundation/go-lachesis.git


The easiest way to build binaries is to do so in a hermetic Docker container.
Use this simple command:

::

	[...]/lachesis$ make dist

This will launch the build in a Docker container and write all the artifacts in
the build/ folder.

::

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

Go Devs
^^^^^^^

Lachesis is written in `Golang <https://golang.org/>`__. Hence, the first step is
to install **Go version 1.9 or above** which is both the programming language
and a CLI tool for managing Go code. Go is very opinionated  and will require
you to `define a workspace <https://golang.org/doc/code.html#Workspaces>`__
where all your go code will reside.

Dependencies
^^^^^^^^^^^^

Lachesis uses `Glide <http://github.com/Masterminds/glide>`__ to manage
dependencies. For Ubuntu users:

::

    [...]/lachesis$ curl https://glide.sh/get | sh
    [...]/lachesis$ glide install

This will download all dependencies and put them in the **vendor** folder.

Testing
^^^^^^^

Lachesis has extensive unit-testing. Use the Go tool to run tests:

::

    [...]/lachesis$ make test

If everything goes well, it should output something along these lines:

::

    ?       github.com/Fantom-foundation/go-lachesis/src/lachesis     [no test files]
    ok      github.com/Fantom-foundation/go-lachesis/src/common     0.015s
    ok      github.com/Fantom-foundation/go-lachesis/src/crypto     0.122s
    ok      github.com/Fantom-foundation/go-lachesis/src/poset  10.270s
    ?       github.com/Fantom-foundation/go-lachesis/src/mobile     [no test files]
    ok      github.com/Fantom-foundation/go-lachesis/src/net        0.012s
    ok      github.com/Fantom-foundation/go-lachesis/src/node       19.171s
    ok      github.com/Fantom-foundation/go-lachesis/src/peers      0.038s
    ?       github.com/Fantom-foundation/go-lachesis/src/proxy      [no test files]
    ok      github.com/Fantom-foundation/go-lachesis/src/proxy/dummy        0.013s
    ok      github.com/Fantom-foundation/go-lachesis/src/proxy/inmem        0.037s
    ok      github.com/Fantom-foundation/go-lachesis/src/proxy/socket       0.009s
    ?       github.com/Fantom-foundation/go-lachesis/src/proxy/socket/app   [no test files]
    ?       github.com/Fantom-foundation/go-lachesis/src/proxy/socket/lachesis        [no test files]
    ?       github.com/Fantom-foundation/go-lachesis/src/service    [no test files]
    ?       github.com/Fantom-foundation/go-lachesis/src/version    [no test files]
    ?       github.com/Fantom-foundation/go-lachesis/cmd/lachesis     [no test files]
    ?       github.com/Fantom-foundation/go-lachesis/cmd/lachesis/commands    [no test files]
    ?       github.com/Fantom-foundation/go-lachesis/cmd/dummy      [no test files]
    ?       github.com/Fantom-foundation/go-lachesis/cmd/dummy/commands     [no test files]
