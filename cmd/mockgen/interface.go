package main

import (
	"fmt"
	"go/ast"
	"sort"
	"strings"
)

type Member struct {
	Name                 string
	ParamsName           []string
	ParamsNameType       []string
	ReturnValuesType     []string
	ReturnValuesNameType []string
}

type Interface struct {
	Name    string
	Members []*Member
}

func ToReturnString(returnValues []string) string {
	if len(returnValues) == 0 {
		return ""
	}
	return fmt.Sprintf(" (%s)", strings.Join(returnValues, ", "))
}

func ToSimpleReturnString(returnValues []string) string {
	if len(returnValues) == 0 {
		return ""
	}
	if len(returnValues) == 1 {
		return " " + returnValues[0]
	}
	return fmt.Sprintf(" (%s)", strings.Join(returnValues, ", "))
}

func (m *Member) ToImpl(interfaceName string) string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("func (m *Mock%s) %s(%s)%s {\n", interfaceName, m.Name, strings.Join(m.ParamsNameType, ", "), ToReturnString(m.ReturnValuesNameType)))
	sb.WriteString(fmt.Sprintf("\tif m.Fn%s != nil {\n", m.Name))
	if len(m.ReturnValuesType) == 0 {
		sb.WriteString(fmt.Sprintf("\t\tm.Fn%s(%s)\n", m.Name, strings.Join(m.ParamsName, ", ")))
		sb.WriteString(fmt.Sprintf("\t} else {\n"))
		sb.WriteString(fmt.Sprintf("\t\tassert.Fail(m.TB, \"%s.%s must not be called\")\n", interfaceName, m.Name))
		sb.WriteString(fmt.Sprintf("\t}\n"))
	} else {
		sb.WriteString(fmt.Sprintf("\t\treturn m.Fn%s(%s)\n", m.Name, strings.Join(m.ParamsName, " ")))
		sb.WriteString(fmt.Sprintf("\t}\n"))
		sb.WriteString(fmt.Sprintf("\tassert.Fail(m.TB, \"%s.%s must not be called\")\n", interfaceName, m.Name))
		sb.WriteString(fmt.Sprintf("\treturn\n"))
	}
	sb.WriteString(fmt.Sprintf("}\n"))

	return sb.String()

}

func (g *Generator) Generate() {
	for _, pkg := range g.Pkgs {
		for _, file := range pkg.Files {
			ast.Inspect(file.File, file.collectInterfaces)
		}
	}

	sort.Slice(g.Pkgs, func(i, j int) bool {
		return g.Pkgs[i].Name < g.Pkgs[j].Name
	})

	for _, pkg := range g.Pkgs {
		sort.Slice(pkg.Files, func(i, j int) bool {
			return pkg.Files[i].Name < pkg.Files[j].Name
		})
		for _, file := range pkg.Files {
			fmt.Printf("// File: %s\n", file.Name)
			fmt.Printf("\n")

			sort.Slice(file.Interfaces, func(i, j int) bool {
				return file.Interfaces[i].Name < file.Interfaces[j].Name
			})

			for _, iface := range file.Interfaces {
				sort.Slice(iface.Members, func(i, j int) bool {
					return iface.Members[i].Name < iface.Members[j].Name
				})
				fmt.Printf("// Mock%s implements a mock %s from %s\n", iface.Name, iface.Name, file.Pkg.Name)
				fmt.Printf("type Mock%s struct {\n", iface.Name)
				fmt.Printf("\tTB testing.TB\n")
				fmt.Printf("\n")
				for _, member := range iface.Members {
					fmt.Printf("\tFn%s func(%s)%s\n", member.Name, strings.Join(member.ParamsNameType, ", "), ToSimpleReturnString(member.ReturnValuesType))
				}
				fmt.Printf("}\n")
				fmt.Printf("\n")
				for _, member := range iface.Members {
					fmt.Printf("%s\n", member.ToImpl(iface.Name))
				}
			}
		}
	}
}
