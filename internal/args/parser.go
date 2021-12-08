package args

import (
	"fmt"
	"os"
	"reflect"

	"github.com/jessevdk/go-flags"
)

type Validater interface {
	Validate() []string
}

var typeValidater = reflect.TypeOf((*Validater)(nil)).Elem()

func ParseArgs(commandLine []string, opts interface{}) {
	parser := flags.NewParser(opts, flags.HelpFlag|flags.PassDoubleDash)
	parser.LongDescription = ``

	// Main arg parsing
	positional, err := parser.ParseArgs(commandLine)
	if err != nil {
		if !IsHelp(err) {
			parser.WriteHelp(os.Stderr)
			_, _ = fmt.Fprintf(os.Stderr, "\n\nerror parsing command line: %v\n", err)
			os.Exit(1)
		}
		parser.WriteHelp(os.Stdout)
		os.Exit(0)
	}

	// Set Positional field (if present) and validate arguments
	errors := validate(opts, positional)

	// Validation results
	if len(errors) > 0 {
		parser.WriteHelp(os.Stderr)
		_, _ = fmt.Fprintf(os.Stderr, "\n\n")
		for _, err := range errors {
			_, _ = fmt.Fprintf(os.Stderr, "error parsing command line: %s\n", err)
		}
		os.Exit(1)
	}
}

func validate(opts interface{}, positional []string) []string {
	v := reflect.ValueOf(opts).Elem()

	var errors []string
	if len(positional) > 0 {
		errors = setPositional(v, positional)
	}

	if validater := tryFindValidater(v); validater != nil {
		errors = append(errors, validater.Validate()...)
	}
	for i := 0; i < v.NumField(); i++ {
		if validater := tryFindValidater(v.Field(i)); validater != nil {
			errors = append(errors, validater.Validate()...)
		}
	}
	return errors
}

func setPositional(v reflect.Value, positional []string) []string {
	if pField := v.FieldByName("Positional"); pField.IsValid() && pField.CanInterface() {
		if _, ok := pField.Interface().([]string); ok {
			pField.Set(reflect.ValueOf(positional))
			return nil
		} else {
			return []string{"Positional field is wrong type"}
		}
	} else {
		return []string{"positional arguments are not allowed"}
	}
}

func tryFindValidater(v reflect.Value) Validater {
	if !v.Type().AssignableTo(typeValidater) || !v.CanInterface() {
		if !v.CanAddr() {
			return nil
		}
		v = v.Addr()
		if !v.Type().AssignableTo(typeValidater) || !v.CanInterface() {
			return nil
		}
	}
	validater, _ := v.Interface().(Validater) // use non-panicking form, just in case
	return validater
}
