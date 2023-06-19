package main

import (
	"errors"
	"strings"
)

type packageNameDef string

type Opts struct {
	Chdir         string             `short:"C" long:"chdir" description:"Directory to run from"`
	Module        string             `short:"m" long:"module"`
	OutputFile    string             `short:"f" long:"file"    required:"true"`
	OutputPackage string             `short:"p" long:"output-package" required:"true"`
	Input         func(string) error `short:"i" long:"input" description:"Format: <importpath>:<interface>[=struct][,<interface>[=struct]]..."`
	pkgs          map[string]map[string]string
}

func parseInput(opts *Opts) func(string) error {
	return func(s string) error {
		parts := strings.SplitN(s, ":", 2)
		if len(parts) != 2 {
			return errors.New("no : found in package name")
		}

		pkgPart, interfacePart := parts[0], parts[1]
		if opts.pkgs == nil {
			opts.pkgs = make(map[string]map[string]string)
		}

		if _, ok := opts.pkgs[pkgPart]; !ok {
			opts.pkgs[pkgPart] = make(map[string]string)
		}
		for _, ifaceDefs := range strings.Split(interfacePart, ",") {
			parts = strings.SplitN(ifaceDefs, "=", 2)
			opts.pkgs[pkgPart][parts[0]] = parts[len(parts)-1] // Use the last, which will either be the first or the second
		}
		return nil
	}
}

func (o *Opts) Validate() []string {
	return nil
}
