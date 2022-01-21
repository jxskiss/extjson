package extjson

import (
	"os"
	"strings"
	"testing"
)

func TestDisableEnv(t *testing.T) {
	jdata := `{
"test_env": @env("SOME_ENV"),
}`
	os.Setenv("SOME_ENV", "some-env-value")
	got := make(map[string]interface{})
	err := Unmarshal([]byte(jdata), &got)
	if !strings.Contains(err.Error(), "env feature is not enabled") {
		t.Fatalf("failed unmarshal extended json: %v", err)
	}
}
