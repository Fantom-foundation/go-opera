BUILD_TAGS?=lachesis
export DOCKER?=docker
export GLIDE?=glide
export GO?=go
export GREP?=grep
export PROTOC?=protoc
export RM?=rm
export SED?=sed
export SH?=sh
export XARGS?=xargs

SUBDIRS := src/.
TARGETS := build proto clean
SUBDIR_TARGETS := $(foreach t,$(TARGETS),$(addsuffix $t,$(SUBDIRS)))

# vendor uses Glide to install all the Go dependencies in vendor/
vendor:
	$(GLIDE) install

# install compiles and places the binary in GOPATH/bin
install:
	$(GO) install --ldflags '-extldflags "-static"' \
		--ldflags "-X github.com/Fantom-foundation/go-lachesis/src/version.GitCommit=`git rev-parse HEAD`" \
		./cmd/lachesis
	$(GO) install --ldflags '-extldflags "-static"' \
		--ldflags "-X github.com/Fantom-foundation/go-lachesis/src/version.GitCommit=`git rev-parse HEAD`" \
		./cmd/network

# build compiles and places the binary in /build
build:
	CGO_ENABLED=0 $(GO) build \
		--ldflags "-X github.com/Fantom-foundation/go-lachesis/src/version.GitCommit=`git rev-parse HEAD`" \
		-o build/lachesis ./cmd/lachesis/main.go
	CGO_ENABLED=0 $(GO) build \
		--ldflags "-X github.com/Fantom-foundation/go-lachesis/src/version.GitCommit=`git rev-parse HEAD`" \
		-o build/network ./cmd/network/

# dist builds binaries for all platforms and packages them for distribution
dist:
	@BUILD_TAGS='$(BUILD_TAGS)' $(SH) -c "'$(CURDIR)/scripts/dist.sh'"

test:
	$(GLIDE) novendor | $(GREP) -v -e "^\.$$" | $(XARGS) $(GO) test -count=1 -tags test -race -timeout 180s

# clean up and generate protobuf files
proto: clean

clean:
	$(RM) -rf vendor

.PHONY: $(TARGETS) $(SUBDIR_TARGETS) vendor install dist test

# static pattern rule, expands into:
# all clean : % : foo/.% bar/.%
$(TARGETS) : % : $(addsuffix %,$(SUBDIRS))

# here, for foo/.all:
#   $(@D) is foo
#   $(@F) is .all, with leading period
#   $(@F:.%=%) is just all
$(SUBDIR_TARGETS) :
	@$(MAKE) -C $(@D) $(@F:.%=%)
