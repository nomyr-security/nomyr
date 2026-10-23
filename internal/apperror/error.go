package apperror

import "fmt"

var registeredCodes = map[string]struct{}{
	"NHI-CORE-0001":   {},
	"NHI-CORE-0002":   {},
	"NHI-CONFIG-0001": {},
	"NHI-CONFIG-0002": {},
}

type Error struct {
	Code          string `json:"code"`
	What          string `json:"what"`
	Cause         string `json:"cause"`
	Fix           string `json:"fix"`
	Docs          string `json:"docs"`
	CorrelationID string `json:"correlation_id,omitempty"`
}

func IsRegisteredCode(code string) bool {
	_, ok := registeredCodes[code]
	return ok
}

func (appError *Error) Error() string {
	message := fmt.Sprintf("%s  %s\n  cause  %s\n  fix    %s\n  docs   %s", appError.Code, appError.What, appError.Cause, appError.Fix, appError.Docs)
	if appError.CorrelationID != "" {
		message += "\n  correlation  " + appError.CorrelationID
	}
	return message
}

func NotImplemented(area string) *Error {
	return &Error{
		Code:  "NHI-CORE-0001",
		What:  area + " is not available in this build",
		Cause: "this capability has not been implemented",
		Fix:   "use the supported local demo commands",
		Docs:  "https://github.com/nomyr-security/nomyr#local-development",
	}
}

func InvalidArgument(reason string) *Error {
	return &Error{
		Code:  "NHI-CONFIG-0002",
		What:  "invalid command argument",
		Cause: reason,
		Fix:   "provide a supported value and retry",
		Docs:  "https://github.com/nomyr-security/nomyr#local-development",
	}
}

func OperationFailed(operation string) *Error {
	return &Error{
		Code:  "NHI-CORE-0002",
		What:  operation + " failed",
		Cause: "the local operation could not be completed",
		Fix:   "check the destination and retry",
		Docs:  "https://github.com/nomyr-security/nomyr#local-development",
	}
}
