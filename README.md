# extjson

[![GoDoc](https://img.shields.io/badge/api-Godoc-blue.svg)][godoc]
[![Go Report Card](https://goreportcard.com/badge/github.com/jxskiss/extjson)][goreport]
[![Issues](https://img.shields.io/github/issues/jxskiss/extjson.svg)][issues]
[![GitHub release](http://img.shields.io/github/release/jxskiss/extjson.svg)][release]
[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg)][license]

[godoc]: https://pkg.go.dev/github.com/jxskiss/extjson
[goreport]: https://goreportcard.com/report/github.com/jxskiss/extjson
[issues]: https://github.com/jxskiss/extjson/issues
[release]: https://github.com/jxskiss/extjson/releases
[license]: https://github.com/jxskiss/extjson/blob/master/LICENSE

extjson extends the JSON syntax with following features:

1. trailing comma of Object and Array
2. traditional, end of line, and pragma comments
3. simple identifier as object key without quotes
4. Python boolean constants
5. Python None as null
6. Python style single quote string
7. read environment variables
8. include other JSON files (with max depth limited)
9. reference to other values in same file, using [gjson] path syntax
10. evaluate expressions at runtime, with frequently used builtin functions

[gjson]: https://github.com/tidwall/gjson

It helps in many scenes, such as:

1. working with data generated by python programs
2. configuration with comments
3. managing many json files with duplicate contents
4. data driven testing by JSON files

See [example](#example) and [godoc] for more details.

## Example

```text
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
    "test_fn1": @fn("nowUnix"),   // Call builtin function nowUnix.
    test_fn2: @fn("nowFormat('2006-01-02')"), // Call builtin function nowFormat("2006-01-02").
    'test_fn3': @fn("uuid"),      // Call builtin function uuid.
    test_fn4: @fn('randN(10)'),   // Call builtin function randN(10).
    test_fn5: @fn('randStr(16)'), // Call builtin function randStr(16).
}
```

With environment variable "SOME_ENV" set to "some-env-value", the above JSON
will be resolved to following:

```json
{
  "obj1": { "foo": "bar" },
  "array": [ 1, 2, 3 ],
  "include": { "foo": "bar" },
  "identifier_simple1": 1234,
  "$identifierSimple2": "abc",
  "obj2": { "foo": "bar" },
  "py_true": true,
  "py_false": false,
  "py_none": null,
  "test_env": "some-env-value",
  "test_ref1": "bar",
  "test_ref2": 3,
  "test_ref3": 3,
  "friends": [
    { "age": 44, "first": "Dale", "last": "Murphy", "nets": [ "ig", "fb", "tw" ] },
    { "age": 68, "first": "Roger", "last": "Craig", "nets": [ "fb", "tw" ] },
    { "age": 47, "first": "Jane", "last": "Murphy", "nets": [ "ig", "tw" ] }
  ],
  "test_ref4": [ "Dale", "Roger", "Jane" ],
  "test_fn1": 1643162035,
  "test_fn2": "2022-01-26",
  "test_fn3": "4de97237-8ff1-4cc6-80db-d5485ad2b82e",
  "test_fn4": 5,
  "test_fn5": "YewEXAuRrsI3pUCC"
}
```
