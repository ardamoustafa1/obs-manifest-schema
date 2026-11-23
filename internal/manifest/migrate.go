package manifest

import (
	"encoding/json"
	"errors"
	"fmt"

	"go.yaml.in/yaml/v3"
)

func (v *Validator) Migrate(data []byte, from, to string) ([]byte, error) {
	if from != "v1" || to != "v2" {
		return nil, fmt.Errorf("unsupported migration %s -> %s; available migration: v1 -> v2", from, to)
	}
	report, _ := v.Validate(data)
	if !report.Valid() {
		return nil, errors.New("source manifest is invalid; run obsctl validate for details")
	}
	if report.APIVersion != APIVersionV1 {
		return nil, fmt.Errorf("source apiVersion is %q, expected %q", report.APIVersion, APIVersionV1)
	}

	raw, err := decodeSingleYAMLDocument(data)
	if err != nil {
		return nil, err
	}
	metadata := raw["metadata"].(map[string]any)
	spec := raw["spec"].(map[string]any)
	spec["ownership"] = map[string]any{
		"team":             metadata["owner"],
		"slackChannel":     metadata["slack_channel"],
		"pagerDutyService": metadata["pagerduty_service"],
	}
	delete(metadata, "owner")
	delete(metadata, "slack_channel")
	delete(metadata, "pagerduty_service")
	raw["apiVersion"] = APIVersionV2

	jsonData, err := json.Marshal(raw)
	if err != nil {
		return nil, fmt.Errorf("normalize migrated manifest: %w", err)
	}
	var ordered ServiceObservability
	if err := json.Unmarshal(jsonData, &ordered); err != nil {
		return nil, fmt.Errorf("decode migrated manifest: %w", err)
	}
	result, err := yaml.Marshal(&ordered)
	if err != nil {
		return nil, fmt.Errorf("encode migrated manifest: %w", err)
	}
	verification, _ := v.Validate(result)
	if !verification.Valid() {
		return nil, errors.New("internal error: migrated manifest failed v2 validation")
	}
	return result, nil
}
