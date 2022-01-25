package extjson

import (
	"encoding/json"
	"testing"
)

func TestBuiltinFunctions(t *testing.T) {
	data := `{
		nowUnix: @fn("nowUnix"),
		nowMilli: @fn("nowMilli"),
		nowNano: @fn("nowNano"),
		nowRFC3339: @fn('nowRFC3339'),
		nowFormat_1: @fn("nowFormat('2006-01-02')"),
		nowFormat_2: @fn('nowFormat("2006-01-02")'),
		uuid: @fn("uuid"),
		rand: @fn("rand"),
		randN: @fn("randN(1000)"),
		randStr: @fn("randStr(32)"),
	}`
	got := make(map[string]interface{})
	err := Unmarshal([]byte(data), &got)
	if err != nil {
		t.Fatalf("failed umarshal expressions, err= %v", err)
	}
	tmp, _ := json.Marshal(got)
	t.Log(string(tmp))
}

func TestUserFunctions(t *testing.T) {
	funcs := FuncMap{
		"myfunc1": func() string {
			return "myfunc1"
		},
		"myfunc2": func(arg int) int {
			return arg * 2
		},
		"myfunc3": func(arg int) (int, error) {
			return arg * 3, nil
		},
	}
	data := `{
		nowUnix: @fn("nowUnix"),
		myfunc1: @fn("myfunc1"),
		myfunc2: @fn("myfunc2(5)"),
		myfunc3: @fn("myfunc3(5)"),
	}`
	got := make(map[string]interface{})
	err := Unmarshal([]byte(data), &got, WithFuncMap(funcs))
	if err != nil {
		t.Fatalf("failed unmarshal with user functions, err= %v", err)
	}
	if x := got["myfunc1"]; x != "myfunc1" {
		t.Fatalf("got unexpected result for myfunc1: %q", x)
	}
	if x := got["myfunc2"]; x != float64(10) {
		t.Fatalf("got unexpected result for myfunc2: %q", x)
	}
	if x := got["myfunc3"]; x != float64(15) {
		t.Fatalf("got unexpected result for myfunc3: %q", x)
	}

	tmp, _ := json.Marshal(got)
	t.Log(string(tmp))
}
