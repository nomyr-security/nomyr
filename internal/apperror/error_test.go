package apperror

import (
	"strings"
	"testing"
)

func TestErrorContractIncludesRequiredFields(t *testing.T) {
	message := NotImplemented("login").Error()
	for _, required := range []string{"NHI-CORE-0001", "cause", "fix", "docs"} {
		if !strings.Contains(message, required) {
			t.Fatalf("error contract missing %q: %s", required, message)
		}
	}
}
