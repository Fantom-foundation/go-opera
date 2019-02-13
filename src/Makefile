export GLIDE?=glide
export GREP?=grep
export SED?=sed
export SH?=sh
SUBDIRS := poset/. proxy/. peers/.
TARGETS := build proto clean
SUBDIR_TARGETS := $(foreach t,$(TARGETS),$(addsuffix $t,$(SUBDIRS)))

TEST_SUBDIRS := $(shell $(GLIDE) novendor | $(GREP) -v -e "^\.$$" | $(SED) -e "s0^./00" -e "s0/...0/.0" )
TEST_TARGETS := test
TEST_SUBDIR_TARGETS := $(foreach t,$(TEST_TARGETS),$(addsuffix $t,$(TEST_SUBDIRS)))

.PHONY: $(TARGETS) $(SUBDIR_TARGETS) $(TEST_TARGETS) $(TEST_SUBDIR_TARGETS)

# static pattern rule, expands into:
# all clean : % : foo/.% bar/.%
$(TARGETS) : % : $(addsuffix %,$(SUBDIRS))

# here, for foo/.all:
#   $(@D) is foo
#   $(@F) is .all, with leading period
#   $(@F:.%=%) is just all
$(SUBDIR_TARGETS) :
	@$(MAKE) -C $(@D) $(@F:.%=%)

# static pattern rule, expands into:
# all clean : % : foo/.% bar/.%
$(TEST_TARGETS) : % : $(addsuffix %,$(TEST_SUBDIRS))

# here, for foo/.all:
#   $(@D) is foo
#   $(@F) is .all, with leading period
#   $(@F:.%=%) is just all
$(TEST_SUBDIR_TARGETS) :
	@$(MAKE) -C $(@D) $(@F:.%=%)
