// Copyright 2009 Dimiter Stanev, malkia@gmail.com.
// Copyright 2010 David Roundy, roundyd@physics.oregonstate.edu.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"os"
	"fmt"
	"strings"
	"io/ioutil"
	"go/parser"
	"go/ast"
	"strconv"
	"path"
	"flag"
	"bytes"
	"go/printer"
	stringslice "./gotgo/slice(string)"
)

var ignoreInstalled = flag.Bool("ignore-installed", true,
	"don't require that installed gotgo packages be present")

var noGotgo = flag.Bool("without-gotgo", true,
	"don't require that gotgo be installed")

func getImports(filename string) (imports map[string]string, names map[string]string, error os.Error) {
	source, error := ioutil.ReadFile(filename)
	if error != nil {
		return
	}
	file, error := parser.ParseFile(filename, source, parser.ImportsOnly)
	if error != nil {
		return
	}
	for _, importDecl := range file.Decls {
		importDecl, ok := importDecl.(*ast.GenDecl)
		if ok {
			for _, importSpec := range importDecl.Specs {
				importSpec, ok := importSpec.(*ast.ImportSpec)
				if ok {
					importPath, _ := strconv.Unquote(string(importSpec.Path.Value))
					if len(importPath) > 0 {
						if imports == nil {
							imports = make(map[string]string)
							names = make(map[string]string)
						}
						if strings.Index(importPath, "(") != -1 {
							nicepath := path.Clean(importPath)
							imports[nicepath] = importPath
							//names[path.Join(dir,path.Clean(importSpec.Name.String()))]=importPath
							names[importSpec.Name.String()]=nicepath
						}
					}
				}
			}
		}
	}
	return
}

func handleFile(f string) {
	imports, names, err := getImports(f)
	if err != nil {
		fmt.Println("# error: ", err)
	}
	for i,ii := range imports {
		// fmt.Printf("# %s imports %s\n", f, i)
		createGofile(i, ii, names)
	}
}

func createGofile(sourcePath, importPath string, names map[string]string) {
	n := strings.Index(sourcePath, "(")
	// Begin by scanning the type parameters in the import
	types := parseImport(sourcePath)
	basename := sourcePath[0:n]
	goname := path.Join(srcpkgdir, importPath+".go")
	nip := strings.Index(importPath, "(")
	installedGotname := path.Join(srcpkgdir, importPath[0:nip]+".got")

	fmt.Printf("\n# Got template: %s\n", basename)
	fmt.Printf("%s:\n\tgoinstall \"%s\"\n", installedGotname, basename)
	fmt.Printf("%s: %s\n\tnewgotgo -o \"%s\" \"%s\"", goname, installedGotname,
		goname, installedGotname)
	for _, a := range types {
		fmt.Printf(` "%s"`, a)
	}
	fmt.Println()
	fmt.Printf("%s.%s: %s\n\t%sg \"$<\"\n", goname[0:len(goname)-3], arch,
		goname, arch)
	return
}

var srcpkgdir = `$GOROOT/src/pkg`
var pkgdir = `$GOROOT/${GOOS}_${GOARCH}`
// path.Join(os.Getenv("GOROOT"),"pkg",os.Getenv("GOOS")+"_"+os.Getenv("GOARCH"))
var arch = func() string {
	return map[string]string{"amd64":"6","386":"8","arm":"5"}[os.Getenv("GOARCH")]
}()

func main() {
	flag.Parse()
	if len(flag.Args()) != 1 {
		fmt.Println(flag.Args())
		dieWith("gotimports requires just one argument: a .go file")
	}
	handleFile(flag.Args()[0])
}

func dieWith(e string) {
	fmt.Fprintln(os.Stderr, e)
	os.Exit(1)
}

// This parses an import statement, such as "./foo/bar(int,baz.T(foo.T))",
// returning a list of types and a list of imports that are needed
// (e.g. baz and foo in the above example).
func parseImport(s string) (types []string) {
	// First, I want to cut off any preliminary directories, so the
	// import should look like a function call.
	n := strings.Index(s, "(")
	if n < 1 { return }
	start := 0
	for i := n-1; i>=0; i-- {
		if s[i] == '/' {
			start = i+1
			break
		}
	}
	s = s[start:]
	// Now we just need to parse the apparent function call...
	x, _ := parser.ParseExpr(s, s)
	callexpr, ok := x.(*ast.CallExpr)
	if !ok { return } // FIXME: need error handling?
	for _, texpr := range callexpr.Args {
		types = stringslice.Append(types, pretty(texpr))
	}
	return types
}

func pretty(expr interface{}) string {
	b := bytes.NewBufferString("")
	printer.Fprint(b, expr)
	// Here we do a hokey trick to make the pretty-printer give paths
	// without extra spaces.  This is a pretty fragile system, and I
	// really ought to just implement a special pretty printer of my
	// own, since these paths are pretty darn simple.
	x := b.String()
	for i := strings.Index(x," / "); i != -1; i = strings.Index(x," / ") {
		x = x[0:i] + "/" + x[i+3:]
	}
	return x
}

func fileexists(f string) bool {
	x, err := os.Stat(f)
	return err == nil && x.IsRegular()
}
