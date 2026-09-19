package apperror

import "fmt"

type Error struct {
	Code          string `json:"code"`
	What          string `json:"what"`
	Cause         string `json:"cause"`
	Fix           string `json:"fix"`
	Docs          string `json:"docs"`
	CorrelationID string `json:"correlation_id,omitempty"`
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
