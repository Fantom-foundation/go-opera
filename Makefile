BUILD_TAGS?=lachesis
export DOCKER?=docker
export GLIDE?=glide
export GO?=go
export GREP?=grep
export MOCKGEN?=mockgen
export PROTOC?=protoc
export RM?=rm
export SED?=sed
export SH?=sh
export XARGS?=xargs
ifeq ($(OS),Windows_NT)
	export BROWSER?=start
endif
ifeq ($(shell uname -s),Darwin)
	export BROWSER?=open
endif
export BROWSER?=sensible-browser
export CGO_ENABLED=0

SUBDIRS := src/.
TARGETS := build proto clean buildtests
SUBDIR_TARGETS := $(foreach t,$(TARGETS),$(addsuffix $t,$(SUBDIRS)))
VENDOR_LDFLAG := --ldflags "-X github.com/Fantom-foundation/go-lachesis/src/version.GitCommit=`git rev-parse HEAD`"

ifeq ($(OS),Windows_NT)
    # EXTLDFLAGS := ""
else
    UNAME_S := $(shell uname -s)
    ifeq ($(UNAME_S),Darwin)
        # EXTLDFLAGS := ""
    else
		EXTLDFLAGS := --ldflags '-extldflags "-static"'
    endif
endif

# vendor uses Glide to install all the Go dependencies in vendor/
vendor:
	$(GLIDE) install

# install compiles and places the binary in GOPATH/bin
install:
	$(GO) install \
		$(EXTLDFLAGS) $(VENDOR_LDFLAG) \
		./cmd/lachesis
	$(GO) install \
		$(EXTLDFLAGS) $(VENDOR_LDFLAG) \
		./cmd/network

# build compiles and places the binary in /build
build:
	$(GO) build \
		$(VENDOR_LDFLAG) \
		-o build/lachesis ./cmd/lachesis/main.go
	$(GO) build \
		$(VENDOR_LDFLAG) \
		-o build/network ./cmd/network/

# dist builds binaries for all platforms and packages them for distribution
dist:
	@BUILD_TAGS='$(BUILD_TAGS)' $(SH) -c "'$(CURDIR)/scripts/dist.sh'"

test: buildtests
	$(GLIDE) novendor | $(GREP) -v -e "^\.$$" | CGO_ENABLED=1 $(XARGS) $(GO) test -run "Test.*" -count=1 -tags test -race -timeout 180s

cover:
	$(GLIDE) novendor | $(GREP) -v -e "^\.$$" | CGO_ENABLED=1 $(XARGS) $(GO) test -coverprofile=coverage.out -count=1 -tags test -race -timeout 180s || true
	$(GO) tool cover -html=coverage.out -o coverage.html
	$(BROWSER) coverage.html

# clean up and generate protobuf files
proto: clean

clean:
	$(GLIDE) cc
	$(RM) -rf vendor glide.lock

.PHONY: $(TARGETS) $(SUBDIR_TARGETS) vendor install dist test cover buildtests

# static pattern rule, expands into:
# all clean : % : foo/.% bar/.%
$(TARGETS) : % : $(addsuffix %,$(SUBDIRS))

# here, for foo/.all:
#   $(@D) is foo
#   $(@F) is .all, with leading period
#   $(@F:.%=%) is just all
$(SUBDIR_TARGETS) :
	@$(MAKE) -C $(@D) $(@F:.%=%)
