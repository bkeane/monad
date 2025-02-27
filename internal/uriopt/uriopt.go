package uriopt

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
)

// GetContent takes a string that may be a file URI and returns its contents.
// If the input is a file URI (starts with "file://"), it reads and returns the file contents.
// Otherwise, it returns the original string.
func String(input string) (string, error) {
	parsedURL, err := url.Parse(input)
	if err != nil {
		return "", err
	}

	switch parsedURL.Scheme {
	case "file":
		// Strip the file:// prefix to get the file path
		filePath := strings.TrimPrefix(input, "file://")

		// Read the file contents
		content, err := os.ReadFile(filePath)
		if err != nil {
			return "", err
		}

		return string(content), nil
	default:
		// Not a file URI, return original string
		return input, nil
	}
}

func Json(input string) (string, error) {
	content, err := String(input)
	if err != nil {
		return "", err
	}

	// Validate JSON using json.Valid
	if !json.Valid([]byte(content)) {
		return "", fmt.Errorf("invalid JSON")
	}

	return content, nil
}
