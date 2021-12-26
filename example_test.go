package extjson

import (
	"encoding/json"
	"fmt"
	"os"
)

func Example() {
	data := `{
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
	}`

	os.Setenv("SOME_ENV", "some-env-value")

	var clean json.RawMessage
	_ = Unmarshal([]byte(data), &clean, EnableEnv())
	out, _ := json.MarshalIndent(clean, "", "  ")
	fmt.Println(string(out))

	// Output:
	// {
	//   "obj1": {
	//     "foo": "bar"
	//   },
	//   "array": [
	//     1,
	//     2,
	//     3
	//   ],
	//   "include": {
	//     "foo": "bar"
	//   },
	//   "identifier_simple1": 1234,
	//   "$identifierSimple2": "abc",
	//   "obj2": {
	//     "foo": "bar"
	//   },
	//   "py_true": true,
	//   "py_false": false,
	//   "py_none": null,
	//   "test_env": "some-env-value",
	//   "test_ref1": "bar",
	//   "test_ref2": 3,
	//   "test_ref3": 3,
	//   "friends": [
	//     {
	//       "first": "Dale",
	//       "last": "Murphy",
	//       "age": 44,
	//       "nets": [
	//         "ig",
	//         "fb",
	//         "tw"
	//       ]
	//     },
	//     {
	//       "first": "Roger",
	//       "last": "Craig",
	//       "age": 68,
	//       "nets": [
	//         "fb",
	//         "tw"
	//       ]
	//     },
	//     {
	//       "first": "Jane",
	//       "last": "Murphy",
	//       "age": 47,
	//       "nets": [
	//         "ig",
	//         "tw"
	//       ]
	//     }
	//   ],
	//   "test_ref4": [
	//     "Dale",
	//     "Roger",
	//     "Jane"
	//   ]
	// }
}
