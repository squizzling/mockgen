package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"reflect"
	"strings"
)

func nameTypesOfFieldNames(fields *ast.FieldList) string {
	var fieldTypeNames []string
	EachField(fields, func(fieldName string, fieldType string, isEllipsis bool) {
		fieldTypeNames = append(fieldTypeNames, fieldName+" "+fieldType)
	})
	return strings.Join(fieldTypeNames, ", ")
}

func typesOfFieldNames(fields *ast.FieldList) string {
	var fieldTypes []string
	EachField(fields, func(fieldName string, fieldType string, isEllipsis bool) {
		fieldTypes = append(fieldTypes, fieldType)
	})
	return strings.Join(fieldTypes, ", ")
}

func ToTypeName(e ast.Expr) string {
	switch t := e.(type) {
	case *ast.ArrayType:
		if t.Len == nil {
			return "[]" + ToTypeName(t.Elt)
		}
		panic("only basic slices are supported")
	case *ast.Ellipsis:
		return "..." + ToTypeName(t.Elt)
	case *ast.Ident:
		return t.Name
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", ToTypeName(t.Key), ToTypeName(t.Value))
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s.%s", ToTypeName(t.X), t.Sel.Name)
	case *ast.StarExpr:
		return "*" + ToTypeName(t.X)
	case *ast.FuncType:
		switch n := t.Results.NumFields(); {
		case n == 0:
			return fmt.Sprintf("func(%s)", nameTypesOfFieldNames(t.Params))
		case n == 1:
			return fmt.Sprintf("func(%s) %s", nameTypesOfFieldNames(t.Params), typesOfFieldNames(t.Results))
		default:
			return fmt.Sprintf("func(%s) (%s)", nameTypesOfFieldNames(t.Params), typesOfFieldNames(t.Results))
		}
	default:
		panic(fmt.Sprintf("Unrecognized type: %s", reflect.TypeOf(e).String()))
	}
}

type EachFieldFunc func(fieldName string, fieldType string, isEllipsis bool)

func EachField(fields *ast.FieldList, fn EachFieldFunc) {
	if fields == nil {
		return
	}
	for i, a := range fields.List {
		var t *ast.Ellipsis
		var isEllipsis bool
		var typeName string
		if t, isEllipsis = a.Type.(*ast.Ellipsis); isEllipsis {
			typeName = ToTypeName(t.Elt)
		} else {
			typeName = ToTypeName(a.Type)
		}
		if len(a.Names) > 0 {
			for _, fieldName := range a.Names {
				fn(fieldName.Name, typeName, isEllipsis)
			}
		} else {
			fn(fmt.Sprintf("p%d", i), typeName, isEllipsis)
		}
	}
}

func (f *File) collectInterfaces(n ast.Node) bool {
	decl, ok := n.(*ast.GenDecl)
	if !ok || decl.Tok != token.TYPE {
		return true
	}
	for _, spec := range decl.Specs {
		typeSpec := spec.(*ast.TypeSpec)
		var ifaceType *ast.InterfaceType
		if ifaceType, ok = typeSpec.Type.(*ast.InterfaceType); !ok {
			continue
		}
		defIface := &Interface{
			Name: typeSpec.Name.Name,
		}
		f.Interfaces = append(f.Interfaces, defIface)
		for _, field := range ifaceType.Methods.List {
			Assert(len(field.Names) == 1)

			defMember := &Member{
				Name: field.Names[0].Name,
			}
			defIface.Members = append(defIface.Members, defMember)
			declFunc := field.Type.(*ast.FuncType)

			EachField(declFunc.Params, func(fieldName string, typeName string, isEllipsis bool) {
				if isEllipsis {
					defMember.ParamsNameType = append(defMember.ParamsNameType, fmt.Sprintf("%s ...%s", fieldName, typeName))
					defMember.ParamsName = append(defMember.ParamsName, fieldName+"...")
				} else {
					defMember.ParamsNameType = append(defMember.ParamsNameType, fmt.Sprintf("%s %s", fieldName, typeName))
					defMember.ParamsName = append(defMember.ParamsName, fieldName)
				}
			})
			EachField(declFunc.Results, func(fieldName string, typeName string, isEllipsis bool) {
				defMember.ReturnValuesNameType = append(defMember.ReturnValuesNameType, fmt.Sprintf("%s %s", fieldName, typeName))
				defMember.ReturnValuesType = append(defMember.ReturnValuesType, typeName)
			})

			Assert(len(defMember.ParamsName) == declFunc.Params.NumFields())
			Assert(len(defMember.ReturnValuesType) == declFunc.Results.NumFields())
		}
	}
	return true
}
