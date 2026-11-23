package manifest

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/acme/obs-manifest-schema/schemas"
	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"
	"go.yaml.in/yaml/v3"
)

const (
	schemaURLV1 = "https://schemas.obs.acme.com/service-observability/v1/schema.json"
	schemaURLV2 = "https://schemas.obs.acme.com/service-observability/v2/schema.json"
)

type Validator struct {
	schemas map[string]*jsonschema.Schema
}

func NewValidator() (*Validator, error) {
	compiler := jsonschema.NewCompiler()
	compiler.DefaultDraft(jsonschema.Draft2020)

	resources := []struct {
		version string
		url     string
		path    string
	}{
		{APIVersionV1, schemaURLV1, "v1/service-observability.schema.json"},
		{APIVersionV2, schemaURLV2, "v2/service-observability.schema.json"},
	}

	for _, resource := range resources {
		data, err := schemas.Files.ReadFile(resource.path)
		if err != nil {
			return nil, fmt.Errorf("read embedded schema %s: %w", resource.version, err)
		}
		var document any
		if err := json.Unmarshal(data, &document); err != nil {
			return nil, fmt.Errorf("decode embedded schema %s: %w", resource.version, err)
		}
		if err := compiler.AddResource(resource.url, document); err != nil {
			return nil, fmt.Errorf("register schema %s: %w", resource.version, err)
		}
	}

	compiled := make(map[string]*jsonschema.Schema, len(resources))
	for _, resource := range resources {
		schema, err := compiler.Compile(resource.url)
		if err != nil {
			return nil, fmt.Errorf("compile schema %s: %w", resource.version, err)
		}
		compiled[resource.version] = schema
	}
	return &Validator{schemas: compiled}, nil
}

func (v *Validator) Validate(data []byte) (Report, *ServiceObservability) {
	report := Report{}
	raw, err := decodeSingleYAMLDocument(data)
	if err != nil {
		report.Issues = append(report.Issues, Issue{Severity: SeverityError, Message: err.Error()})
		return report, nil
	}

	apiVersion, _ := raw["apiVersion"].(string)
	report.APIVersion = apiVersion
	schema, ok := v.schemas[apiVersion]
	if !ok {
		report.Issues = append(report.Issues, Issue{
			Severity: SeverityError,
			Path:     "/apiVersion",
			Message:  fmt.Sprintf("unsupported apiVersion %q; supported versions are %s and %s", apiVersion, APIVersionV1, APIVersionV2),
		})
		return report, nil
	}

	if err := schema.Validate(raw); err != nil {
		report.Issues = append(report.Issues, Issue{Severity: SeverityError, Message: err.Error()})
		return report, nil
	}

	jsonData, err := json.Marshal(raw)
	if err != nil {
		report.Issues = append(report.Issues, Issue{Severity: SeverityError, Message: fmt.Sprintf("encode normalized manifest: %v", err)})
		return report, nil
	}
	var manifest ServiceObservability
	if err := json.Unmarshal(jsonData, &manifest); err != nil {
		report.Issues = append(report.Issues, Issue{Severity: SeverityError, Message: fmt.Sprintf("decode typed manifest: %v", err)})
		return report, nil
	}

	report.Issues = append(report.Issues, semanticIssues(&manifest)...)
	if apiVersion == APIVersionV1 {
		report.Issues = append(report.Issues, Issue{
			Severity: SeverityWarning,
			Path:     "/apiVersion",
			Message:  "obs.acme.com/v1 is deprecated but remains supported; migrate with: obsctl migrate v1 v2 observability.yaml",
		})
	}
	return report, &manifest
}

func decodeSingleYAMLDocument(data []byte) (map[string]any, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	var raw map[string]any
	if err := decoder.Decode(&raw); err != nil {
		return nil, fmt.Errorf("parse YAML: %w", err)
	}
	if raw == nil {
		return nil, errors.New("manifest is empty")
	}
	var extra any
	err := decoder.Decode(&extra)
	if err == nil {
		return nil, errors.New("exactly one YAML document is allowed")
	}
	if !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("parse trailing YAML document: %w", err)
	}
	return raw, nil
}
