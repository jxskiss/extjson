package extjson

import (
	"encoding/json"
	"os"

	"github.com/jxskiss/extjson/internal/parser"
)

//go:generate peg -output ./internal/parser/json.peg.go json.peg

// Unmarshal parses the JSON-encoded data and stores the result in the
// value pointed to by v.
//
// In addition to features of encoding/json, it enables extended features
// such as "trailing comma", "comments", "file including", "refer", etc.
// The extended features are documented in the README file.
func Unmarshal(data []byte, v interface{}, options ...ExtOption) error {
	opt := new(extOptions).apply(options...)
	includeRoot, err := opt.getIncludeRoot()
	if err != nil {
		return err
	}
	if err = opt.validateFuncs(); err != nil {
		return err
	}
	data, err = parser.Parse(data, includeRoot, opt.EnableEnv, opt.FuncMap)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// Clean parses data with extended feature and returns it as normal
// spec-compliant JSON data.
func Clean(data []byte, options ...ExtOption) ([]byte, error) {
	var raw json.RawMessage
	err := Unmarshal(data, &raw, options...)
	if err != nil {
		return nil, err
	}
	return raw, nil
}

// Load reads JSON-encoded data from the named file at path and stores
// the result in the value pointed to by v.
//
// In additional to features of encoding/json, it enables extended features
// such as "trailing comma", "comments", "file including", "refer" etc.
// The extended features are documented in the README file.
func Load(path string, v interface{}, options ...ExtOption) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return Unmarshal(data, v, options...)
}

// Dump writes v to the named file at path using JSON encoding.
// It disables HTMLEscape.
// Optionally indent can be applied to the output, empty prefix and
// indent disables indentation.
// The output is friendly to read by humans.
func Dump(path string, v interface{}, prefix, indent string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	enc := json.NewEncoder(file)
	enc.SetEscapeHTML(false)
	enc.SetIndent(prefix, indent)
	err = enc.Encode(v)
	return err
}
