package commands

import (
	"fmt"
	"io"
	"os"
)

// readContent returns message text from either an inline flag value or a file
// ("-" reads stdin). Exactly one of the two must be provided; flagName and
// fileFlagName appear in error messages.
func readContent(text, file, flagName, fileFlagName string) (string, error) {
	switch {
	case text != "" && file != "":
		return "", fmt.Errorf("use %s or %s, not both", flagName, fileFlagName)
	case text != "":
		return text, nil
	case file == "":
		return "", fmt.Errorf("provide %s or %s", flagName, fileFlagName)
	case file == "-":
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("reading stdin: %w", err)
		}
		return string(data), nil
	default:
		data, err := os.ReadFile(file)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}
}
