# Copyright 2010 David Roundy, roundyd@physics.oregonstate.edu.
# All rights reserved.

include $(GOROOT)/src/Make.$(GOARCH)

ifndef GOBIN
GOBIN=$(HOME)/bin
endif

# ugly hack to deal with whitespaces in $GOBIN
nullstring :=
space := $(nullstring) # a space at the end
bindir=$(subst $(space),\ ,$(GOBIN))
pkgdir=$(subst $(space),\ ,$(GOROOT)/pkg/$(GOOS)_$(GOARCH))
srcpkgdir=$(subst $(space),\ ,$(GOROOT)/src/pkg)

.PHONY: test
.SUFFIXES: .$(O) .go .got

all: example
test: example
	./example	

.go.$(O):
	$(GC) -o "$@" "$<"

example.$(O): example.go \
	$(pkgdir)/gotgo/slice(string).$(O) \
	$(pkgdir)/gotgo/slice(int).$(O) \
	$(pkgdir)/gotgo/list(int).$(O) \
	$(pkgdir)/gotgo/slice(gotgo/list(int).List).$(O) \
	$(pkgdir)/gotgo/sort(float64).$(O) \
	test(int).$(O) \
	test(string).$(O)

example: example.$(O)
	$(LD) -o $@ $<

$(pkgdir)/gotgo/slice(string).go: $(srcpkgdir)/gotgo/slice.got
	gotgo -o "$@" "$<" string
$(pkgdir)/gotgo/slice(int).go: $(srcpkgdir)/gotgo/slice.got
	gotgo -o "$@" "$<" int

$(pkgdir)/gotgo/list(int).go: $(srcpkgdir)/gotgo/list.got
	gotgo -o "$@" "$<" int
$(pkgdir)/gotgo/slice(gotgo/list(int).List).go: $(srcpkgdir)/gotgo/slice.got \
		$(pkgdir)/gotgo/list(int).$(O)
	gotgo -o "$@" "$<" "gotgo/list(int).List"

$(pkgdir)/gotgo/sort(float64).go: $(srcpkgdir)/gotgo/sort.got
	gotgo -o "$@" "$<" float64

test(int).go: test.got
	gotgo -o "$@" "$<" int
test(string).go: test.got
	gotgo -o "$@" "$<" string
