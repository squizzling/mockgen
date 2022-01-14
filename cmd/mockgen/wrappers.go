package main

import (
	"go/types"
	"sort"
)

type IfaceWrapper struct {
	LongestMethodName int
	Methods           []*FuncWrapper
}

type FuncWrapper struct {
	Name     string
	Variadic bool
	Params   []*types.Var
	Results  []*types.Var
}

func NewInterface(iface *types.Interface) *IfaceWrapper {
	iw := &IfaceWrapper{}
	for i := 0; i < iface.NumMethods(); i++ {
		fn := NewFunc(iface.Method(i).Name(), iface.Method(i).Type().(*types.Signature))
		iw.Methods = append(iw.Methods, fn)
		iw.LongestMethodName = MaxInt(iw.LongestMethodName, len(fn.Name))
	}
	sort.Slice(iw.Methods, func(i, j int) bool {
		return iw.Methods[i].Name < iw.Methods[j].Name
	})
	return iw
}

func NewFunc(name string, sig *types.Signature) *FuncWrapper {
	params := sig.Params()
	results := sig.Results()
	fw := &FuncWrapper{
		Name:     name,
		Params:   make([]*types.Var, 0, params.Len()),
		Results:  make([]*types.Var, 0, results.Len()),
		Variadic: sig.Variadic(),
	}
	for i := 0; i < params.Len(); i++ {
		fw.Params = append(fw.Params, params.At(i))
	}

	for i := 0; i < results.Len(); i++ {
		fw.Results = append(fw.Results, results.At(i))
	}
	return fw
}
