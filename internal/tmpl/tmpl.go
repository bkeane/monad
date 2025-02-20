package tmpl

import (
	"bytes"
	"fmt"
	"text/template"
)

func Template(name string, tmpl string, data interface{}) (string, error) {
	t, err := template.New("").Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("failed to parse %s template: %w", name, err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute %s template: %w", name, err)
	}

	return buf.String(), nil
}
