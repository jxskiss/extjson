package extjson

import (
	"encoding/json"
	"os"
)

// Load reads JSON-encoded data from the named file at path and stores
// the result in the value pointed to by v.
func Load(path string, v interface{}) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	err = json.NewDecoder(file).Decode(v)
	return err
}

// Dump writes v to the named file at path using JSON encoding.
// It disables HTMLEscape.
// Optionally indent can be applied to the output,
// empty prefix and indent disables indentation.
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
