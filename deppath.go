// The deppath command prints out all paths between two packages.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

func usage() {
	fmt.Printf(`Usage: deppath <src-pkg> <dst-pkg>

If dst-pkg starts with a slash, it is interpreted as a substring to
match rather than an exact package name.
`)
	os.Exit(1)
}

func main() {
	flag.Usage = usage
	flag.Parse()
	if flag.NArg() != 2 {
		usage()
	}
	from, to := flag.Arg(0), flag.Arg(1)

	cfg := &packages.Config{
		Mode: packages.NeedImports,
	}
	pkgs, err := packages.Load(cfg, from)
	if err != nil {
		log.Fatalf("Loading package %q: %v", from, err)
	}
	if len(pkgs) != 1 {
		log.Fatalf("Loading package %q: got %d packages, want 1", from, len(pkgs))
	}
	var pkgMap = map[string]*packages.Package{}
	var visitPkg func(*packages.Package)
	visitPkg = func(pkg *packages.Package) {
		if _, ok := pkgMap[pkg.ID]; ok {
			return
		}
		pkgMap[pkg.ID] = pkg
		for _, ipkg := range pkg.Imports {
			visitPkg(ipkg)
		}
	}
	for _, pkg := range pkgs[0].Imports {
		visitPkg(pkg)
	}
	pkgMap[from] = pkgs[0]

	var matches []string
	var stack []string
	var uselessRoot = map[string]bool{} // import path => true
	var visit func(string)
	visit = func(pkgName string) {
		if uselessRoot[pkgName] {
			return
		}
		matches0 := len(matches)
		stack = append(stack, pkgName)
		defer func() {
			stack = stack[:len(stack)-1]
			if matches0 == len(matches) {
				uselessRoot[pkgName] = true
			}
		}()

		if pkgName == to || (strings.HasPrefix(to, "/") && strings.Contains(pkgName, to[1:])) {
			matches = append(matches, fmt.Sprintf("%q", stack))
			return
		}
		pkg, ok := pkgMap[pkgName]
		if !ok {
			return
		}
		for next := range pkg.Imports {
			visit(next)
		}
	}
	visit(from)
	sort.Strings(matches)
	for _, s := range matches {
		fmt.Println(s)
	}
}
