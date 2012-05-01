Gotgo
=====

This document describes the third iteration of my attempt at a
reasonable implementation of generics for [go](http://golang.org)
based on the idea of template packages.

Quick start
-----------

You can compile `gotgo` by typing

    ./build

and you can run it on an example template file by typing

    ./gotgo/gotgo pkg/gotgo/slice.got int string

If you want to see what a template package will look like, check out
[slice.got][1], which is a simple package exporting handy functions for slices,
such as Map, Fold, Filter, Append, Cat (concat).

[1]: http://github.com/droundy/gotgo/blob/master/gotgo/slice.got

The got file
============

A template package is contained in a got file, ending with the `.got`
extension.  As you might be able to guess, "got" stands for "go
template".  If you call `gotgo` with a `--package-name` flag, you can
generate a go file which is a member of any package you like.

The got file differs from an ordinary go file only in its package
line, which will be something like

    package slice(type a, type b a)

This defines a package template named "slice", with two type
parameters.  The first could be any type, which will be called "a",
and the second type, called "b" will default to be the same as "a", if
no second type parameter is provided.  The default type argument is
optional, and will itself default to interface{}.  The default type
should be either an interface type, or a type defined in terms of
other template parameters, although I'm not sure whether/how to
enforce this.  Your got file *should* be able to compile with the
default type!

Within this package, the types "a" and "b" will be in scope, so the
slice got file might contain:

    func Append(slice []a, datum a) []a {
        // first expand if needed...
        slice = slice[0:len(slice)+1]
        slice[len(slice)-1] = datum
        return slice
    }

We could also have a function that uses both "a" and "b":

    func Map(f func(a) b, slice []a) []b {
        out := make([]b, len(slice))
        for i,v := range slice {
            out[i] = f(v)
        }
        return out
    }

Importing a got package
=======================

If you want to use the above package, you typically do so by creating
an ordinary go file that has a special import statement that specifies
the types desired.  Currently, this is not automated, as gotgo is in
rapid flux.

    import intslice "./slice(int)"

    func foo(x int, xs []int, ...) {
        ...
        intslice.Append(xs,x)

This import is valid go (since the language spec states that the
interpretation of the import string is implementation-dependent), and
is accepted by the go compiler, so long as the `slice(int)` package
has been generated and compiled.

The package name of this templated import may be dependent on the
gotgo implementation unless you use the `--package-name` flag when
running `gotgo`.

Getting fancier with interface types
====================================

We can get a bit fancier in the package specification by using
interfaces.

    package list(type a fmt.Stringer)

    import fmt "fmt"

    type T struct {
        Datum a
        Next *T
    }
    func (l T) String() String {
        if l.Next != nil {
            return l.Datum.String() + ", " + l.Next.String()
        } else {
            return l.Datum.String()
        }
    }

This isn't much different from what I've already described, but there
is (will be) an additional twist that isn't shown, which is that if
you make the default type an interface, then any type parameter passed
to the template *must satisfy that interface*! This is valuable
because it means that you can guarantee that you can't break the
compile of any client code unless you change the template signature.

If you write a template that requires a numeric type, such as

    package maxmin(type a int)
    
    func max(x,y a) a {
        if x > y { return x }
        return y
    }

then you'll need to use a numeric type as your default.  The gotgo
system will then enforce that this package may only be parametrized
with types that are assignment-compatible[3] with `int` (or whatever
default type you choose).

[3]: http://golang.org/doc/go_spec.html#Assignment_compatibility

How does it work?
=================

There is a simple program called `gotgo`, which you pass the filepath
to a `.got` file and a list of types.  By default, `gotgo` outputs the
go code to stdout, but if you specify `-o foo.go`, it will create a
file named `foo.go`.

A zeroconf go build system will need to track down and build imported
packages, and in order to work with `gotgo`, will need to know how to
build templated packages.  An sketch of a helper for such a build
system is provided in the `gotimports` program, which (non-recursively
and incorrectly) dumps some make rules that might be helpful in
creating your own Makefile.


Problems solved by gotgo
------------------------

- In go, there is no way to write a generic data structure, such as a
  linked list.  As the FAQ points out, this isn't *such* a problem,
  since go has a pretty nice stable of built-in data structures, but
  it *is* a problem, since sooner or later you'll run into a need for
  more complicated structures (e.g. perhaps a red-black tree), and you
  don't want to reimplement them for every data type you use.

- In go, there is no way to write a generic function whose return type
  depends on its input, or with two arguments of the same (but
  variable type).  This is a much more serious shortcoming than the
  previous one, as it means you can't even write generic functions for
  the existing built-in generic data structures.  For example,
  consider the generic version of

      func Append(x int, xs []int) []int

  which will append the given value to the slice, expanding it if
  needed.  This is precisely the sort of function you'd like to write
  only once (or twice perhaps, to treat "large" data types
  differently), but currently copy-and-paste is the only way to reuse
  this code.  (And similarly for methods...)

- Compile time penalty of templates.  In C++, using templates means
  recompiling the template every time you use it.  Using `gotgo`,
  however, templates are separately compiled, just like any other go
  package.  So a template need be compiled once by gotgo, and then
  once for each type it is parametrized by, which is the minimal
  number of recompiles consistent with allowing the compiler to
  generate optimal code for each type.

- In the latest `gotgo`, you can use the `--package-name` flag to
  generate a go file in a given package.  This makes it possible to
  use generics with private data types, or in the same package that
  defines a data type.
