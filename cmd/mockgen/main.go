package main

import (
	"fmt"
	"os"

	"golang.org/x/tools/go/packages"

	"github.com/squizzling/mockgen/internal/args"
)


func MustLoadPackages(packagesToGenerate map[string]map[string]string) []*packages.Package {
	cfg := &packages.Config{
		Mode:  packages.NeedName | packages.NeedFiles | packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedModule | packages.NeedImports,
		Tests: false,
	}

	pkgNames := make([]string, 0, len(packagesToGenerate))
	for pkgName, _ := range packagesToGenerate {
		pkgNames = append(pkgNames, pkgName)
	}

	pkgs, err := packages.Load(cfg, pkgNames...)
	if err != nil {
		fmt.Printf("failed to load packages: %s\n", err)
		os.Exit(1)
	}

	anyErr := false
	for _, pkg := range pkgs {
		for _, err = range pkg.Errors {
			fmt.Printf("error loading package %s: %s\n", pkg.PkgPath, err)
			anyErr = true
		}
	}

	if anyErr {
		os.Exit(1)
	}

	return pkgs
}

func main() {
	var opts Opts
	opts.Input = parseInput(&opts)
	args.ParseArgs(os.Args[1:], &opts)

	if opts.Chdir != "" {
		if err := os.Chdir(opts.Chdir); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Failed to change to %s: %s\n", opts.Chdir, err)
			os.Exit(1)
		}
	}

	pkgs := MustLoadPackages(opts.pkgs)
	g := NewGenerator(opts.OutputPackage, opts.Module, pkgs, opts.pkgs)

	data := g.Generate()

	if opts.OutputFile == "" {
		_, _ = os.Stdout.Write([]byte(data))
	} else {
		f, err := os.Create(opts.OutputFile)
		if err == nil {
			_, err = f.Write([]byte(data))
			_ = f.Close()
		}
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "failed to write %s: %s\n", opts.OutputFile, err)
		}
	}
}
