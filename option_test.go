package extjson

import (
	"os"
	"testing"
)

func TestDisableEnv(t *testing.T) {
	jdata := `{
"test_env": @env("SOME_ENV"),
}`
	os.Setenv("SOME_ENV", "some-env-value")
	got := make(map[string]interface{})
	err := Unmarshal([]byte(jdata), &got)
	if err != nil {
		t.Fatalf("failed unmarshal extended json: %v", err)
	}
	if got["test_env"] != "" {
		t.Fatalf("got unexpected env value: %q", got["test_env"])
	}
}
