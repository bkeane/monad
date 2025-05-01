package tmpl

import (
	"bytes"
	"fmt"
	"text/template"
)

type TemplateData struct {
	Git struct {
		Branch string
		Sha    string
		Owner  string
		Repo   string
	}

	Caller struct {
		AccountId string
		Region    string
	}

	Registry struct {
		Id     string
		Region string
	}

	Resource struct {
		Name struct {
			Prefix string
			Full   string
		}
		Path struct {
			Prefix string
			Full   string
		}
	}

	Lambda struct {
		Region      string
		FunctionArn string
		PolicyArn   string
		RoleArn     string
	}

	ApiGateway struct {
		Region string
		Id     string
	}

	Cloudwatch struct {
		Region      string
		LogGroupArn string
	}

	EventBridge struct {
		Region  string
		RuleArn string
		BusName string
	}
}

func Template(name string, tmpl string, data TemplateData) (string, error) {
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
