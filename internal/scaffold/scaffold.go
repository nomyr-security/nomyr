package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"nomyr/internal/apperror"
)

var validName = regexp.MustCompile(`^[a-z][a-z0-9-]{0,62}$`)

func Create(directory, kind, name string) (string, error) {
	if kind != "connector" && kind != "rule" {
		return "", apperror.InvalidArgument("scaffold kind must be connector or rule")
	}
	if !validName.MatchString(name) {
		return "", apperror.InvalidArgument("name must start with a lowercase letter and contain only lowercase letters, digits, or hyphens")
	}
	content, err := template(kind, name)
	if err != nil {
		return "", err
	}
	target := filepath.Join(directory, name+".yaml")
	file, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		if os.IsExist(err) {
			return "", apperror.InvalidArgument("target file already exists: " + target)
		}
		return "", apperror.OperationFailed("create scaffold file")
	}
	if _, err = file.WriteString(content); err != nil {
		_ = file.Close()
		_ = os.Remove(target)
		return "", apperror.OperationFailed("write scaffold file")
	}
	if err = file.Close(); err != nil {
		_ = os.Remove(target)
		return "", apperror.OperationFailed("close scaffold file")
	}
	return target, nil
}

func template(kind, name string) (string, error) {
	var builder strings.Builder
	switch kind {
	case "connector":
		fmt.Fprintf(&builder, "# scaffold-only: no provider support or capabilities are implemented\napiVersion: nhi.connector/v1\nkind: CapabilityManifest\nmetadata:\n  name: %s\n  version: 0.1.0\ncapabilities: []\n", name)
	case "rule":
		fmt.Fprintf(&builder, "kind: RuleScaffold\nname: %s\nstatus: scaffold-only\ninputs: []\nexpression: null\n", name)
	default:
		return "", apperror.InvalidArgument("scaffold kind must be connector or rule")
	}
	return builder.String(), nil
}
