SUBDIRS := poset/. proxy/.
TARGETS := build proto clean
SUBDIR_TARGETS := $(foreach t,$(TARGETS),$(addsuffix $t,$(SUBDIRS)))

.PHONY: $(TARGETS) $(SUBDIR_TARGETS)

# static pattern rule, expands into:
# all clean : % : foo/.% bar/.%
$(TARGETS) : % : $(addsuffix %,$(SUBDIRS))

# here, for foo/.all:
#   $(@D) is foo
#   $(@F) is .all, with leading period
#   $(@F:.%=%) is just all
$(SUBDIR_TARGETS) :
	@$(MAKE) -C $(@D) $(@F:.%=%)
