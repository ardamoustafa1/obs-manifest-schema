package schemas

import (
	"encoding/json"
	"testing"
)

func TestReleasedSchemasDeclareIdentity(t *testing.T) {
	t.Parallel()
	tests := []struct {
		path    string
		id      string
		version string
	}{
		{"v1/service-observability.schema.json", "https://schemas.obs.acme.com/service-observability/v1/schema.json", "1.0.0"},
		{"v2/service-observability.schema.json", "https://schemas.obs.acme.com/service-observability/v2/schema.json", "2.0.0"},
	}
	for _, test := range tests {
		t.Run(test.version, func(t *testing.T) {
			data, err := Files.ReadFile(test.path)
			if err != nil {
				t.Fatal(err)
			}
			var schema map[string]any
			if err := json.Unmarshal(data, &schema); err != nil {
				t.Fatalf("invalid JSON: %v", err)
			}
			if schema["$schema"] != "https://json-schema.org/draft/2020-12/schema" {
				t.Fatalf("schema does not declare Draft 2020-12")
			}
			if schema["$id"] != test.id || schema["x-schema-version"] != test.version {
				t.Fatalf("unexpected identity: %#v", schema)
			}
		})
	}
}
