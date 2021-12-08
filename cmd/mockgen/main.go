package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"os"

	"golang.org/x/tools/go/packages"

	"github.com/squizzling/mockgen/internal/args"
)

type Generator struct {
	Pkgs []*Package
}

type Package struct {
	Name  string
	Files []*File
}

type File struct {
	Name       string
	Pkg        *Package
	File       *ast.File
	Interfaces []*Interface
}

func main() {
	var opts Opts

	args.ParseArgs(os.Args[1:], &opts)

	cfg := &packages.Config{
		Mode:  packages.NeedName | packages.NeedFiles | packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedModule,
		Tests: false,
		Fset:  token.NewFileSet(),
	}

	pkgs, err := packages.Load(cfg, opts.Package...)
	if err != nil {
		panic(err)
	}
	for _, pkg := range pkgs {
		for _, err = range pkg.Errors {
			fmt.Printf("err: %#v\n", err)
		}
	}
	if err != nil {
		panic(err)
	}

	g := &Generator{}
	for _, pkg := range pkgs {
		g.AddPackage(cfg.Fset, pkg)
	}
	g.Generate()
}

func (g *Generator) AddPackage(fset *token.FileSet, pkg *packages.Package) {
	p := &Package{
		Name: pkg.PkgPath,
	}
	g.Pkgs = append(g.Pkgs, p)

	for _, file := range pkg.Syntax {
		p.Files = append(p.Files, &File{
			Name: fset.File(file.Package).Name(),
			Pkg:  p,
			File: file,
		})
	}
}

func Assert(b bool) {
	if !b {
		panic("")
	}
}
