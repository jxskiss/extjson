package extjson

import (
	"os"
	"reflect"
	"testing"
)

var malformedJSONData = `
{
	// A comment! You normally can't put these in JSON
	"obj1": {
		"foo": "bar", // <-- A trailing comma! No worries.
	},
	/*
	This style of comments will also be safely removed.
	*/
	"array": [1, 2, 3, ], // Trailing comma in array.
	"include": @incl("testdata.json"), // Include another json file.
	identifier_simple1: 1234,
	$identifierSimple2: "abc",
	"obj2": {
		"foo": "bar", /* Another style inline comment. */
	}, // <-- Another trailing comma!
	'py_true': True, // Single quote string and True as true value.
	py_false: False, // Simple identifier and Python False as false value.
	py_none: None,   /* Simple identifier and Python None as null value. */
	"test_env": @env("SOME_ENV"),  // Read environment variable.
	"test_ref1": @ref("obj1.foo"), // Reference to other values, wil be "bar".
	"test_ref2": @ref("array.2"),  // Another reference, will be 3.
	"test_ref3": @ref("array.#"),  // Get length of "array", will be 3.
	"friends": [
		{"first": "Dale", "last": "Murphy", "age": 44, "nets": ["ig", "fb", "tw"]},
		{"first": "Roger", "last": "Craig", "age": 68, "nets": ["fb", "tw"]},
		{"first": "Jane", "last": "Murphy", "age": 47, "nets": ["ig", "tw"]}
	],
	"test_ref4": @ref("friends.#.first"), // Will be ["Dale","Roger","Jane"].
}
`

func TestUnmarshal(t *testing.T) {
	want := map[string]interface{}{
		"obj1": map[string]interface{}{
			"foo": "bar",
		},
		"array": []interface{}{float64(1), float64(2), float64(3)},
		"include": map[string]interface{}{
			"foo": "bar",
		},
		"identifier_simple1": float64(1234),
		"$identifierSimple2": "abc",
		"obj2": map[string]interface{}{
			"foo": "bar",
		},
		"py_true":   true,
		"py_false":  false,
		"py_none":   nil,
		"test_env":  "some-env-value",
		"test_ref1": "bar",
		"test_ref2": float64(3),
		"test_ref3": float64(3),
		"friends": []interface{}{
			map[string]interface{}{"first": "Dale", "last": "Murphy", "age": float64(44), "nets": []interface{}{"ig", "fb", "tw"}},
			map[string]interface{}{"first": "Roger", "last": "Craig", "age": float64(68), "nets": []interface{}{"fb", "tw"}},
			map[string]interface{}{"first": "Jane", "last": "Murphy", "age": float64(47), "nets": []interface{}{"ig", "tw"}},
		},
		"test_ref4": []interface{}{"Dale", "Roger", "Jane"},
	}

	os.Setenv("SOME_ENV", "some-env-value")
	got := make(map[string]interface{})
	err := Unmarshal([]byte(malformedJSONData), &got, EnableEnv())
	if err != nil {
		t.Fatalf("failed unmarshal malformed json: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expecting equal: got = %v, want = %v", got, want)
	}
}

func TestUnmarshal_UnicodeEscape(t *testing.T) {
	jsonData := `["Grammar \u0026 Mechanics \/ Word Work"]`
	got := make([]string, 0)
	err := Unmarshal([]byte(jsonData), &got)
	if err != nil {
		t.Errorf("failed unmarshal unicode escape char: %v", err)
	}
}

func TestUnmarshal_SingleQuote(t *testing.T) {
	jsonData := `{'key\'': 'value"'}`
	got := make(map[string]string)
	err := Unmarshal([]byte(jsonData), &got)
	if err != nil {
		t.Errorf("failed unmarshal single quoted string: %v", err)
	}
	if got["key'"] != "value\"" {
		t.Errorf("unmarshal single quoted string: incorrect key value")
	}
}
