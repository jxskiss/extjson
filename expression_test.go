package extjson

import (
	"encoding/json"
	"testing"
)

func TestExpression(t *testing.T) {
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
