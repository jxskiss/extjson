package extjson

import (
	"fmt"
	"os"
	"reflect"
	"unicode"
)

// EnableEnv enables reading environment variables.
// By default, it is disabled for security consideration.
func EnableEnv() ExtOption {
	return ExtOption{
		apply: func(options *extOptions) {
			options.EnableEnv = true
		}}
}

// IncludeRoot specifies the root directory to use with the extended file
// including feature.
func IncludeRoot(dir string) ExtOption {
	return ExtOption{
		apply: func(options *extOptions) {
			options.IncludeRoot = dir
		}}
}

// FuncMap is the type of the map defining the mapping from names to functions.
// Each function must have either a single return value, or two return values of
// which the second has type error. In that case, if the second (error)
// return value evaluates to non-nil during execution, execution terminates and
// the error will be returned.
type FuncMap map[string]interface{}

// WithFuncMap specifies additional functions to use with the "@fn" directive.
func WithFuncMap(funcMap FuncMap) ExtOption {
	return ExtOption{
		apply: func(options *extOptions) {
			options.FuncMap = funcMap
		},
	}
}

// ExtOption represents an option to customize the extended features.
type ExtOption struct {
	apply func(options *extOptions)
}

type extOptions struct {
	EnableEnv   bool
	IncludeRoot string
	FuncMap     FuncMap
}

func (o *extOptions) apply(opts ...ExtOption) *extOptions {
	for _, opt := range opts {
		opt.apply(o)
	}
	return o
}

func (o *extOptions) getIncludeRoot() (string, error) {
	if o.IncludeRoot != "" {
		return o.IncludeRoot, nil
	}
	return os.Getwd()
}

func (o *extOptions) validateFuncs() error {
	for name, fn := range o.FuncMap {
		if !goodName(name) {
			return fmt.Errorf("function name %q is not a valid identifier", name)
		}
		typ := reflect.TypeOf(fn)
		if typ.Kind() != reflect.Func {
			return fmt.Errorf("value for %q is not a function", name)
		}
		if !goodFunc(typ) {
			return fmt.Errorf("function %q has invalid signature", name)
		}
	}
	return nil
}

var errorType = reflect.TypeOf((*error)(nil)).Elem()

// goodFunc reports whether the function or method has the right result signature.
func goodFunc(typ reflect.Type) bool {
	// We allow functions with 1 result or 2 results where the second is an error.
	switch {
	case typ.NumOut() == 1:
		return true
	case typ.NumOut() == 2 && typ.Out(1) == errorType:
		return true
	}
	return false
}

// goodName reports whether the function name is a valid identifier.
func goodName(name string) bool {
	if name == "" {
		return false
	}
	for i, r := range name {
		switch {
		case r == '_':
		case i == 0 && !unicode.IsLetter(r):
			return false
		case !unicode.IsLetter(r) && !unicode.IsDigit(r):
			return false
		}
	}
	return true
}
