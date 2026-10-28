package middleware

import (
	"strings"
	"testing"
)

func TestResultForLogRedactsPasswords(t *testing.T) {
	result := resultForLog([]byte(`{"code":0,"data":{"password":"secret-value"}}`), operationLogParamsMax)
	if strings.Contains(result, "secret-value") {
		t.Fatalf("response password was recorded: %s", result)
	}
	if !strings.Contains(result, `"password":"***"`) {
		t.Fatalf("response password was not redacted: %s", result)
	}
}

func TestResultForLogOmitsUnparseableResponse(t *testing.T) {
	result := resultForLog([]byte(`password=secret-value`), operationLogParamsMax)
	if strings.Contains(result, "secret-value") {
		t.Fatalf("unparseable response was recorded: %s", result)
	}
}
