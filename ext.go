package extjson

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/jxskiss/extjson/parser"
)

//go:generate peg -output ./parser/json.peg.go json.peg

// UnmarshalExt parses the JSON-encoded data and stores the result in the
// value pointed to by v.
//
// In addition to features of encoding/json, it enables some extended
// features such as "trailing comma", "comments", "file including", etc.
// The extended features are documented in the README file.
func UnmarshalExt(data []byte, v interface{}, options ...ExtOption) error {
	opt := new(extOptions).apply(options...)
	includeRoot, err := opt.getIncludeRoot()
	if err != nil {
		return err
	}
	data, err = parser.Parse(data, includeRoot)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// LoadExt reads JSON-encoded data from the named file at path and stores
// the result in the value pointed to by v.
//
// In additional to features of encoding/json, it enables some extended
// features such as "trailing comma", "comments", "file including" etc.
// The extended features are documented in the README file.
func LoadExt(path string, v interface{}, options ...ExtOption) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}
	return UnmarshalExt(data, v, options...)
}

// IncludeRoot specifies the root directory to use with the extended file
// including feature.
func IncludeRoot(dir string) ExtOption {
	return ExtOption{apply: func(options *extOptions) {
		options.IncludeRoot = dir
	}}
}

// ExtOption represents an option to customize the extended features.
type ExtOption struct {
	apply func(options *extOptions)
}

type extOptions struct {
	IncludeRoot string
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
