package main

import (
	"fmt"
	"go/types"
	"os"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

type Generator struct {
	outputPackage                        string
	loadedPackages                       map[string]*packages.Package
	module                               string
	imports                              map[string]string
	baseImports                          SetString
	externalImports                      SetString
	localImports                         SetString
	thingsToGenerateSortedKeys           []string
	thingsToGenerateInterfacesSortedKeys map[string][]string
	thingsToGenerate                     map[string]map[string]string
}

func NewGenerator(outputPackage string, module string, pkgs []*packages.Package, thingsToGenerate map[string]map[string]string) *Generator {
	g := &Generator{
		outputPackage:                        outputPackage,
		loadedPackages:                       make(map[string]*packages.Package),
		module:                               module,
		imports:                              make(map[string]string),
		thingsToGenerateInterfacesSortedKeys: make(map[string][]string),
		baseImports:                          make(SetString),
		externalImports:                      make(SetString),
		localImports:                         make(SetString),
		thingsToGenerate:                     thingsToGenerate,
	}

	for _, pkg := range pkgs {
		g.loadedPackages[pkg.PkgPath] = pkg
	}

	// this is dirty but it's getting late.
	for k, v1 := range g.thingsToGenerate {
		g.thingsToGenerateSortedKeys = append(g.thingsToGenerateSortedKeys, k)
		for v2, _ := range v1 {
			g.thingsToGenerateInterfacesSortedKeys[k] = append(g.thingsToGenerateInterfacesSortedKeys[k], v2)
		}
		sort.Strings(g.thingsToGenerateInterfacesSortedKeys[k])
	}
	sort.Strings(g.thingsToGenerateSortedKeys)

	g.addImport("testing", "testing")
	g.addImport("github.com/stretchr/testify/assert", "assert")

	for packageName, interfaceFromTo := range g.thingsToGenerate {
		for sourceInterfaceName, _ := range interfaceFromTo {
			ifaceDef := g.FindInterfaceTypeInPackages(packageName, sourceInterfaceName)
			if ifaceDef == nil {
				_, _ = fmt.Fprintf(os.Stderr, "unable to find %s.%s\n", packageName, sourceInterfaceName)
				_, _ = fmt.Fprintf(os.Stderr, "names: %s\n", g.loadedPackages[packageName].Types.Scope().Names())
				os.Exit(1)
			}
			for _, methodDef := range ifaceDef.Methods {
				for i := 0; i < len(methodDef.Params); i++ {
					g.collectImport(methodDef.Params[i].Type())
				}
				for i := 0; i < len(methodDef.Results); i++ {
					g.collectImport(methodDef.Results[i].Type())
				}
			}
		}
	}

	return g
}

func (g *Generator) typeToString(pType types.Type) string {
	switch pType := pType.(type) {
	case *types.Array:
		return fmt.Sprintf("[%d]%s", pType.Len(), g.typeToString(pType.Elem()))
	case *types.Basic:
		return pType.String()
	case *types.Chan:
		if pType.Dir() == 1 {
			return fmt.Sprintf("chan<- %s", g.typeToString(pType.Elem()))
		} else if pType.Dir() == 2 {
			return fmt.Sprintf("<-chan %s", g.typeToString(pType.Elem()))
		} else {
			return fmt.Sprintf("chan %s", g.typeToString(pType.Elem()))
		}
	case *types.Interface:
		if !pType.Empty() {
			panic("not supported")
		}
		return "interface{}"
	case *types.Named:
		obj := pType.Obj()
		objPkg := obj.Pkg()
		objName := obj.Name()
		if objPkg == nil {
			return objName
		}
		return fmt.Sprintf("%s.%s", g.imports[objPkg.Path()], objName)
	case *types.Map:
		return fmt.Sprintf("map[%s]%s", g.typeToString(pType.Key()), g.typeToString(pType.Elem()))
	case *types.Pointer:
		return fmt.Sprintf("*%s", g.typeToString(pType.Elem()))
	case *types.Signature:
		return fmt.Sprintf("func%s", g.RenderParamResults(NewFunc("", pType)))
	case *types.Slice:
		return fmt.Sprintf("[]%s", g.typeToString(pType.Elem()))
	default:
		fmt.Printf("%#v\n", pType)
		return "derp"
	}
}

func (g *Generator) RenderStructField(fn *FuncWrapper, maxLength int) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Fn%-*s func", maxLength, fn.Name))

	sb.WriteString(g.RenderParamResults(fn))
	return sb.String()
}

func (g *Generator) RenderParamResults(fn *FuncWrapper) string {
	var sb strings.Builder
	sb.WriteString("(")
	for idx, p := range fn.Params {
		if idx > 0 {
			sb.WriteString(", ")
		}

		if p.Name() != "" {
			sb.WriteString(p.Name())
			sb.WriteByte(' ')
		}
		pType := p.Type()
		if fn.Variadic && idx == len(fn.Params)-1 {
			sb.WriteString("...")
			pType = pType.(*types.Slice).Elem()
		}
		sb.WriteString(g.typeToString(pType))
	}

	sb.WriteString(")")
	switch len(fn.Results) {
	case 0:
	case 1:
		sb.WriteByte(' ')
		sb.WriteString(g.typeToString(fn.Results[0].Type()))
	default:
		sb.WriteString(" (")
		for idx, r := range fn.Results {
			if idx > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(g.typeToString(r.Type()))
		}
		sb.WriteString(")")
	}
	return sb.String()
}

func (g *Generator) RenderFuncParams(fn *FuncWrapper) string {
	var sb strings.Builder

	for idx, p := range fn.Params {
		if idx > 0 {
			sb.WriteString(", ")
		}

		name := p.Name()
		if name == "" {
			name = fmt.Sprintf("a%d", idx)
		}
		sb.WriteString(fmt.Sprintf("%s ", name))
		pType := p.Type()
		if fn.Variadic && idx == len(fn.Params)-1 {
			sb.WriteString("...")
			pType = pType.(*types.Slice).Elem()
		}
		sb.WriteString(g.typeToString(pType))
	}
	return sb.String()
}

func (g *Generator) RenderFuncInvokeParams(fn *FuncWrapper) string {
	var sb strings.Builder

	for idx, p := range fn.Params {
		if idx > 0 {
			sb.WriteString(", ")
		}

		name := p.Name()
		if name == "" {
			name = fmt.Sprintf("a%d", idx)
		}
		sb.WriteString(name)
		if fn.Variadic && idx == len(fn.Params)-1 {
			sb.WriteString("...")
		}
	}
	return sb.String()
}

func (g *Generator) RenderFuncResults(fn *FuncWrapper) string {
	var sb strings.Builder

	if len(fn.Results) == 0 {
		return ""
	}

	sb.WriteString(" (")
	for idx, r := range fn.Results {
		if idx > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("p%d %s", idx, g.typeToString(r.Type())))
	}
	sb.WriteString(")")
	return sb.String()
}

func (g *Generator) FindInterfaceTypeInPackages(packagePath string, interfaceName string) *IfaceWrapper {
	pkg, ok := g.loadedPackages[packagePath]
	if !ok {
		return nil
	}
	pkgScope := pkg.Types.Scope()
	nameType, ok := pkgScope.Lookup(interfaceName).(*types.TypeName)
	if !ok {
		return nil
	}

	namedType, ok := nameType.Type().(*types.Named)
	if !ok {
		return nil
	}

	ifaceType, ok := namedType.Underlying().(*types.Interface)
	if !ok {
		return nil
	}
	return NewInterface(ifaceType)
}

func (g *Generator) collectImport(t types.Type) {
	switch paramType := t.(type) {
	case *types.Array:
		g.collectImport(paramType.Elem())
	case *types.Basic:
	case *types.Chan:
		g.collectImport(paramType.Elem())
	case *types.Interface:
		if !paramType.Empty() {
			panic("not supported")
		}
	case *types.Map:
		g.collectImport(paramType.Key())
		g.collectImport(paramType.Elem())
	case *types.Named:
		if paramType.Obj().Pkg() != nil {
			pkg := paramType.Obj().Pkg()
			g.addImport(string(pkg.Path()), pkg.Name())
		} else if paramType.Obj().Name() != "error" {
			fmt.Printf("/*UNEXPECTED\n")
			fmt.Printf("%#v\n", paramType)
			fmt.Printf("-----\n")
			fmt.Printf("*/\n")
		}
	case *types.Pointer:
		g.collectImport(paramType.Elem())
	case *types.Signature:
		for i := 0; i < paramType.Params().Len(); i++ {
			g.collectImport(paramType.Params().At(i).Type())
		}
		for i := 0; i < paramType.Results().Len(); i++ {
			g.collectImport(paramType.Results().At(i).Type())
		}
	case *types.Slice:
		g.collectImport(paramType.Elem())
	default:
		fmt.Printf("%#v\n", paramType)
		panic("unhandled element type")
	}
}

func (g *Generator) addImport(pkgPath string, alias string) {
	g.imports[pkgPath] = alias
	parts := strings.Split(pkgPath, "/")
	if strings.HasPrefix(pkgPath, g.module) {
		g.localImports.Add(pkgPath)
	} else if strings.Contains(parts[0], ".") {
		g.externalImports.Add(pkgPath)
	} else {
		g.baseImports.Add(pkgPath)
	}
}

func (g *Generator) RenderImports(imports []string) string {
	var sb fmtBuilder
	for _, importPath := range imports {
		importAlias := g.imports[string(importPath)]
		if importAlias == "" || importAlias == importPath || strings.HasSuffix(importPath, "/"+importAlias) {
			_, _ = sb.WriteStringf("\t\"%s\"\n", importPath)
		} else {
			_, _ = sb.WriteStringf("\t%s \"%s\"\n", importAlias, importPath)
		}
	}
	return sb.String()
}

func (g *Generator) RenderBody(mockName string, methodDef *FuncWrapper) string {
	var sb fmtBuilder
	_, _ = sb.WriteStringf("func (m *Mock%s) %s(%s)%s {\n", mockName, methodDef.Name, g.RenderFuncParams(methodDef), g.RenderFuncResults(methodDef))
	_, _ = sb.WriteStringf("\tif m.Fn%s != nil {\n", methodDef.Name)
	if len(methodDef.Results) == 0 {
		_, _ = sb.WriteStringf("\t\tm.Fn%s(%s)\n", methodDef.Name, g.RenderFuncInvokeParams(methodDef))
		_, _ = sb.WriteStringf("\t} else {\n")
		_, _ = sb.WriteStringf("\t\tassert.Fail(m.TB, \"%s.%s must not be called\")\n", mockName, methodDef.Name)
		_, _ = sb.WriteStringf("\t}\n")
	} else {
		_, _ = sb.WriteStringf("\t\treturn m.Fn%s(%s)\n", methodDef.Name, g.RenderFuncInvokeParams(methodDef))
		_, _ = sb.WriteStringf("\t}\n")
		_, _ = sb.WriteStringf("\tassert.Fail(m.TB, \"%s.%s must not be called\")\n", mockName, methodDef.Name)
		_, _ = sb.WriteStringf("\treturn\n")
	}
	_, _ = sb.WriteStringf("}\n")
	return sb.String()
}

func (g *Generator) Generate() string {
	var sb fmtBuilder

	_, _ = sb.WriteStringf("package %s\n", g.outputPackage)
	_, _ = sb.WriteString("\n")

	_, _ = sb.WriteString("// Code generated by github.com/squizzling/mockgen. DO NOT EDIT.\n")
	_, _ = sb.WriteString("\n")

	_, _ = sb.WriteStringf("import (\n")
	_, _ = sb.WriteString(g.RenderImports(g.baseImports.Sorted()))
	_, _ = sb.WriteStringf("\n")
	_, _ = sb.WriteString(g.RenderImports(g.externalImports.Sorted()))
	if len(g.localImports) > 0 {
		_, _ = sb.WriteStringf("\n")
		_, _ = sb.WriteString(g.RenderImports(g.localImports.Sorted()))
	}
	_, _ = sb.WriteStringf(")\n")

	for _, packageName := range g.thingsToGenerateSortedKeys {
		interfaceNames := g.thingsToGenerate[packageName]
		interfaceNamesSorted := g.thingsToGenerateInterfacesSortedKeys[packageName]

		for _, sourceInterfaceName := range interfaceNamesSorted {
			generatedInterfaceName := interfaceNames[sourceInterfaceName]
			_, _ = sb.WriteStringf("\n")
			_, _ = sb.WriteStringf("// Mock%s implements a mock %s.%s from %s\n", generatedInterfaceName, g.loadedPackages[packageName].Name, sourceInterfaceName, packageName)
			_, _ = sb.WriteStringf("type Mock%s struct {\n", generatedInterfaceName)
			_, _ = sb.WriteStringf("\tTB testing.TB\n")
			_, _ = sb.WriteStringf("\n")
			ifaceDef := g.FindInterfaceTypeInPackages(packageName, sourceInterfaceName)
			for _, methodDef := range ifaceDef.Methods {
				_, _ = sb.WriteStringf("\t%s\n", g.RenderStructField(methodDef, ifaceDef.LongestMethodName))
			}
			_, _ = sb.WriteStringf("}\n")

			for _, methodDef := range ifaceDef.Methods {
				_, _ = sb.WriteStringf("\n")
				_, _ = sb.WriteString(g.RenderBody(generatedInterfaceName, methodDef))
			}
		}
	}

	return sb.String()
}
