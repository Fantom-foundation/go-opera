export GO?=go
export MOCKGEN?=mockgen

SUBDIRS := internal/.
TARGETS := build proto clean
SUBDIR_TARGETS := $(foreach t,$(TARGETS),$(addsuffix $t,$(SUBDIRS)))

.PHONY: $(TARGETS) $(SUBDIR_TARGETS) test

# static pattern rule, expands into:
# all clean : % : foo/.% bar/.%
$(TARGETS) : % : $(addsuffix %,$(SUBDIRS))

# here, for foo/.all:
#   $(@D) is foo
#   $(@F) is .all, with leading period
#   $(@F:.%=%) is just all
$(SUBDIR_TARGETS) :
	@$(MAKE) -C $(@D) $(@F:.%=%)

test:
	$(GO) test -race -cover -count=1 -timeout 45s

buildtests: mock_test.go

mock_test.go: handlers.go
	$(MOCKGEN) -package=proxy -source=handlers.go -destination=mock_test.go App,Node,Consensus

