package main

import (
	"sort"
	"strings"
)

// SetString implements a basic set of strings
type SetString map[string]struct{}

func NewSetString(strings []string) SetString {
	ss := make(SetString)
	for _, s := range strings {
		ss[s] = struct{}{}
	}
	return ss
}

func NewSetStringStr(commaSepStrings string) SetString {
	return NewSetString(strings.Split(commaSepStrings, ","))
}

// ContainsAll indicates if every item in ssOther is present in ss
func (ss SetString) ContainsAll(ssOther SetString) bool {
	count := len(ssOther)
	for v := range ssOther {
		if _, ok := ss[v]; ok {
			count--
		} else {
			return false
		}
	}
	return count == 0
}

func (ss SetString) Add(s string) SetString {
	ss[s] = struct{}{}
	return ss
}

func (ss SetString) Sorted() []string {
	ks := make([]string, 0, len(ss))
	for k, _ := range ss {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}


func (ss SetString) IsEmpty() bool {
	return len(ss) == 0
}

func (ss SetString) JoinComma() string {
	items := make([]string, 0, len(ss))
	for k, _ := range ss {
		items = append(items, k)
	}
	return strings.Join(items, ",")
}

func (ss SetString) JoinCommaP() *string {
	s := ss.JoinComma()
	return &s
}
