package main

import (
	"os"
	"fmt"
	"flag"
	"path"
	"strings"
	"io/ioutil"
	"go/token"
	"go/scanner"
	stringslice "./slice"
)

func dieWith(e string) {
	fmt.Fprintln(os.Stderr, e)
	os.Exit(1)
}

var pname = flag.String("package-name", "NAME", "name of output package")
var prefix = flag.String("prefix", "NAME", "prefix for top-level functions")
var outname = flag.String("o", "GONAME", "name of output file")

func main() {
	flag.Parse()
	if len(flag.Args()) < 1 {
		fmt.Println(flag.Args())
		dieWith("gotgo requires at least one argument: the .got file")
	}
	if flag.Args()[0][len(flag.Args()[0])-4:] != ".got" {
		dieWith("the argument should end with .got")
	}
	var e os.Error
	outf := os.Stdout
	if *outname != "GONAME" {
		dir,_ := path.Split(*outname)
		os.MkdirAll(dir, 0777)
		outf, e = os.Open(*outname, os.O_WRONLY+os.O_CREAT+os.O_TRUNC,0666)
		if e != nil { dieWith(e.String()) }
	}
	e = writeGotGotgo(flag.Args()[0], outf, flag.Args()[1:])
	if e != nil { dieWith(e.String()) }
}

func writeGotGotgo(filename string, out *os.File, actualtypes []string) (e os.Error) {
	x, e := ioutil.ReadFile(filename)
	if e != nil { return }
	var scan scanner.Scanner
	scan.Init(filename, x, nil, 0)
	tok := token.COMMA // anything but EOF or PACKAGE
	for tok != token.EOF && tok != token.PACKAGE {
		_, tok, _ = scan.Scan()
	}
	if tok == token.EOF { return os.NewError("Unexpected EOF...") }
	_, tok, gotpname := scan.Scan()
	if tok != token.IDENT {
		return os.NewError("Expected package ident, not "+string(gotpname))
	}
	if *pname == "NAME" {
		*pname = string(gotpname)
	}
	if *prefix == "NAME" {
		*prefix = ""
	}
	_, tok, lit := scan.Scan()
	if tok != token.LPAREN {
		return os.NewError("Expected (, not "+string(lit))
	}
	params, types, restpos, e := getTypes(&scan)
	if e != nil { return }
	vartypes := make(map[string]string)
	imports := []string{}
	for i,t := range types {
		if i < len(actualtypes) {
			if dot := strings.LastIndex(actualtypes[i], "."); dot != -1 {
				// We've got to add an import!
				imports = stringslice.Append(imports,
					"import " + params[i] + ` "` + actualtypes[i][0:dot]+`"`)
				vartypes[params[i]] = params[i]+actualtypes[i][dot:]
			} else {
				vartypes[params[i]] = actualtypes[i]
			}
		} else {
			if t == params[0] {
				vartypes[params[i]] = vartypes[params[0]]
			} else {
				vartypes[params[i]] = t
			}
		}
	}
	lastpos := restpos.Offset+1
	// Now let's write the package file...
	fmt.Fprintf(out, "package %s\n\n", *pname)
	for _,imp := range imports {
		// These are extra imports for data types...
		fmt.Fprintf(out, "%s\n", imp)
	}
	pos,tok,lit := scan.Scan();
	for tok != token.EOF {
		if t,ok := vartypes[string(lit)]; ok {
			fmt.Fprint(out, string(x[lastpos:pos.Offset]))
			fmt.Fprint(out, t)
			lastpos = pos.Offset + len(lit)
		}
		newpos,newtok,newlit := scan.Scan()
		if string(lit) == string(gotpname) && newtok == token.PERIOD {
			fmt.Fprint(out, string(x[lastpos:pos.Offset]))
			fmt.Fprint(out, *prefix)
			lastpos = newpos.Offset + len(newlit)
			pos,tok,lit = scan.Scan()
		} else {
			pos, tok, lit = newpos, newtok, newlit
		}
	}
	fmt.Fprintf(out, string(x[lastpos:]))

	fmt.Fprintf(out,`
// Here we will test that the types parameters are ok...
func %stestTypes(arg0 %s`, *prefix, vartypes[params[0]])
	for i,p := range params[1:] {
		t := vartypes[p]
		if t == params[0] { t = vartypes[params[0]] }
		fmt.Fprintf(out, `, arg%d %s`, i+1, t)
	}
	fmt.Fprintf(out, `) {
    f := func(%s`, types[0])
	for _,t := range types[1:] {
		if t == params[0] { t = types[0] }
		fmt.Fprintf(out, `, %s`, t)
	}
  fmt.Fprint(out, ") { } // this func does nothing...")
	convert := func(t string, argnum int) string {
		arg := "arg" + fmt.Sprint(argnum)
		if strings.Index(t, "{") == -1 {
			return t+"("+arg+")"
		}
		return arg  // it's an interface, so we needn't convert
	}
	fmt.Fprintf(out, `
    f(%s`, convert(types[0], 0))
	for i,t := range types[1:] {
		if t == params[0] { t = types[0] }
		fmt.Fprintf(out, `, %s`, convert(t, i+1))
	}
	fmt.Fprint(out, `)
}
`)
	return
}


func getTypes(s *scanner.Scanner) (params []string, types []string, pos token.Position, error os.Error) {
	tok := token.COMMA
	var lit []byte
	for tok == token.COMMA {
		pos, tok,lit = s.Scan()
		if tok != token.TYPE {
			error = os.NewError("Expected 'type', not "+string(lit))
			return
		}
		var tname string
		var par []byte
		pos, tok,par = s.Scan()
		if tok != token.IDENT {
			error = os.NewError("Identifier expected, not "+string(par))
			return
		}
		tname,pos,tok,lit = getType(s)
		params = stringslice.Append(params, string(par))
		types = stringslice.Append(types, string(tname))
	}
	if tok != token.RPAREN {
		error = os.NewError(fmt.Sprintf("inappropriate token %v with lit: %s",
			tok, lit))
	}
	return
}

func getType(s *scanner.Scanner) (t string, pos token.Position, tok token.Token, lit []byte)  {
	pos, tok, lit = s.Scan()
	for tok != token.RPAREN && tok != token.COMMA {
		t += string(lit)
		pos, tok, lit = s.Scan()
	}
	if t == "" { t = "interface{}" }
	return
}
