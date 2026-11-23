package manifest

import (
	"strings"
	"testing"
)

func TestMigrateV1ToV2(t *testing.T) {
	t.Parallel()
	validator := mustValidator(t)
	migrated, err := validator.Migrate(readFixture(t, "valid", "http.yaml"), "v1", "v2")
	if err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	report, parsed := validator.Validate(migrated)
	if !report.Valid() {
		t.Fatalf("migrated manifest is invalid: %v\n%s", report.Issues, migrated)
	}
	if report.APIVersion != APIVersionV2 {
		t.Fatalf("got apiVersion %q", report.APIVersion)
	}
	if parsed.Metadata.Owner != "" || parsed.Spec.Ownership == nil {
		t.Fatalf("ownership was not moved: %#v", parsed)
	}
	if parsed.Spec.Ownership.Team != "team-checkout" || parsed.Spec.Ownership.SlackChannel != "#checkout-alerts" {
		t.Fatalf("unexpected v2 ownership: %#v", parsed.Spec.Ownership)
	}
	if strings.Contains(string(migrated), "slack_channel:") {
		t.Fatalf("legacy metadata fields remain in output:\n%s", migrated)
	}
}

func TestMigrateRejectsUnsupportedPath(t *testing.T) {
	t.Parallel()
	validator := mustValidator(t)
	_, err := validator.Migrate(readFixture(t, "valid", "http.yaml"), "v2", "v3")
	if err == nil || !strings.Contains(err.Error(), "unsupported migration") {
		t.Fatalf("expected unsupported migration error, got %v", err)
	}
}
