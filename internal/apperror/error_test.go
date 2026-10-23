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

func TestErrorHelpersUseRegisteredCodes(t *testing.T) {
	for _, appError := range []*Error{NotImplemented("login"), InvalidArgument("bad input"), OperationFailed("scaffold")} {
		if !IsRegisteredCode(appError.Code) {
			t.Errorf("error code %q is not registered", appError.Code)
		}
	}
}
