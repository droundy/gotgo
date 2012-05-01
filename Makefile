# Copyright 2010 David Roundy, roundyd@physics.oregonstate.edu.
# All rights reserved.

# TODO(ijt): Remove this file

ifndef GOBIN
GOBIN=$(HOME)/bin
endif

# ugly hack to deal with whitespaces in $GOBIN
nullstring :=
space := $(nullstring) # a space at the end
bindir=$(subst $(space),\ ,$(GOBIN))
pkgdir=$(subst $(space),\ ,$(GOROOT)/pkg/$(GOOS)_$(GOARCH))
srcpkgdir=$(subst $(space),\ ,$(GOROOT)/src/pkg)

.PHONY: test binaries packages install clean
.SUFFIXES: .$(O) .go .got

all: binaries
install: all $(bindir)/gotgo $(bindir)/gotimports .installpkgs
clean:
	rm -f bin/* */*.$(O) */*/*.$(O) */*/*/*.$(O)

binaries:  bin/gotgo bin/gotimports
packages:  pkg/gotgo/box.got pkg/gotgo/finalizer.got pkg/gotgo/slice.got

.go.$(O):
	cd `dirname "$<"`; $(GC) `basename "$<"`

$(srcpkgdir)/%.got: pkg/%.got
	@test -d $(QUOTED_GOROOT)/src/pkg && mkdir -p `dirname $(srcpkgdir)/$*`
	cp $< $@

src/got/buildit.$(O): src/got/buildit.go src/gotgo/slice(string).$(O)
src/got/gotgo.$(O): src/got/gotgo.go src/got/buildit.$(O)
src/gotgo/slice(string).$(O): src/gotgo/slice(string).go

ifneq ($(strip $(shell which gotgo)),)
src/gotgo/slice(string).go: pkg/gotgo/slice.got
	bin/gotgo -o "$@" "$<" string
endif

bin/gotgo: src/gotgo.$(O)
	@mkdir -p bin
	$(LD) -o $@ $<
$(bindir)/gotgo: bin/gotgo
	cp $< $@
src/gotgo.$(O): src/gotgo.go src/gotgo/slice(string).$(O)

bin/gotimports: src/gotimports.$(O)
	@mkdir -p bin
	$(LD) -o $@ $<
$(bindir)/gotimports: bin/gotimports
	cp $< $@
src/gotimports.$(O): src/gotimports.go src/gotgo/slice(string).$(O)

.installpkgs: $(wildcard pkg/gotgo/*.got)
	@mkdir -p $(srcpkgdir)/gotgo/
	cp $? $(srcpkgdir)/gotgo/
	@touch .installpkgs

test:
	cd tests && make test
