package main

import (
	"fmt"
	"strings"
)

type fmtBuilder struct {
	strings.Builder
}

func (f *fmtBuilder) WriteStringf(msg string, args ...interface{}) (int, error) {
	return f.WriteString(fmt.Sprintf(msg, args...))
}
