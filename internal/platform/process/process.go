package process

import (
	"encoding/json"
	"fmt"
	"os"

	"nomyr/internal/buildinfo"
)

func Run(name string) {
	status := map[string]any{
		"binary": name,
		"build":  buildinfo.Current(),
		"status": "not-implemented",
	}
	if encodeError := json.NewEncoder(os.Stdout).Encode(status); encodeError != nil {
		fmt.Fprintln(os.Stderr, encodeError)
		os.Exit(1)
	}
}
